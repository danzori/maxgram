package bridge

import (
	"context"

	"github.com/danzori/maxgram/internal/domain/message"
	"github.com/danzori/maxgram/internal/domain/topic"
)

type Target struct {
	ThreadID int
	WithChat bool
}

type Delivery struct {
	Message message.Message
	Targets []Target
}

type Messenger interface {
	EnsureForum(ctx context.Context) error
	CreateTopic(ctx context.Context, name string) (int, error)
	DeleteTopic(ctx context.Context, threadID int) error
	TopicExists(ctx context.Context, threadID int) (bool, error)
	Deliver(ctx context.Context, d Delivery) error
	Notify(ctx context.Context, threadID int, text string) error
}

type MaxGateway interface {
	SendText(ctx context.Context, chatID int64, text string) (string, error)
}

type TopicRepository interface {
	ByChat(ctx context.Context, chatID int64) (topic.Topic, error)
	ByThread(ctx context.Context, threadID int) (topic.Topic, error)
	All(ctx context.Context) ([]topic.Topic, error)
	Save(ctx context.Context, t topic.Topic) error
	Delete(ctx context.Context, chatID int64) error
}
