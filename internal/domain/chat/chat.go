package chat

import (
	"fmt"
	"time"
)

type Kind string

const (
	KindDialog  Kind = "DIALOG"
	KindChannel Kind = "CHANNEL"
)

type Chat struct {
	ID   int64
	Kind Kind

	Title  string
	PeerID int64

	LastActive time.Time
}

func (c Chat) DisplayTitle() string {
	if c.Title != "" {
		return c.Title
	}

	if c.PeerID != 0 {
		return fmt.Sprintf("id %d", c.PeerID)
	}

	return fmt.Sprintf("chat %d", c.ID)
}

func (c Chat) ActiveSince(since time.Time) bool {
	return !c.LastActive.IsZero() && c.LastActive.After(since)
}
