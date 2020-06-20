package hooks

import "github.com/sirupsen/logrus"

// ChainElement is a partial implementation of Chain interface that provides delegation
//
// This type has `Next` method to access the next hook in the chain of
// hooks. It does not have `Fire` and `Levels` methods, they have to be
// implemented by the custom hooks.
type ChainElement struct {
	next logrus.Hook
}

// Next is an accessor to the next element of the Chain
func (el *ChainElement) Next() logrus.Hook {
	return el.next
}

// ChainImpl is a partial implementation of Chain interface that has Level method
//
// This type provides `Levels` method which simply delegates the call to
// the next hook in the chain. This is the expected behavior of hooks that
// implement some custom action but do not care about the log level and let
// someone decide whether hook should fire or not.
type ChainImpl struct {
	ChainElement
}

// Levels delegates the query for supported logging levels to the next chain element
func (impl *ChainImpl) Levels() []logrus.Level {
	return impl.Next().Levels()
}
