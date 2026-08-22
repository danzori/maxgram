package handler

import (
	"context"
	"log/slog"

	"github.com/mymmrac/telego"

	"github.com/danzori/maxgram/internal/application/bridge"
	"github.com/danzori/maxgram/internal/delivery/telegram/mapper"
	tgclient "github.com/danzori/maxgram/internal/infrastructure/client/telegram"
)

type Handler struct {
	client *tgclient.Client
	svc    *bridge.Service
	log    *slog.Logger
	chatID int64
}

func New(client *tgclient.Client, svc *bridge.Service, log *slog.Logger, chatID int64) *Handler {
	return &Handler{
		client: client,
		svc:    svc,
		log:    log.With("component", "telegram.handler"),
		chatID: chatID,
	}
}

func (h *Handler) Handle(ctx context.Context, update telego.Update) {
	u, ok := mapper.Message(update)
	if !ok || u.ChatID != h.chatID {
		return
	}

	switch u.Kind {
	case mapper.KindTopicCreated:
		return
	case mapper.KindService:
		h.remove(ctx, u)

		return
	default:
	}

	if u.Command != "" && u.ThreadID != 0 {
		return
	}

	switch u.Command {
	case "/start":
		h.provision(
			ctx, u, h.svc.Start,
			"Checking the topics, this may take a minute.",
			"Done. Messages from Max will arrive in the topics.",
		)
	case "/reset":
		h.provision(
			ctx, u, h.svc.Reset,
			"Deleting the topics one by one. Telegram rate limits this, so it can take a few minutes - wait for the confirmation.",
			"Done, every topic is deleted. Send /start to set them up again.",
		)
	default:
	}
}

func (h *Handler) provision(ctx context.Context, u mapper.Update, run func(context.Context) error, notice, done string) {
	h.reply(ctx, u, notice)

	go func() {
		if err := run(ctx); err != nil {
			h.log.Error("provision", "command", u.Command, "err", err)
			h.reply(ctx, u, "Setup failed, see the service logs.")

			return
		}

		h.reply(ctx, u, done)
	}()
}

func (h *Handler) reply(ctx context.Context, to mapper.Update, text string) {
	if err := h.client.Reply(ctx, to.ThreadID, to.MessageID, text); err != nil {
		h.log.Error("reply", "thread_id", to.ThreadID, "err", err)
	}
}

func (h *Handler) remove(ctx context.Context, u mapper.Update) {
	if err := h.client.Remove(ctx, u.MessageID); err != nil {
		h.log.Debug("could not remove a forum service message", "message_id", u.MessageID, "err", err)

		return
	}

	h.log.Debug("removed a forum service message", "message_id", u.MessageID)
}
