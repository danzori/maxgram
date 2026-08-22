package max

import (
	"encoding/json/jsontext"
)

type Opcode int

const (
	OpHeartbeat    Opcode = 1
	OpHandshake    Opcode = 6
	OpAuthSnapshot Opcode = 19
	OpContactGet   Opcode = 32
	OpGetMessages  Opcode = 49
	OpChatMedia    Opcode = 51
	OpSendMessage  Opcode = 64
	OpVideoURL     Opcode = 83
	OpGetFileURL   Opcode = 88
	OpDispatch     Opcode = 128
)

const (
	cmdRequest  = 0
	cmdResponse = 1
	cmdError    = 3
)

type outPacket struct {
	Ver     int    `json:"ver"`
	Cmd     int    `json:"cmd"`
	Seq     int    `json:"seq"`
	Opcode  Opcode `json:"opcode"`
	Payload any    `json:"payload"`
}

type inPacket struct {
	Ver     int            `json:"ver"`
	Cmd     int            `json:"cmd"`
	Seq     int            `json:"seq"`
	Opcode  Opcode         `json:"opcode"`
	Payload jsontext.Value `json:"payload"`
}
