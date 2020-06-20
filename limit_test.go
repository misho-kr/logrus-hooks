package hooks

import (
	"fmt"
	"golang.org/x/time/rate"
	"testing"
)

func TestRateLimit(t *testing.T) {
	testData := []rate.Limit{
		1, 10, 50, 100,
	}

	for _, td := range testData {
		ratePerSecond := int(td)
		hook := RateLimitHook(
			&mockRetryHook{},
			PerSecond(ratePerSecond),
		)

		for i := 0; i < ratePerSecond; i++ {
			if hook.Fire(nil) != nil {
				fmt.Errorf("hook was limited [rate=%d/sec] too early: %d", ratePerSecond, i)
			}
		}

		if hook.Fire(nil) == nil {
			fmt.Errorf("hook was not limited: rate=%d/sec", ratePerSecond)
		}
	}
}

func TestBurst(t *testing.T) {
	testData := []int{
		1, 10, 50, 100,
	}

	for _, td := range testData {
		hook := RateLimitHook(
			&mockRetryHook{},
			PerSecond(1),
			Burst(td),
		)

		for i := 0; i < td; i++ {
			if hook.Fire(nil) != nil {
				fmt.Errorf("hook was limited [burst=%d/sec] too early: %d", td, i)
			}
		}

		if hook.Fire(nil) == nil {
			fmt.Errorf("hook was not limited: burst=%d/sec", td)
		}
	}
}
