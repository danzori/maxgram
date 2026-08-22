package max

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"

	"github.com/coder/websocket"
)

type result struct {
	payload jsontext.Value
	err     error
}

func (c *Client) send(ctx context.Context, op Opcode, payload any) error {
	conn := c.conn.Load()
	if conn == nil {
		return ErrDisconnected
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	seq := c.nextSeq()

	return c.write(ctx, conn, seq, op, payload)
}

func (c *Client) call(ctx context.Context, op Opcode, payload any) (jsontext.Value, error) {
	conn := c.conn.Load()
	if conn == nil {
		return jsontext.Value{}, ErrDisconnected
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.RequestTimeout)
	defer cancel()

	ch := make(chan result, 1)

	c.writeMu.Lock()
	seq := c.nextSeq()

	c.mu.Lock()
	c.pending[seq] = ch
	c.mu.Unlock()

	err := c.write(ctx, conn, seq, op, payload)
	c.writeMu.Unlock()

	if err != nil {
		c.take(seq)

		return nil, err
	}

	select {
	case res := <-ch:
		return res.payload, res.err
	case <-ctx.Done():
		c.take(seq)

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: opcode %d", ErrTimeout, op)
		}

		return nil, ctx.Err()
	}
}

func (c *Client) callAs[T any](ctx context.Context, op Opcode, payload any) (T, error) {
	var out T

	raw, err := c.call(ctx, op, payload)
	if err != nil {
		return out, err
	}

	if err = json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode opcode %d: %w", op, err)
	}

	return out, nil
}

func (c *Client) write(ctx context.Context, conn *websocket.Conn, seq int, op Opcode, payload any) error {
	raw, err := json.Marshal(
		outPacket{
			Ver:     c.cfg.ProtocolVersion,
			Cmd:     cmdRequest,
			Seq:     seq,
			Opcode:  op,
			Payload: payload,
		},
	)
	if err != nil {
		return fmt.Errorf("encode packet: %w", err)
	}

	if err = conn.Write(ctx, websocket.MessageText, raw); err != nil {
		return fmt.Errorf("write opcode %d: %w", op, err)
	}

	return nil
}

func (c *Client) nextSeq() int {
	seq := c.seq
	c.seq++

	return seq
}

func (c *Client) take(seq int) (chan result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch, ok := c.pending[seq]
	if ok {
		delete(c.pending, seq)
	}

	return ch, ok
}

func (c *Client) resetSeq() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.seq = 0
}

func (c *Client) failPending(err error) {
	c.mu.Lock()
	pending := c.pending
	c.pending = make(map[int]chan result)
	c.mu.Unlock()

	for _, ch := range pending {
		ch <- result{err: err}
	}
}
