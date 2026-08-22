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
}

func New(cfg config.Max, log *slog.Logger) *Client {
	return &Client{
		cfg: cfg,
		log: log.With("component", "max.client"),
	}
}

func (c *Client) Run(ctx context.Context) error {
	for {
		started := time.Now()
		err := c.connect(ctx)

		if ctx.Err() != nil {
			return nil //nolint:nilerr // context cancellation is a normal shutdown
		}

		if err != nil && !errors.Is(err, context.Canceled) {
			c.log.Error("session ended", "err", err, "uptime", time.Since(started).Round(time.Second))
		}

		c.log.Info("reconnecting", "in", c.cfg.ReconnectInterval)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.cfg.ReconnectInterval):
		}
	}
}
