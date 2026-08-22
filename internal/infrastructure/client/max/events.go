package max

import (
	"context"
	"sync"

	"github.com/danzori/maxgram/internal/domain/chat"
)

type EventKind int

const (
	EventReady EventKind = iota + 1
	EventDisconnected
	EventMessage
)

type Event struct {
	Kind EventKind

	Chats []chat.Chat

	Message *Message
	ChatID  int64
	Own     bool
}

type queue struct {
	mu     sync.Mutex
	items  []Event
	notify chan struct{}
	closed bool
}

func newQueue() *queue {
	return &queue{notify: make(chan struct{}, 1)}
}

func (q *queue) push(e Event) int {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()

		return 0
	}

	q.items = append(q.items, e)
	depth := len(q.items)
	q.mu.Unlock()

	select {
	case q.notify <- struct{}{}:
	default:
	}

	return depth
}

func (q *queue) pop(ctx context.Context) (Event, bool) {
	for {
		q.mu.Lock()

		if len(q.items) > 0 {
			e := q.items[0]
			q.items = q.items[1:]
			q.mu.Unlock()

			return e, true
		}

		closed := q.closed

		q.mu.Unlock()

		if closed {
			return Event{}, false
		}

		select {
		case <-ctx.Done():
			return Event{}, false
		case <-q.notify:
		}
	}
}

func (q *queue) close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()

	select {
	case q.notify <- struct{}{}:
	default:
	}
}
