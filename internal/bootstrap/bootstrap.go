package bootstrap

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/danzori/maxgram/internal/config"
	"github.com/danzori/maxgram/internal/delivery/max/websocket"
	maxclient "github.com/danzori/maxgram/internal/infrastructure/client/max"
	logger "github.com/danzori/maxgram/internal/observability/logger/slog"
)

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.New(cfg.Log)
	log.Info("maxgram started")

	client := maxclient.New(cfg.Max, log)
	listener := websocket.New(
		client.Events(), func(_ context.Context, ev maxclient.Event) {
			if ev.Kind == maxclient.EventMessage {
				log.Info("message", "chat_id", ev.ChatID, "sender", ev.Message.Sender, "text", ev.Message.Text)
			}
		},
	)

	go func() {
		_ = listener.Run(ctx)
	}()

	return client.Run(ctx)
}
