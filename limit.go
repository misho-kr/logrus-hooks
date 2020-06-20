package hooks

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	// default rate limits, completely arbitrary
	defaultRatePerSecond = 10
	defaultBurst         = 3
)

// retryHook is a Logrus hook that enforces a rate limit on the logged messages
type rareLimitHook struct {
	ChainImpl
	limiter *rate.Limiter
}

type rateLimit struct {
	limitPeSecond int
	burst         int
}

// constructor --------------------------------------------------------

// RateLimitOption is a functional option to update the rate limit hook configuration
type RateLimitOption func(conf *rateLimit)

// PerSecond sets the maximum number of messages that can be logged per second
func PerSecond(n int) RateLimitOption {
	return func(conf *rateLimit) {
		conf.limitPeSecond = n
	}
}

// Burst sets the maximum number of messages that can be logged per second in a burst
func Burst(n int) RateLimitOption {
	return func(conf *rateLimit) {
		conf.burst = n
	}
}

// RateLimitHook creates a Logrus hook that enforces a rate limit on the logged messages
func RateLimitHook(next logrus.Hook, opts ...RateLimitOption) logrus.Hook {

	// default configuration
	conf := rateLimit{
		limitPeSecond: defaultRatePerSecond,
		burst:         defaultBurst,
	}
	for _, opt := range opts {
		opt(&conf)
	}

	hook := &rareLimitHook{
		ChainImpl: ChainImpl{
			ChainElement{
				next: next,
			},
		},
		limiter: rate.NewLimiter(
			rate.Limit(conf.limitPeSecond),
			conf.burst,
		),
	}

	return hook
}

// implementation -----------------------------------------------------

// Fire makes multiple attempts to deliver the message to the next hook
func (h *rareLimitHook) Fire(entry *logrus.Entry) error {

	if !h.limiter.Allow() {
		return fmt.Errorf("rate limit [%f/sec, burst=%d] exceeded",
			h.limiter.Limit(), h.limiter.Burst())
	}

	return h.next.Fire(entry)
}
