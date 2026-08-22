package bootstrap

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/danzori/maxgram/internal/config"
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

	<-ctx.Done()

	return nil
}
