package telegram

import (
	"fmt"
	"html"

	"github.com/danzori/maxgram/internal/domain/chat"
	"github.com/danzori/maxgram/internal/domain/message"
)

const (
	timeLayout     = "15:04"
	ownSenderLabel = "You"
)

//nolint:revive // withChat selects between two presentation contexts
func header(msg message.Message, withChat bool) string {
	oneVoice := msg.ChatKind == chat.KindDialog || msg.ChatKind == chat.KindChannel

	if !withChat {
		if oneVoice && !msg.Own {
			return ""
		}

		return fmt.Sprintf("<b>%s</b>", esc(senderLabel(msg)))
	}

	title := fmt.Sprintf("<b>%s</b>", esc(msg.DisplayChatTitle()))

	switch msg.ChatKind {
	case chat.KindDialog:
	case chat.KindChannel:
		title = "Channel: " + title
	default:
		title = "Chat: " + title
	}

	if oneVoice && !msg.Own {
		return title
	}

	return fmt.Sprintf("%s | From <b>%s</b>", title, esc(senderLabel(msg)))
}

func timeFooter(msg message.Message) string {
	if msg.SentAt.IsZero() {
		return ""
	}

	return "\n<blockquote>Sent at " + msg.SentAt.Format(timeLayout) + "</blockquote>"
}

func senderLabel(msg message.Message) string {
	if msg.Own {
		return ownSenderLabel
	}

	return msg.Sender.DisplayName()
}

func renderBody(msg message.Message) string {
	if msg.Text == "" {
		return ""
	}

	return esc(msg.Text)
}

func esc(s string) string {
	return html.EscapeString(s)
}
