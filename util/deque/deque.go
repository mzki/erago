package deque

import (
	"sync"
)

// represents event queue.
// references for golang.org/x/exp/shiny/driver/internal/event/event.go
type EventDeque struct {
	first []interface{} // LIFO
	last  []interface{} // FIFO

	mu   *sync.Mutex
	cond *sync.Cond
}

func NewEventDeque() EventDeque {
	mu := new(sync.Mutex)
	return EventDeque{
		mu:   mu,
		cond: sync.NewCond(mu),
	}
}

// send event to last of buffer.
func (b *EventDeque) Send(v interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.last = append(b.last, v)
	b.cond.Signal()
}

// send event to first of buffer.
func (b *EventDeque) SendFirst(v interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.first = append(b.first, v)
	b.cond.Signal()
}

// return buffered event.
// it will block until any event is occured.
func (b *EventDeque) NextEvent() interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	for {
		if l := len(b.first); l > 0 {
			e := b.first[l-1]
			b.first = b.first[:l-1]
			return e
		}

		if l := len(b.last); l > 0 {
			e := b.last[0]
			b.last = b.last[1:]
			return e
		}

		b.cond.Wait()
	}
}
