package telegram

import (
	"context"
	"log/slog"

	"golang.org/x/time/rate"

	"github.com/mymmrac/telego"
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
