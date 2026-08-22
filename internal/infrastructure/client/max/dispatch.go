package max

import (
	"context"
	"encoding/json"
	"encoding/json/jsontext"
	"errors"
	"fmt"
)

func fatal(err error) bool {
	return errors.Is(err, ErrAuthRejected) ||
		errors.Is(err, ErrAuthTimeout) ||
		errors.Is(err, ErrSnapshot)
}

func (c *Client) handle(ctx context.Context, data []byte, a *attempt) error {
	var pkt inPacket

	if err := json.Unmarshal(data, &pkt); err != nil {
		return fmt.Errorf("decode packet: %w", err)
	}

	c.log.Debug("packet", "opcode", int(pkt.Opcode), "cmd", pkt.Cmd, "payload", string(pkt.Payload))

	if pkt.Cmd == cmdError {
		if pkt.Opcode == OpAuthSnapshot {
			err := fmt.Errorf(
				"%w: %s - check MAX_TOKEN and MAX_DEVICE_ID (snapshot_chats_count=%d)",
				ErrAuthRejected, remoteReason(pkt.Payload), c.cfg.ChatsSnapshot,
			)
			a.failed.CompareAndSwap(nil, &err)

			return err
		}

		return fmt.Errorf("%w: opcode %d: %s", ErrRemote, pkt.Opcode, string(pkt.Payload))
	}

	switch {
	case pkt.Opcode == OpHandshake && pkt.Cmd == cmdResponse:
		c.log.Info("handshake accepted, authorizing", "app_version", c.cfg.AppVersion, "snapshot_chats_count", c.cfg.ChatsSnapshot)

		return c.authorize(ctx)

	case pkt.Opcode == OpAuthSnapshot && pkt.Cmd == cmdResponse:
		return c.onAuthorized(pkt.Payload, a)

	case pkt.Opcode == OpHeartbeat:
		return nil

	default:
		c.log.Debug("unhandled packet", "opcode", int(pkt.Opcode), "cmd", pkt.Cmd, "payload", string(pkt.Payload))

		return nil
	}
}

func (c *Client) authorize(ctx context.Context) error {
	return c.send(
		ctx, OpAuthSnapshot, authPayload{
			ChatsCount:  c.cfg.ChatsSnapshot,
			Interactive: true,
			Token:       c.cfg.Token,
		},
	)
}

func (c *Client) onAuthorized(payload jsontext.Value, a *attempt) error {
	var snap snapshot

	if err := json.Unmarshal(payload, &snap); err != nil {
		return fmt.Errorf("%w: %w", ErrSnapshot, err)
	}

	c.selfID.Store(snap.Profile.id())
	a.authOK.Store(true)

	c.log.Info("authorized", "self_id", snap.Profile.id(), "chats", len(snap.Chats))

	return nil
}

type errorPayload struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func remoteReason(payload jsontext.Value) string {
	var p errorPayload

	if err := json.Unmarshal(payload, &p); err != nil || p.Error == "" {
		return string(payload)
	}

	return p.Error + ": " + p.Message
}
