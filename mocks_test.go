package hooks

import (
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// mockCannedHook is a simple hook that always returns the same canned results
type mockCannedHook struct {
	levels     []logrus.Level
	fireResult error
}

func (mock *mockCannedHook) Levels() []logrus.Level {
	return mock.levels
}

func (mock *mockCannedHook) Fire(entry *logrus.Entry) error {
	return mock.fireResult
}

// mockRecordingHook is hook that keeps all messages it has received
type mockRecordingHook struct {
	ChainImpl
	messages sync.Map
}

// Fire stores a copy of the message
func (mock *mockRecordingHook) Fire(entry *logrus.Entry) error {
	entry2 := *entry
	mock.messages.Store(entry2.Message, &entry2)

	return nil
}

// compare reports discrepancies between messages in maps passed as arguments
func (mock *mockRecordingHook) len(*testing.T) int {
	nReceived := 0
	mock.messages.Range(func(key, _ interface{}) bool {
		nReceived++
		return true
	})

	return nReceived
}

// resets empties the buffer of received messages
func (mock *mockRecordingHook) reset(*testing.T) {
	mock.messages = sync.Map{}
}

// compare reports discrepancies between messages in maps passed as arguments
func (mock *mockRecordingHook) compare(t *testing.T, sent []*logrus.Entry) {
	sentMap := make(map[string]*logrus.Entry, len(sent))
	for _, msg := range sent {
		sentMap[msg.Message] = msg
		if _, found := mock.messages.Load(msg.Message); !found {
			t.Errorf("this message was not received: %s", msg.Message)
			continue
		}
	}

	nReceived := 0
	mock.messages.Range(func(key, _ interface{}) bool {
		nReceived++

		msg, ok := key.(string)
		if !ok {
			// this should never happen
			t.Fatalf("invalid type of message key, expected string: %s", key)
		}

		if _, found := sentMap[msg]; !found {
			t.Errorf("this message was not sent: %s", msg)
		}

		return true
	})

	if len(sent) != nReceived {
		t.Errorf("number of sent message [%d] != [%d] number of received message",
			len(sent), nReceived)
	}
}
