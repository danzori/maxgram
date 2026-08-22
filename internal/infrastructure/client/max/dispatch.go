package max

import (
	"encoding/json"
	"encoding/json/jsontext"
	"fmt"
)

func (c *Client) handle(data []byte) error {
	var pkt inPacket

	if err := json.Unmarshal(data, &pkt); err != nil {
		return fmt.Errorf("decode packet: %w", err)
	}

	c.log.Debug("packet", "opcode", int(pkt.Opcode), "cmd", pkt.Cmd, "payload", string(pkt.Payload))

	if pkt.Cmd == cmdError {
		return fmt.Errorf("%w: opcode %d: %s", ErrRemote, pkt.Opcode, string(pkt.Payload))
	}

	if pkt.Opcode == OpHandshake && pkt.Cmd == cmdResponse {
		c.log.Info("handshake accepted", "app_version", c.cfg.AppVersion)
	}

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
