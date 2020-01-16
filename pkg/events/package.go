package events

import (
	"sync"

	"github.com/fergusn/muzeum/pkg/model"
)

type PackageEvents struct {
	Pulled PulledEvent
	Pushed PushedEvent
}

type PulledEvent struct {
	subscribers []chan *Pulled
	lock        sync.RWMutex
}
type PushedEvent struct {
	subscribers []chan *Pushed
	lock        sync.RWMutex
}

// Pulled is emitted by repositories when a packaged was pulled from the repository
type Pulled struct {
	Registry string
	Package  *model.Package
	Location string
	Size     int64
}

// Pushed is emitted by repositories when a packaged was pushed to the repository
type Pushed struct {
	Registry string
	Package  *model.Package
	Token    string
	Location string
}

// Receive return a channel that send all Pull events
func (e *PulledEvent) Receive() <-chan *Pulled {
	e.lock.Lock()
	defer e.lock.Unlock()

	c := make(chan *Pulled)
	e.subscribers = append(e.subscribers, c)
	return c
}

// Emit published a Pulled event
func (e *PulledEvent) Emit(ev *Pulled) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	for _, x := range e.subscribers {
		x <- ev
	}
}

// Receive subscribe to a Pushed event
func (e *PushedEvent) Receive() chan *Pushed {
	e.lock.Lock()
	defer e.lock.Unlock()

	c := make(chan *Pushed, 5)
	e.subscribers = append(e.subscribers, c)
	return c
}
func (e *PushedEvent) Emit(ev *Pushed) {
	e.lock.RLock()
	e.lock.RUnlock()

	for _, x := range e.subscribers {
		x <- ev
	}
}

var Package PackageEvents = PackageEvents{
	Pulled: PulledEvent{
		subscribers: []chan *Pulled{},
	},
}
