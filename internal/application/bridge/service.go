package bridge

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/danzori/maxgram/internal/config"
	"github.com/danzori/maxgram/internal/domain/chat"
)

type Settings struct {
	TopicMode  config.TopicMode
	ActiveDays int
	SelfMode   config.SelfMode
}

type Service struct {
	cfg      Settings
	tg       Messenger
	topics   TopicRepository
	excluded map[int64]struct{}
	log      *slog.Logger

	topicMu sync.Mutex

	ready   atomic.Bool
	pending atomic.Pointer[[]chat.Chat]
}

func NewService(cfg Settings, tg Messenger, topics TopicRepository, excluded map[int64]struct{}, log *slog.Logger) *Service {
	return &Service{
		cfg:      cfg,
		tg:       tg,
		topics:   topics,
		excluded: excluded,
		log:      log.With("component", "bridge"),
	}
}

func (s *Service) Ready() bool {
	return s.ready.Load()
}

func (s *Service) Notify(ctx context.Context, text string) {
	if !s.ready.Load() {
		return
	}

	if err := s.tg.Notify(ctx, mainArea, text); err != nil {
		s.log.Error("notify", "err", err)
	}
}

func (s *Service) isExcluded(chatID int64) bool {
	_, ok := s.excluded[chatID]

	return ok
}
