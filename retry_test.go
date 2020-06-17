package hooks

import (
	"testing"
	"time"
)

func TestIncrDelay(t *testing.T) {
	testData := []struct {
		delay     time.Duration
		factorPct int64
		addDelay  time.Duration
	}{
		{time.Second, 100, time.Second},
		{10 * time.Second, 50, 5 * time.Second},
		{10 * time.Second, 10, time.Second},
		{10 * time.Second, 0, 0},
	}

	for _, td := range testData {
		if delay := incrDelay(td.delay, td.factorPct); delay != td.addDelay {
			t.Errorf("delay=%s, factor=%d, adjust delay is wrong: actual=%s, expected=%s",
				td.delay, td.factorPct, delay, td.addDelay)
		}
	}
}

func TestMakeJitter(t *testing.T) {
	testData := []struct {
		delay     time.Duration
		jitterPct int64
		addDelay  time.Duration
	}{
		{time.Second, 100, time.Second},
		{10 * time.Second, 50, 5 * time.Second},
		{10 * time.Second, 10, time.Second},
		{10 * time.Second, 0, 0},
	}

	for _, td := range testData {
		if delay := makeJitter(td.delay, td.jitterPct); delay != td.addDelay {
			t.Errorf("delay=%s, jitter=%d, adjust delay is wrong: actual=%s, expected=%s",
				td.delay, td.jitterPct, delay, td.addDelay)
		}
	}
}

func TestRetryFailure(t *testing.T) {

	nTests := 8

	for i := 0; i < nTests; i++ {
		hook := RetryHook(
			&mockRetryHook{maxFailures: i + 1},
			time.Microsecond,
			Retries(i),
		)

		if err := hook.Fire(nil); err == nil {
			t.Errorf("success with %d retries", i)
		}
	}
}

func TestRetrySuccess(t *testing.T) {

	nTests := 8

	for i := 0; i < nTests; i++ {
		hook := RetryHook(
			&mockRetryHook{maxFailures: i},
			time.Microsecond,
			Retries(i),
		)

		if err := hook.Fire(nil); err != nil {
			t.Errorf("failed with %d retries", i)
		}
	}
}
