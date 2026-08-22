package telegram

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/time/rate"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/danzori/maxgram/internal/config"
)

type Client struct {
	bot     *telego.Bot
	cfg     config.Telegram
	chatID  telego.ChatID
	log     *slog.Logger
	limiter *rate.Limiter
}

func New(bot *telego.Bot, cfg config.Telegram, log *slog.Logger) *Client {
	return &Client{
		bot:     bot,
		cfg:     cfg,
		chatID:  tu.ID(cfg.ChatID),
		log:     log.With("component", "telegram.client"),
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit/60), 5),
	}
}

func (c *Client) Notify(ctx context.Context, threadID int, text string) error {
	_, err := call(
		ctx, c, "sendMessage", func(ctx context.Context) (*telego.Message, error) {
			return c.bot.SendMessage(
				ctx, &telego.SendMessageParams{
					ChatID:          c.chatID,
					MessageThreadID: threadID,
					Text:            text,
					ParseMode:       telego.ModeHTML,
				},
			)
		},
	)

	return err
}

func (c *Client) Reply(ctx context.Context, threadID, replyTo int, text string) error {
	_, err := call(
		ctx, c, "sendMessage", func(ctx context.Context) (*telego.Message, error) {
			return c.bot.SendMessage(
				ctx, &telego.SendMessageParams{
					ChatID:              c.chatID,
					MessageThreadID:     threadID,
					Text:                text,
					DisableNotification: true,
					ReplyParameters:     &telego.ReplyParameters{MessageID: replyTo},
				},
			)
		},
	)

	return err
}

func (c *Client) Remove(ctx context.Context, messageID int) error {
	err := callVoid(
		ctx, c, "deleteMessage", func(ctx context.Context) error {
			return c.bot.DeleteMessage(
				ctx, &telego.DeleteMessageParams{
					ChatID:    c.chatID,
					MessageID: messageID,
				},
			)
		},
	)

	if apiErr, ok := errors.AsType[*telegoapi.Error](err); ok && apiErr.ErrorCode == 400 {
		return nil
	}

	return err
}
