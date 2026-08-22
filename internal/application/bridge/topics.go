package bridge

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/danzori/maxgram/internal/config"
	"github.com/danzori/maxgram/internal/domain/chat"
	"github.com/danzori/maxgram/internal/domain/topic"
)

const mainArea = 0

func (s *Service) Bootstrap(ctx context.Context, chats []chat.Chat) error {
	if err := s.tg.EnsureForum(ctx); err != nil {
		return fmt.Errorf("prepare forum: %w", err)
	}

	s.pending.Store(&chats)

	if !s.ready.Load() {
		ready, err := s.syncTopics(ctx)
		if err != nil {
			return fmt.Errorf("check forum state: %w", err)
		}
		if !ready {
			s.log.Info("no topics in telegram, waiting for /start", "chats", len(chats))

			return nil
		}

		s.ready.Store(true)
	}

	return s.provision(ctx, chats)
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.tg.EnsureForum(ctx); err != nil {
		return fmt.Errorf("prepare forum: %w", err)
	}

	s.syncTopics(ctx)

	var chats []chat.Chat
	if pending := s.pending.Load(); pending != nil {
		chats = *pending
	}

	if err := s.provision(ctx, chats); err != nil {
		return err
	}

	s.log.Info("forum provisioned", "chats", len(chats))

	return nil
}

func (s *Service) Reset(ctx context.Context) error {
	if err := s.tg.EnsureForum(ctx); err != nil {
		return fmt.Errorf("prepare forum: %w", err)
	}

	stored, err := s.topics.All(ctx)
	if err != nil {
		return fmt.Errorf("list topics: %w", err)
	}

	s.log.Info("deleting topics", "count", len(stored))

	for i, t := range stored {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := s.tg.DeleteTopic(ctx, t.ThreadID); err != nil {
			s.log.Warn("delete topic", "thread_id", t.ThreadID, "name", t.Name, "err", err)

			continue
		}

		s.log.Info("topic deleted", "progress", fmt.Sprintf("%d/%d", i+1, len(stored)), "name", t.Name)
	}

	for _, t := range stored {
		if err := s.topics.Delete(ctx, t.ChatID); err != nil {
			return fmt.Errorf("drop topic %q: %w", t.Name, err)
		}
	}

	s.ready.Store(false)
	s.log.Info("forum reset, waiting for /start")

	return nil
}

func (s *Service) syncTopics(ctx context.Context) (bool, error) {
	stored, err := s.topics.All(ctx)
	if err != nil {
		return false, err
	}
	if len(stored) == 0 {
		return false, nil
	}

	probe := stored[0]
	exists, err := s.tg.TopicExists(ctx, probe.ThreadID)
	if err != nil {
		s.log.Warn(
			"probe topic",
			"thread_id", probe.ThreadID,
			"name", probe.Name,
			"err", err,
		)

		return true, nil
	}

	if exists {
		return true, nil
	}

	s.log.Info(
		"stored topics are gone from telegram, forgetting them",
		"probed_thread_id", probe.ThreadID,
		"topics", len(stored),
	)

	for _, t := range stored {
		if err := s.topics.Delete(ctx, t.ChatID); err != nil {
			return false, fmt.Errorf("drop topic %q: %w", t.Name, err)
		}
	}

	return false, nil
}

func (s *Service) provision(ctx context.Context, chats []chat.Chat) error {
	s.ready.Store(true)

	wanted := s.chatsNeedingTopics(chats)
	if len(wanted) == 0 {
		return nil
	}

	s.log.Info("pre-creating topics", "mode", int(s.cfg.TopicMode), "chats", len(wanted))

	for _, c := range wanted {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := s.chatTopic(ctx, c.ID, c.DisplayTitle()); err != nil {
			s.log.Error("pre-create topic", "chat_id", c.ID, "title", c.DisplayTitle(), "err", err)
		}
	}

	return nil
}

func (s *Service) chatsNeedingTopics(chats []chat.Chat) []chat.Chat {
	if s.cfg.TopicMode == config.TopicModeLazy {
		return nil
	}

	var since time.Time
	if s.cfg.TopicMode == config.TopicModeActive {
		since = time.Now().AddDate(0, 0, -s.cfg.ActiveDays)
	}

	out := make([]chat.Chat, 0, len(chats))

	for _, c := range chats {
		if s.isExcluded(c.ID) {
			continue
		}

		if s.cfg.TopicMode == config.TopicModeActive && !c.ActiveSince(since) {
			continue
		}

		if c.Title == "" {
			s.log.Debug("chat has no title, deferring its topic", "chat_id", c.ID)

			continue
		}

		out = append(out, c)
	}

	if slices.ContainsFunc(
		out, func(c chat.Chat) bool {
			return !c.LastActive.IsZero()
		},
	) {
		slices.SortStableFunc(
			out, func(a, b chat.Chat) int {
				return a.LastActive.Compare(b.LastActive)
			},
		)
	} else {
		slices.Reverse(out)
	}

	return out
}

func (s *Service) chatTopic(ctx context.Context, chatID int64, title string) (topic.Topic, error) {
	s.topicMu.Lock()
	defer s.topicMu.Unlock()

	t, err := s.topics.ByChat(ctx, chatID)
	if err == nil {
		return t, nil
	}

	if !errors.Is(err, topic.ErrNotFound) {
		return topic.Topic{}, fmt.Errorf("lookup topic of chat %d: %w", chatID, err)
	}

	name := topic.NormalizeName(title)
	if name == "" {
		name = fmt.Sprint("chat %d", chatID)
	}

	threadID, err := s.tg.CreateTopic(ctx, name)
	if err != nil {
		return topic.Topic{}, fmt.Errorf("create topic %q: %w", name, err)
	}

	if threadID == mainArea {
		return topic.Topic{}, fmt.Errorf("create topic %q: %w", name, topic.ErrCreateDenied)
	}

	t = topic.Topic{
		ChatID:    chatID,
		ThreadID:  threadID,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.topics.Save(ctx, t); err != nil {
		return topic.Topic{}, fmt.Errorf("save topic %q: %w", name, err)
	}

	s.log.Info("topic created", "chat_id", chatID, "thread_id", threadID, "name", name)

	return t, nil
}
