package hooks

// _RetryHook_ repeats the `Fire` call several times with pauses between
// attempts until either the logging message is processed successfully or
// the maximum number of retries is reached.
//
// Credit:
//
// Code is based on this implementation from the `gorealis` project:
//
// https://github.com/paypal/gorealis/blob/master/retry.go

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// default backoff values
	backoffFactorPct = 100
	backoffJitterPct = 10
	maxRetries       = 3
)

// Backoff determines how the retry mechanism should react after each
// failure and how many failures it should tolerate
type backoff struct {

	// retryDelay is the base delay between retries
	retryDelay time.Duration

	// factorPct is the increase (in percents of the delay) to the delay after each retry
	factorPct int64

	// jitterPct is random delay (in percents of the delay) to add to each retry
	jitterPct int64

	// maxRetries is the maximum number of retries
	maxRetries int
}

// retryHook is a Logrus hook inserted in a chain of hooks to add retry capability
type retryHook struct {
	ChainImpl
	backoff
}

// constructor --------------------------------------------------------

type RetryOption func(conf *backoff)

func FactorPct(n int64) RetryOption {
	return func(conf *backoff) {
		conf.factorPct = n
	}
}

func JitterPct(n int64) RetryOption {
	return func(conf *backoff) {
		conf.jitterPct = n
	}
}

func Retries(n int) RetryOption {
	return func(conf *backoff) {
		if n >= 0 {
			conf.maxRetries = n
		}
	}
}

func RetryHook(next logrus.Hook, delay time.Duration, opts ...RetryOption) logrus.Hook {

	hook := &retryHook{
		ChainImpl: ChainImpl{
			ChainElement{
				next: next,
			},
		},
		// default backoff
		backoff: backoff{
			retryDelay: delay,
			factorPct:  backoffFactorPct,
			jitterPct:  backoffJitterPct,
			maxRetries: maxRetries,
		},
	}

	for _, opt := range opts {
		opt(&hook.backoff)
	}

	return hook
}

// implementation -----------------------------------------------------

// Fire makes multiple attempts to deliver the message to the next hook
func (h *retryHook) Fire(entry *logrus.Entry) error {

	delay := h.retryDelay

	var err error
	for retries := 0; retries <= h.maxRetries; retries++ {
		if err = h.next.Fire(entry); err == nil {
			// message logged successfully
			return nil
		}
		if retries == h.maxRetries {
			// maximum number of retries reached
			return err
		}

		adjustedDelay := delay
		if retries > 0 {
			adjustedDelay += makeJitter(delay, h.jitterPct)
			delay += incrDelay(delay, h.factorPct)
		}

		// pause between reties
		time.Sleep(adjustedDelay)
	}

	// all retries failed
	return fmt.Errorf("failed after [%d] retries: %w", h.maxRetries, err)
}

func incrDelay(delay time.Duration, factorPct int64) time.Duration {
	return time.Duration(delay.Nanoseconds() * factorPct / 100)
}

func makeJitter(delay time.Duration, jitterPct int64) time.Duration {
	return time.Duration(delay.Nanoseconds() * jitterPct / 100)
}
