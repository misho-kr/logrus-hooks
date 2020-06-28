package hooks

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimit(t *testing.T) {
	testData := []rate.Limit{
		1, 10, 25, 50, 100,
	}

	for _, td := range testData {
		ratePerSecond := int(td)
		hook := RateLimitHook(
			&mockCannedHook{},
			PerSecond(ratePerSecond),
		)

		t.Run(fmt.Sprintf("%d-per-sec", ratePerSecond), func(t *testing.T) {
			for i := 0; i < ratePerSecond; i++ {
				if err := hook.Fire(nil); err != nil {
					t.Fatalf("rate limited too early after %d times: %s", i, err)
				}
				if i < (ratePerSecond - 1) {
					time.Sleep(time.Duration(time.Second.Nanoseconds() / int64(ratePerSecond)))
				}
			}

			if hook.Fire(nil) == nil {
				t.Fatalf("hook was not limited after %d times", ratePerSecond)
			}
		})
	}
}

func TestBurst(t *testing.T) {
	testData := []int{
		1, 10, 25, 50, 100,
	}

	for _, td := range testData {
		hook := RateLimitHook(
			&mockCannedHook{},
			PerSecond(1),
			Burst(td),
		)

		t.Run(fmt.Sprintf("burst of %d", td), func(t *testing.T) {
			for i := 0; i < td; i++ {
				if err := hook.Fire(nil); err != nil {
					t.Fatalf("rate limited too early after %d times: %s", i, err)
				}
			}

			if hook.Fire(nil) == nil {
				t.Fatalf("hook was not limited after %d times", td)
			}
		})
	}
}
