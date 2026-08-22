package bridge

import (
	"context"
	"fmt"

	"github.com/danzori/maxgram/internal/config"
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

	targets, err := s.targets(ctx, msg)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	if err = s.tg.Deliver(
		ctx, Delivery{
			Message: msg,
			Targets: targets,
		},
	); err != nil {
		return fmt.Errorf("deliver message from chat %d: %w", msg.ChatID, err)
	}

	return nil
}

func (s *Service) targets(ctx context.Context, msg message.Message) ([]Target, error) {
	chatTopic, err := s.chatTopic(ctx, msg.ChatID, msg.DisplayChatTitle())
	if err != nil {
		s.log.Error("resolve chat topic", "chat_id", msg.ChatID, "err", err)

		return []Target{
			{
				ThreadID: mainArea,
				WithChat: true,
			},
		}, nil
	}

	return []Target{{ThreadID: chatTopic.ThreadID}}, nil
}
