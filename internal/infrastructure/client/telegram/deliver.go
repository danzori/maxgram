package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"

	"github.com/danzori/maxgram/internal/application/bridge"
	"github.com/danzori/maxgram/internal/domain/message"
)

func (c *Client) Deliver(ctx context.Context, d bridge.Delivery) error {
	body := renderBody(d.Message)
	if body == "" {
		return nil
	}

	var (
		delivered int
		failures  []error
	)

	for _, target := range d.Targets {
		text := compose(d.Message, target, body, false)

		_, err := call(
			ctx, c, "sendMessage", func(ctx context.Context) (*telego.Message, error) {
				return c.bot.SendMessage(
					ctx, &telego.SendMessageParams{
						ChatID:          c.chatID,
						MessageThreadID: target.ThreadID,
						Text:            text,
						ParseMode:       telego.ModeHTML,
						LinkPreviewOptions: &telego.LinkPreviewOptions{
							IsDisabled: true,
						},
					},
				)
			},
		)
		if err != nil {
			failures = append(failures, fmt.Errorf("thread %d: %w", target.ThreadID, err))

			continue
		}
		delivered++
	}

	if delivered == 0 {
		return fmt.Errorf("%w: %w", message.ErrNotDelivered, errors.Join(failures...))
	}

	for _, err := range failures {
		c.log.Error("partial delivery failure", "err", err)
	}

	return nil
}

func compose(msg message.Message, target bridge.Target, body string, asCaption bool) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{header(msg, target.WithChat), body, timeFooter(msg)} {
		if part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, "\n")
}
