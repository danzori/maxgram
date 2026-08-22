package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danzori/maxgram/internal/application/bridge"
	"github.com/danzori/maxgram/internal/delivery/max/mapper"
	maxclient "github.com/danzori/maxgram/internal/infrastructure/client/max"
)

type Handler struct {
	client *maxclient.Client
	svc    *bridge.Service
	log    *slog.Logger
}

func New(client *maxclient.Client, svc *bridge.Service, log *slog.Logger) *Handler {
	return &Handler{
		client: client,
		svc:    svc,
		log:    log.With("component", "max.handler"),
	}
}

func (h *Handler) Handle(ctx context.Context, ev maxclient.Event) {
	switch ev.Kind {
	case maxclient.EventReady:
		h.onReady(ctx, ev)
	case maxclient.EventDisconnected:
		h.onDisconnected(ctx)
	case maxclient.EventMessage:
		h.onMessage(ctx, ev)
	}
}

func (h *Handler) onReady(ctx context.Context, ev maxclient.Event) {
	if err := h.svc.Bootstrap(ctx, ev.Chats); err != nil {
		h.log.Error("bootstrap forum", "err", err)
		h.svc.Notify(ctx, "MAX: could not prepare the topics, see the service logs.")

		return
	}

	h.svc.Notify(ctx, fmt.Sprintf("MAX: connected, %d chats.", len(ev.Chats)))
}

func (h *Handler) onDisconnected(ctx context.Context) {
	h.svc.Notify(ctx, "MAX: connection lost, reconnecting")
}

func (h *Handler) onMessage(ctx context.Context, ev maxclient.Event) {
	if ev.Message == nil {
		return
	}

	msg := mapper.Message(ctx, h.client.Directory(), ev.ChatID, ev.Message, ev.Own)
	if err := h.svc.Incoming(ctx, msg); err != nil {
		h.log.Error("forward message", "chat_id", ev.ChatID, "message_id", msg.ID, "err", err)
	}
}
