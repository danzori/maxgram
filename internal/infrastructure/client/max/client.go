package max

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"github.com/danzori/maxgram/internal/config"
)

type Client struct {
	cfg config.Max
	log *slog.Logger

	conn atomic.Pointer[websocket.Conn]

	writeMu sync.Mutex
	seq     int

	selfID atomic.Int64

	queue  *queue
	events chan Event
}

func New(cfg config.Max, log *slog.Logger) *Client {
	return &Client{
		cfg:    cfg,
		log:    log.With("component", "max.client"),
		queue:  newQueue(),
		events: make(chan Event),
	}
}

func (c *Client) Events() <-chan Event {
	return c.events
}

func (c *Client) Run(ctx context.Context) error {
	go c.pump(ctx)
	defer c.queue.close()

	for {
		started := time.Now()
		err := c.connect(ctx)

		if ctx.Err() != nil {
			return nil //nolint:nilerr // context cancellation is a normal shutdown
		}

		if err != nil && !errors.Is(err, context.Canceled) {
			c.log.Error("session ended", "err", err, "uptime", time.Since(started).Round(time.Second))
		}

		c.queue.push(Event{Kind: EventDisconnected})

		delay := c.backoff(err)
		c.log.Info("reconnecting", "in", delay)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.cfg.ReconnectInterval):
		}
	}
}

func (c *Client) pump(ctx context.Context) {
	defer close(c.events)

	for {
		e, ok := c.queue.pop(ctx)
		if !ok {
			return
		}

		select {
		case c.events <- e:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) backoff(err error) time.Duration {
	if errors.Is(err, ErrAuthRejected) || errors.Is(err, ErrAuthTimeout) {
		return max(c.cfg.ReconnectInterval, authFailureBackoff)
	}

	return c.cfg.ReconnectInterval
}
