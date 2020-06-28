package hooks

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
)

// mockRelayHook is a simple hook that relays everything to the next hook
type mockRelayHook struct {
	ChainImpl
}

func (mock *mockRelayHook) Fire(entry *logrus.Entry) error {
	return mock.Next().Fire(entry)
}

func TestChain_Fire(t *testing.T) {
	testData := []error{
		nil,
		ErrBufferFull,
		ErrNotRunning,
		fmt.Errorf("fancy error"),
	}

	cannedHook := mockCannedHook{}
	hook := mockRelayHook{
		ChainImpl{
			ChainElement{
				next: &cannedHook,
			},
		},
	}
	for i, td := range testData {

		// set desired result of Fire() call
		cannedHook.fireResult = td

		if fireResult := hook.Fire(nil); fireResult != td {
			t.Errorf("wrong result from Fire call at [test=%d]: expected=%s, received=%s",
				i, td, fireResult)
		}
	}
}

func TestChain_Levels(t *testing.T) {
	testData := []struct {
		levels []logrus.Level
	}{
		{[]logrus.Level{}},
		{[]logrus.Level{logrus.InfoLevel}},
		{[]logrus.Level{logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel}},
		{[]logrus.Level{logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel}},
	}

	for i, td := range testData {
		hook := mockRelayHook{
			ChainImpl{
				ChainElement{
					next: &mockCannedHook{
						levels: td.levels,
					},
				},
			},
		}

		levels := hook.Levels()
		if len(td.levels) != len(levels) {
			t.Errorf("number of supported log levels is wrong at [test=%d]: expected=%d, found=%d",
				i, len(td.levels), len(levels))
		}

		for j, level := range td.levels {
			// avoid out-of-range errors
			if j >= len(levels) {
				break
			}

			if level != levels[j] {
				t.Errorf("unexpected level at [test=%d, pod=%d: expected=%s, found=%s",
					i, j, level, levels[j])
			}
		}
	}
}
