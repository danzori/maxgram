package message

import (
	"fmt"
	"strconv"
	"time"

	"github.com/danzori/maxgram/internal/domain/chat"
)

type QuoteKind string

const (
	QuoteReply   QuoteKind = "reply"
	QuoteForward QuoteKind = "forward"
)

type Sender struct {
	ID   int64
	Name string
}

func (s Sender) DisplayName() string {
	if s.Name != "" {
		return s.Name
	}

	if s.ID != 0 {
		return strconv.FormatInt(s.ID, 10)
	}

	return "unknown"
}

type Quote struct {
	Kind        QuoteKind
	SenderName  string
	Text        string
	Attachments []Attachment
}

type Message struct {
	ID     string
	SentAt time.Time
	Own    bool

	ChatID    int64
	ChatTitle string
	ChatKind  chat.Kind
	Sender    Sender

	Text        string
	Quote       *Quote
	Attachments []Attachment
}

func (m Message) IsEmpty() bool {
	if m.Text != "" || len(m.Attachments) > 0 {
		return false
	}

	if m.Quote != nil && (m.Quote.Text != "" || len(m.Quote.Attachments) > 0) {
		return false
	}

	return true
}

func (m Message) Validate() error {
	if m.ChatID == 0 {
		return ErrNoChat
	}

	if m.IsEmpty() {
		return ErrEmpty
	}

	return nil
}

func (m Message) DisplayChatTitle() string {
	if m.ChatTitle != "" {
		return m.ChatTitle
	}

	return fmt.Sprintf("chat %d", m.ChatID)
}
