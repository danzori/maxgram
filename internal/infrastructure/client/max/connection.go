package max

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

const (
	readLimit   = 32 << 20
	dialTimeout = 30 * time.Second
)

func (c *Client) connect(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
	defer dialCancel()

	conn, _, err := websocket.Dial(
		dialCtx, c.cfg.WebSocketURL, &websocket.DialOptions{
			HTTPHeader: c.webSocketHeaders(),
		},
	)
	if err != nil {
		return fmt.Errorf("dial %s: %w", c.cfg.WebSocketURL, err)
	}

	defer func() {
		_ = conn.CloseNow()
	}()
	conn.SetReadLimit(readLimit)

	c.log.Info("connected", "url", c.cfg.WebSocketURL)

	c.conn.Store(conn)
	c.resetSeq()

	defer c.conn.Store(nil)

	if err = c.send(ctx, OpHandshake, c.handshake()); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}

	go c.heartbeat(ctx)

	for {
		_, data, readErr := conn.Read(ctx)
		if readErr != nil {
			return fmt.Errorf("read: %w", readErr)
		}

		if err = c.handle(data); err != nil {
			c.log.Error("handle packet", "err", err)
		}
	}
}

func (c *Client) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.send(ctx, OpHeartbeat, heartbeatPayload{Interactive: false}); err != nil {
				c.log.Debug("heartbeat failed", "err", err)

				return
			}
		}
	}
}

func (c *Client) handshake() handshakePayload {
	return handshakePayload{
		DeviceID: c.cfg.DeviceID,
		UserAgent: userAgent{
			DeviceName:      c.cfg.DeviceName,
			DeviceType:      "WEB",
			PushDeviceType:  "WEBPUSH",
			DeviceLocale:    c.cfg.Locale,
			OSVersion:       c.cfg.Platform,
			HeaderUserAgent: c.cfg.UserAgent,
			AppVersion:      c.cfg.AppVersion,
			Screen:          c.cfg.Screen,
			TimeZone:        c.cfg.TimeZone,
		},
	}
}

func (c *Client) webSocketHeaders() http.Header {
	h := c.browserHeaders()

	h.Set("Origin", "https://web.max.ru")
	h.Set("Cache-Control", "no-cache")
	h.Set("Pragma", "no-cache")

	return h
}

func (c *Client) browserHeaders() http.Header {
	h := http.Header{}

	h.Set("User-Agent", c.cfg.UserAgent)
	h.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	h.Set("Sec-Ch-Ua", c.cfg.ClientHints)
	h.Set("Sec-Ch-Ua-Mobile", "?0")
	h.Set("Sec-Ch-Ua-Platform", `"`+c.cfg.Platform+`"`)

	return h
}
