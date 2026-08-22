package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/danzori/maxgram/internal/domain/topic"
)

const topicsTable = "topics"

type TopicRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewTopicRepository(db *sql.DB) *TopicRepository {
	return &TopicRepository{
		db: db,
		sb: squirrel.StatementBuilder.RunWith(db),
	}
}

var topicColumns = []string{"chat_id", "thread_id", "name", "created_at"}

func (r *TopicRepository) ByChat(ctx context.Context, chatID int64) (topic.Topic, error) {
	row := r.sb.
		Select(topicColumns...).
		From(topicsTable).
		Where(squirrel.Eq{"chat_id": chatID}).
		QueryRowContext(ctx)

	return scanTopic(row, fmt.Sprintf("chat %d", chatID))
}

func (r *TopicRepository) All(ctx context.Context) ([]topic.Topic, error) {
	rows, err := r.sb.
		Select(topicColumns...).
		From(topicsTable).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var out []topic.Topic
	for rows.Next() {
		t, scanErr := scanTopic(rows, "")
		if scanErr != nil {
			return nil, scanErr
		}

		out = append(out, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}

	return out, nil
}

func (r *TopicRepository) Save(ctx context.Context, t topic.Topic) error {
	created := t.CreatedAt
	if created.IsZero() {
		created = time.Now()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("save topic %q: %w", t.Name, err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	_, err = r.sb.
		Delete(topicsTable).
		Where(squirrel.Eq{"thread_id": t.ThreadID}).
		Where(squirrel.NotEq{"chat_id": t.ChatID}).
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("clear stale mapping of thread %d: %w", t.ThreadID, err)
	}

	_, err = r.sb.
		Insert(topicsTable).
		Columns(topicColumns...).
		Values(t.ChatID, t.ThreadID, t.Name, created).
		Suffix(
			`
		ON CONFLICT(chat_id) DO UPDATE SET
			thread_id = excluded.thread_id,
			name = excluded.name
		`,
		).
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("save topic %q: %w", t.Name, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("save topic %q: %w", t.Name, err)
	}

	return nil
}

func (r *TopicRepository) Delete(ctx context.Context, chatID int64) error {
	_, err := r.sb.
		Delete(topicsTable).
		Where(squirrel.Eq{"chat_id": chatID}).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("delete topic of chat %d: %w", chatID, err)
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTopic(s scanner, subject string) (topic.Topic, error) {
	var (
		t       topic.Topic
		created int64
	)

	if err := s.Scan(&t.ChatID, &t.ThreadID, &t.Name, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return topic.Topic{}, fmt.Errorf("%w: %s", topic.ErrNotFound, subject)
		}

		return topic.Topic{}, fmt.Errorf("scan topic: %w", err)
	}

	t.CreatedAt = time.Unix(created, 0)

	return t, nil
}
