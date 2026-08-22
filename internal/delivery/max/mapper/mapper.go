package mapper

import (
	"context"
	"time"

	"github.com/danzori/maxgram/internal/domain/chat"
	"github.com/danzori/maxgram/internal/domain/message"
	maxclient "github.com/danzori/maxgram/internal/infrastructure/client/max"
)

//nolint:revive // own describes message ownership, not control flow
func Message(ctx context.Context, dir *maxclient.Directory, chatID int64, raw *maxclient.Message, own bool) message.Message {
	msg := message.Message{
		ID:     raw.ID,
		ChatID: chatID,
		Sender: message.Sender{
			ID:   raw.Sender,
			Name: dir.UserName(ctx, raw.Sender),
		},
		Text: raw.Text,
		Own:  own,
	}

	if raw.Time > 0 {
		msg.SentAt = time.UnixMilli(raw.Time)
	}

	if c, ok := dir.Chat(chatID); ok {
		msg.ChatTitle = c.Title
		msg.ChatKind = c.Kind
	}

	if msg.ChatTitle == "" && msg.ChatKind == chat.KindDialog && !own && raw.Sender != 0 {
		msg.ChatTitle = msg.Sender.Name
	}

	return msg
}
