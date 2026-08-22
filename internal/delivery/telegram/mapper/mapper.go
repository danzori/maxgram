package mapper

import (
	"strings"

	"github.com/mymmrac/telego"
)

type Kind int

const (
	KindText Kind = iota + 1
	KindService
	KindTopicCreated
)

type Update struct {
	Kind      Kind
	ChatID    int64
	ThreadID  int
	MessageID int
	Command   string
	Text      string
}

func Message(u telego.Update) (Update, bool) {
	msg := u.Message
	if msg == nil {
		return Update{}, false
	}

	return Update{
		Kind:      kindOf(msg),
		ChatID:    msg.Chat.ID,
		ThreadID:  msg.MessageThreadID,
		MessageID: msg.MessageID,
		Command:   command(msg.Text),
		Text:      msg.Text,
	}, true
}

func kindOf(msg *telego.Message) Kind {
	switch {
	case msg.ForumTopicCreated != nil:
		return KindTopicCreated
	case msg.ForumTopicEdited != nil,
		msg.ForumTopicClosed != nil,
		msg.ForumTopicReopened != nil,
		msg.GeneralForumTopicHidden != nil,
		msg.GeneralForumTopicUnhidden != nil:
		return KindService
	default:
		return KindText
	}
}

func command(text string) string {
	if text == "" || text[0] != '/' {
		return ""
	}

	name, _, _ := strings.Cut(text, " ")

	return name
}
