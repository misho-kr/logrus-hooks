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

// retryHook is a Logrus hook inserted in a chain of hooks to limit the rate of logging
type rareLimitHook struct {
	ChainImpl
	limiter *rate.Limiter
}

type rateLimit struct {
	limitPeSecond int
	burst         int
}

// constructor --------------------------------------------------------

type RateLimitOption func(conf *rateLimit)

func PerSecond(n int) RateLimitOption {
	return func(conf *rateLimit) {
		conf.limitPeSecond = n
	}
}

func Burst(n int) RateLimitOption {
	return func(conf *rateLimit) {
		conf.burst = n
	}
}

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

	if ! h.limiter.Allow() {
		return fmt.Errorf("rate limit [%f/sec, burst=%d] exceeded",
			h.limiter.Limit(), h.limiter.Burst())
	}

	return h.next.Fire(entry)
}
