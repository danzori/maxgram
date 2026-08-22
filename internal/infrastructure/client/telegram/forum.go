package telegram

import (
	"context"
	"errors"
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"

	"github.com/danzori/maxgram/internal/domain/topic"
)

func (c *Client) EnsureForum(ctx context.Context) error {
	info, err := call(
		ctx, c, "getChat", func(ctx context.Context) (*telego.ChatFullInfo, error) {
			return c.bot.GetChat(ctx, &telego.GetChatParams{ChatID: c.chatID})
		},
	)
	if err != nil {
		return fmt.Errorf("read chat %d: %w", c.cfg.ChatID, err)
	}

	me, err := call(
		ctx, c, "getMe", func(ctx context.Context) (*telego.User, error) {
			return c.bot.GetMe(ctx)
		},
	)
	if err != nil {
		return fmt.Errorf("identify the bot: %w", err)
	}

	return validateSetup(c.cfg.ChatID, info.Type, me)
}

func (c *Client) CreateTopic(ctx context.Context, name string) (int, error) {
	created, err := call(
		ctx, c, "CreateForumTopic", func(ctx context.Context) (*telego.ForumTopic, error) {
			return c.bot.CreateForumTopic(
				ctx, &telego.CreateForumTopicParams{
					ChatID: c.chatID,
					Name:   name,
				},
			)
		},
	)
	if err != nil {
		if apiErr, ok := errors.AsType[*telegoapi.Error](err); ok && apiErr.ErrorCode == 400 {
			return 0, fmt.Errorf("%w: %s", topic.ErrCreateDenied, apiErr.Description)
		}

		return 0, err
	}

	if created.MessageThreadID == 0 {
		return 0, fmt.Errorf("%w: telegram returned no thread id for topic %q", topic.ErrCreateDenied, name)
	}

	c.log.Info("topic created", "name", name, "thread_id", created.MessageThreadID)

	return created.MessageThreadID, nil
}

func (c *Client) DeleteTopic(ctx context.Context, threadID int) error {
	if threadID == 0 {
		return fmt.Errorf("%w: refusing to delete the built-in General topic", topic.ErrCreateDenied)
	}

	err := callVoid(
		ctx, c, "deleteForumTopic", func(ctx context.Context) error {
			return c.bot.DeleteForumTopic(
				ctx, &telego.DeleteForumTopicParams{
					ChatID:          c.chatID,
					MessageThreadID: threadID,
				},
			)
		},
	)

	if apiErr, ok := errors.AsType[*telegoapi.Error](err); ok && apiErr.ErrorCode == 400 {
		return nil
	}

	return err
}

func (c *Client) TopicExists(ctx context.Context, threadID int) (bool, error) {
	err := callVoid(
		ctx, c, "sendChatAction", func(ctx context.Context) error {
			return c.bot.SendChatAction(
				ctx, &telego.SendChatActionParams{
					ChatID:          c.chatID,
					MessageThreadID: threadID,
					Action:          telego.ChatActionTyping,
				},
			)
		},
	)
	if err == nil {
		return true, nil
	}

	if apiErr, ok := errors.AsType[*telegoapi.Error](err); ok && apiErr.ErrorCode == 400 {
		return false, nil
	}

	return false, err
}

func validateSetup(chatID int64, chatType string, me *telego.User) error {
	switch chatType {
	case telego.ChatTypePrivate:
		if !me.HasTopicsEnabled {
			return fmt.Errorf(
				"%w: Turn it on in @BotFather -> Open Mini App -> Select bot -> Bot settings -> Turn on \"Threaded Mode\"",
				topic.ErrForumRequired,
			)
		}

		return nil
	default:
		return fmt.Errorf(
			"%w: TG_CHAT_ID=%d is a %q chat. Use a private chat with the bot",
			topic.ErrForumRequired, chatID, chatType,
		)
	}
}
