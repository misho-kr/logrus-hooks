package hooks

import "github.com/sirupsen/logrus"

// ChainElement is a partial implementation of Chain interface that provides delegation
type ChainElement struct {
	next logrus.Hook
}

// Next is an accessor to the next element of the Chain
func (el *ChainElement) Next() logrus.Hook {
	return el.next
}

// ChainElement is a partial implementation of Chain interface that is used by other hooks
type ChainImpl struct {
	ChainElement
}

// Levels delegates the query for supported logging levels to the next chain element
func (impl *ChainImpl) Levels() []logrus.Level {
	return impl.Next().Levels()
}
