// Package hooks provides several useful Logrus hooks
//
// These hooks are used as decorators of other hooks and provide enhanced
// functionality:
//
//	- retry transmission with exponential backoff and jitter
//	- rate limits on the number of logging messages
//	- asynchronous execution
//
package hooks

import "github.com/sirupsen/logrus"

// Chain is a single-linked list of hooks that work together one after another
type Chain interface {
	logrus.Hook

	// Next is the element of the chain that follows this one
	Next() logrus.Hook
}

// RunningHook is a Logrus hook that can be started and stopped
type RunningHook interface {
	logrus.Hook

	// Start prepares the hook to be able to send alerts
	Start() error

	// Stop transitions the hook to a state in which it does not send alerts
	Stop() error
}
