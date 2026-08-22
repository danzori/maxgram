package bridge

import (
	"context"
	"fmt"

	"github.com/danzori/maxgram/internal/config"
	"github.com/danzori/maxgram/internal/domain/chat"
	"github.com/danzori/maxgram/internal/domain/message"
)

func (s *Service) Incoming(ctx context.Context, msg message.Message) error {
	if !s.ready.Load() {
		s.log.Debug("waiting for /start, message not forwarded", "chat_id", msg.ChatID)

		return nil
	}

	if s.isExcluded(msg.ChatID) {
		s.log.Debug("chat excluded", "chat_id", msg.ChatID)

		return nil
	}

	if msg.Own && s.cfg.SelfMode == config.SelfModeSkip {
		return nil
	}

	if err := msg.Validate(); err != nil {
		s.log.Debug("skip message", "chat_id", msg.ChatID, "err", err)

		return nil
	}

	targets := s.targets(ctx, msg)
	if len(targets) == 0 {
		return nil
	}

	if err := s.tg.Deliver(
		ctx, Delivery{
			Message: msg,
			Targets: targets,
		},
	); err != nil {
		return fmt.Errorf("deliver message from chat %d: %w", msg.ChatID, err)
	}

	return nil
}

func (s *Service) Outgoing(ctx context.Context, threadID int, text string) error {
	t, err := s.topics.ByThread(ctx, threadID)
	if err != nil {
		return err
	}

	if s.isExcluded(t.ChatID) {
		return fmt.Errorf("%w: chat %d", chat.ErrExcluded, t.ChatID)
	}

	if _, err = s.max.SendText(ctx, t.ChatID, text); err != nil {
		return err
	}

	return nil
}

func (s *Service) Bound(ctx context.Context, threadID int) bool {
	t, err := s.topics.ByThread(ctx, threadID)
	if err != nil {
		return false
	}

	return !s.isExcluded(t.ChatID)
}

func (s *Service) targets(ctx context.Context, msg message.Message) []Target {
	chatTopic, err := s.chatTopic(ctx, msg.ChatID, msg.DisplayChatTitle())
	if err != nil {
		s.log.Error("resolve chat topic", "chat_id", msg.ChatID, "err", err)

		return []Target{
			{
				ThreadID: mainArea,
				WithChat: true,
			},
		}
	}

	return []Target{{ThreadID: chatTopic.ThreadID}}
}
