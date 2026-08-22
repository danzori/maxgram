package max

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

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

func (c *Client) resetSeq() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.seq = 0
}
