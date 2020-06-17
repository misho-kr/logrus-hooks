package hooks

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

const (
	// asyncSenders is the number of goroutines working to send messages
	asyncSenders = 4

	// asyncSenders is the number of extra goroutines working to send messages
	asyncBoostSenders = 64

	// asyncBuffers is the number of messages that can be stored for sending
	asyncBuffers = 32
)

var (
	errBufferFull error
	errNotRunning error
)

func init() {
	errBufferFull = errors.New("logrus hook failed to send message, buffer is full")
	errNotRunning = errors.New("logrus hook can not send message before it is started")
}

// asyncHook is a Logrus hook uses goroutines to invoke the next hook
type asyncHook struct {
	sync.Mutex

	ChainImpl

	conf      asyncParams
	running   bool
	errLogger *log.Logger

	// messages is a buffer for log entries that will be sent by goroutines
	messages chan *logrus.Entry

	// sendersTracker keeps track of the goroutines that read
	// from the buffer and send out the queued messages
	sendersTracker sync.WaitGroup

	// boostSendersTracker keeps track of the extra goroutines that read
	// from the buffer and send out the queued messages
	boostSendersTracker sync.WaitGroup

	// nBoostSenders is the number of currently running extra goroutines to
	// send out the queued messages
	nBoostSenders uint32
}

// asyncParams defines the performance options of the hook
type asyncParams struct {
	numSenders      uint32
	numBoostSenders uint32
	bufferLen       uint32
}

// constructor --------------------------------------------------------

type AsyncOption func(conf *asyncParams)

// Senders sets the number of senders goroutines of the new hook
func Senders(n uint32) AsyncOption {
	return func(conf *asyncParams) {
		conf.numSenders = n
	}
}

// BoostSenders sets the number of boost senders goroutines of the new hook
func BoostSenders(n uint32) AsyncOption {
	return func(conf *asyncParams) {
		conf.numBoostSenders = n
	}
}

// BufferLen sets the maximum number of messages that can be queued for transmission
func BufferLen(n uint32) AsyncOption {
	return func(conf *asyncParams) {
		conf.bufferLen = n
	}
}

func AsyncHook(next logrus.Hook, opts ...AsyncOption) logrus.Hook {

	hook := &asyncHook{
		ChainImpl: ChainImpl{
			ChainElement{
				next: next,
			},
		},
		// default configuration
		conf: asyncParams{
			numSenders:      asyncSenders,
			numBoostSenders: asyncBoostSenders,
			bufferLen:       asyncBuffers,
		},
	}

	for _, opt := range opts {
		opt(&hook.conf)
	}

	return hook
}

// implementation -----------------------------------------------------

// Fire makes multiple attempts to deliver the message to the next hook
func (h *asyncHook) Fire(entry *logrus.Entry) error {

	// the hook must be in running state
	h.Lock()
	defer h.Unlock()

	if !h.isRunning() {
		return errNotRunning
	}

	select {
	case h.messages <- entry:
		// message was passed to the senders, no error
	default:
		// buffer is full because senders are too busy or too slow
		// try to boost the senders if possible
		return h.boostAndWork(entry)
	}

	return nil
}

// IsRunning queries the state of the hook that safe for concurrent access
func (h *asyncHook) IsRunning() bool {
	h.Lock()
	defer h.Unlock()

	return h.isRunning()
}

// IsRunning queries the state of the hook that is NOT safe for concurrent access
func (h *asyncHook) isRunning() bool {
	return h.running
}

// Start prepares the hook to send messages via goroutines
func (h *asyncHook) Start() error {
	h.Lock()
	defer h.Unlock()

	h.messages = make(chan *logrus.Entry, h.conf.bufferLen)
	h.sendersTracker.Add(int(h.conf.numSenders))
	for i := 0; i < int(h.conf.numSenders); i++ {
		go h.worker()
	}

	h.running = true

	return nil
}

// Stop transitions the hook to a state in which it does not send messages
func (h *asyncHook) Stop() error {
	h.Lock()
	defer h.Unlock()

	// wait for all booster senders to complete and exit
	h.boostSendersTracker.Wait()

	// stop accepting more messages
	close(h.messages)

	// wait for all senders to complete and exit
	h.sendersTracker.Wait()

	h.running = false

	return nil
}

// worker runs in a loop to send out messages that were queued in the buffer
func (h *asyncHook) worker() {
	defer h.sendersTracker.Done()

	for entry := range h.messages {
		if err := h.next.Fire(entry); err != nil {
			h.errLogger.Print(err)
		}
	}
}

// boostAndWork starts an extra goroutine to help empty out the message buffer
// note: this function must be called with the hook's mutex locked
func (h *asyncHook) boostAndWork(entry *logrus.Entry) error {
	nBoostSenders := atomic.LoadUint32(&h.nBoostSenders)
	if nBoostSenders >= h.conf.numBoostSenders {
		return errBufferFull
	}

	atomic.AddUint32(&h.nBoostSenders, 1)
	h.boostSendersTracker.Add(1)

	go h.booster(entry)

	return nil
}

// booster helps to empty the message buffer while trying to queue up the new message
func (h *asyncHook) booster(entry *logrus.Entry) {
	defer func() {
		// decrement the number of booster workers by 1
		atomic.AddUint32(&h.nBoostSenders, ^uint32(0))
		h.boostSendersTracker.Done()
	}()

	var (
		haveMessage bool
		msg2        *logrus.Entry
	)

	for {
		haveMessage = false

		// pick up one message from the buffer and hopefully that
		// will make room for message that was passed to this function
		select {
		case msg2 = <-h.messages:
			// the buffer was not empty
			haveMessage = true
		default:
			// when the buffer is empty and the message had been queued up
			// this booster worker is not needed
			if entry == nil {
				return
			}
		}

		// quickly try to queue up the message that was passed to this
		// function, if not done so already
		if entry != nil {
			select {
			case h.messages <- entry:
				// finally the message is in the queue
				entry = nil
			default:
				// buffer is still full, will keep trying
			}
		}

		if haveMessage {
			// send out the message that was picked up from the buffer
			if err := h.next.Fire(msg2); err != nil {
				h.errLogger.Printf("booster worker of async logrus hook: %s", err)
			}
		}
	}
}
