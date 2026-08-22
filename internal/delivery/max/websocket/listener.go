package websocket

import (
	"context"

	maxclient "github.com/danzori/maxgram/internal/infrastructure/client/max"
)

type Handler func(ctx context.Context, ev maxclient.Event)

type Listener struct {
	events  <-chan maxclient.Event
	handler Handler
}

func New(events <-chan maxclient.Event, handler Handler) *Listener {
	return &Listener{
		events:  events,
		handler: handler,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-l.events:
			if !ok {
				return nil
			}

			l.handler(ctx, ev)
		}
	}
}
