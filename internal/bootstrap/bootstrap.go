package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/mymmrac/telego"

	"github.com/danzori/maxgram/internal/application/bridge"
	"github.com/danzori/maxgram/internal/config"
	maxhandler "github.com/danzori/maxgram/internal/delivery/max/handler"
	"github.com/danzori/maxgram/internal/delivery/max/websocket"
	tghandler "github.com/danzori/maxgram/internal/delivery/telegram/handler"
	"github.com/danzori/maxgram/internal/delivery/telegram/polling"
	maxclient "github.com/danzori/maxgram/internal/infrastructure/client/max"
	tgclient "github.com/danzori/maxgram/internal/infrastructure/client/telegram"
	"github.com/danzori/maxgram/internal/infrastructure/persistence/sqlite"
	logger "github.com/danzori/maxgram/internal/observability/logger/slog"
)

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := newApp(ctx, cfg)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := a.close(); closeErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "maxgram: close:", closeErr)
		}
	}()

	return a.run(ctx)
}

type app struct {
	log *slog.Logger
	db  *sql.DB

	max         *maxclient.Client
	maxListener *websocket.Listener
	tgListener  *polling.Listener
}

func newApp(ctx context.Context, cfg config.Config) (*app, error) {
	log := logger.New(cfg.Log)

	db, err := sqlite.Open(ctx, cfg.Storage.SQLitePath)
	if err != nil {
		return nil, err
	}

	bot, err := telego.NewBot(cfg.Telegram.BotToken, telego.WithDiscardLogger())
	if err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	me, err := bot.GetMe(ctx)
	if err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("verify telegram bot: %w", err)
	}

	log.Info("telegram bot ready", "username", me.Username)

	messenger := tgclient.New(bot, cfg.Telegram, log)
	maxClient := maxclient.New(cfg.Max, log)

	service := bridge.NewService(
		bridge.Settings{
			TopicMode:  cfg.Topics.Mode,
			ActiveDays: cfg.Topics.ActiveDays,
			SelfMode:   cfg.Topics.SelfMode,
		}, messenger,
		sqlite.NewTopicRepository(db),
		cfg.ExcludedSet(),
		log,
	)

	maxHandler := maxhandler.New(maxClient, service, log)
	tgHandler := tghandler.New(messenger, service, log, cfg.Telegram.ChatID)

	return &app{
		log:         log,
		db:          db,
		max:         maxClient,
		maxListener: websocket.New(maxClient.Events(), maxHandler.Handle),
		tgListener:  polling.New(bot, tgHandler.Handle, log),
	}, nil
}

func (a *app) run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error { return a.max.Run(ctx) })
	group.Go(func() error { return a.maxListener.Run(ctx) })
	group.Go(func() error { return a.tgListener.Run(ctx) })

	a.log.Info("maxgram started")

	return group.Wait()
}

func (a *app) close() error {
	return a.db.Close()
}
