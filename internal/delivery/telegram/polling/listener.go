package polling

import (
	"context"
	"log/slog"

	"github.com/mymmrac/telego"
)

const timeout = 30

type Handler func(ctx context.Context, update telego.Update)

type Listener struct {
	bot     *telego.Bot
	handler Handler
	log     *slog.Logger
}

func New(bot *telego.Bot, handler Handler, log *slog.Logger) *Listener {
	return &Listener{
		bot:     bot,
		handler: handler,
		log:     log.With("component", "telegram.polling"),
	}
}

func (l *Listener) Run(ctx context.Context) error {
	updates, err := l.bot.UpdatesViaLongPolling(
		ctx, &telego.GetUpdatesParams{
			Timeout:        timeout,
			AllowedUpdates: []string{"message"},
		},
	)
	if err != nil {
		return err
	}

	l.log.Info("listening for replies")

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}

			l.handler(ctx, update)
		}
	}
}
