package hooks

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAsyncHook_Send(t *testing.T) {

	nTests := 64

	var mockHook mockRecordingHook
	hook := AsyncHook(&mockHook, Senders(uint32(nTests)))

	theHook, ok := hook.(*asyncHook)
	if !ok {
		t.Fatalf("test hook is not of the expected asyncHook type: %v", hook)
	}
	if err := theHook.Start(); err != nil {
		t.Fatalf("failed to start the async hook: %s", err)
	}

	sentMessages := make([]*logrus.Entry, 0, nTests)
	for i := 0; i < nTests; i++ {

		testMessage := logrus.NewEntry(logrus.StandardLogger())
		testMessage.Message = fmt.Sprintf("test message: %d", i)

		select {
		case theHook.messages <- testMessage:
			// message sent
			sentMessages = append(sentMessages, testMessage)
		default:
			// buffer is full, this unittest does not involve boosters
		}
	}

	if err := theHook.Stop(); err != nil {
		t.Fatalf("failed to stop the async hook: %s", err)
	}

	if mockHook.len(t) == 0 {
		t.Errorf("no messages were sent, expected at least (%d out of %d)",
			theHook.conf.numSenders, nTests)
	}

	mockHook.compare(t, sentMessages)
}

func TestAsync_boostAndWork_full(t *testing.T) {

	nTests := 64

	var mockHook mockRecordingHook
	hook := AsyncHook(&mockHook, BoostSenders(0))

	theHook, ok := hook.(*asyncHook)
	if !ok {
		t.Fatalf("test hook is not of the expected asyncHook type: %v", hook)
	}
	if err := theHook.Start(); err != nil {
		t.Fatalf("failed to start the async hook: %s", err)
	}

	sentMessages := make([]*logrus.Entry, 0, nTests)
	for i := 0; i < nTests; i++ {

		testMessage := logrus.NewEntry(logrus.StandardLogger())
		testMessage.Message = fmt.Sprintf("test message: %d", i)

		err := theHook.boostAndWork(testMessage)
		if err == nil {
			t.Errorf("boost-and-work did not fail at round: %d", i)
		} else if err != errBufferFull {
			t.Errorf("unexpected error from boost-and-work at round [%d]: %s", i, err)
		}
	}

	if err := theHook.Stop(); err != nil {
		t.Fatalf("failed to stop the async hook: %s", err)
	}

	mockHook.compare(t, sentMessages)
}

func TestAsync_boostAndWork(t *testing.T) {

	nTests := 64

	var mockHook mockRecordingHook
	hook := AsyncHook(&mockHook, Senders(0))

	theHook, ok := hook.(*asyncHook)
	if !ok {
		t.Fatalf("test hook is not of the expected asyncHook type: %v", hook)
	}

	sentMessages := make([]*logrus.Entry, 0, nTests)
	for i := 1; i < nTests; i++ {

		// set the number of booster workers
		theHook.conf.numBoostSenders = uint32(i)

		if err := theHook.Start(); err != nil {
			t.Fatalf("failed to start the async hook at round [%d]: %s", i, err)
		}

		// truncate the sent messages
		sentMessages = sentMessages[:0]
		mockHook.reset(t)

		for j := 0; j < i; j++ {

			testMessage := logrus.NewEntry(logrus.StandardLogger())
			testMessage.Message = fmt.Sprintf("test message: %d", j)

			err := theHook.boostAndWork(testMessage)
			if err != nil {
				t.Errorf("boost-and-work failed at round [%d/%d]: %s", j, i, err)
			} else {
				sentMessages = append(sentMessages, testMessage)
			}
		}

		if err := theHook.Stop(); err != nil {
			t.Fatalf("failed to stop the async hook at round [%d]: %s", i, err)
		}

		mockHook.compare(t, sentMessages)
	}
}

func TestAsync_Fire(t *testing.T) {

	nTests := 1024

	var mockHook mockRecordingHook

	nSenders := uint32(1 + nTests/10)
	hook := AsyncHook(&mockHook, Senders(nSenders), BoostSenders(uint32(nTests)-nSenders))

	theHook, ok := hook.(*asyncHook)
	if !ok {
		t.Fatalf("test hook is not of the expected asyncHook type: %v", hook)
	}

	if err := theHook.Start(); err != nil {
		t.Fatalf("failed to start the async hook: %s", err)
	}

	sentMessages := make([]*logrus.Entry, 0, nTests)
	for i := 0; i < nTests; i++ {

		testMessage := logrus.NewEntry(logrus.StandardLogger())
		testMessage.Message = fmt.Sprintf("test message: %d", i)

		err := hook.Fire(testMessage)
		if err != nil {
			t.Errorf("boost-and-work failed at round [%d]: %s", i, err)
		} else {
			sentMessages = append(sentMessages, testMessage)
		}
	}

	if err := theHook.Stop(); err != nil {
		t.Fatalf("failed to stop the async hook: %s", err)
	}

	mockHook.compare(t, sentMessages)
}
