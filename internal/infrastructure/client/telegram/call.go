package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mymmrac/telego/telegoapi"
)

func call[T any](ctx context.Context, c *Client, op string, fn func(context.Context) (T, error)) (T, error) {
	var (
		zero    T
		lastErr error
	)

	for attempt := 1; attempt <= max(c.cfg.MaxRetries, 1); attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return zero, err
		}

		reqCtx, reqCancel := context.WithTimeout(ctx, c.cfg.SendTimeout)

		out, err := fn(reqCtx)

		reqCancel()

		if err == nil {
			return out, nil
		}

		lastErr = err

		wait, retry := retryDelay(err, attempt)
		if !retry {
			return zero, fmt.Errorf("%s: %w", op, err)
		}

		c.log.Warn(
			"telegram request failed, retrying",
			"op", op,
			"attempt", attempt,
			"wait", wait,
			"err", err,
		)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(wait):
		}
	}

	return zero, fmt.Errorf("%s: %w", op, lastErr)
}

func callVoid(ctx context.Context, c *Client, op string, fn func(context.Context) error) error {
	_, err := call(
		ctx, c, op, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, fn(ctx)
		},
	)

	return err
}

func retryDelay(err error, attempt int) (time.Duration, bool) {
	if apiErr, ok := errors.AsType[*telegoapi.Error](err); ok {
		if apiErr.Parameters != nil && apiErr.Parameters.RetryAfter > 0 {
			return time.Duration(apiErr.Parameters.RetryAfter) * time.Second, true
		}

		if apiErr.ErrorCode >= 400 && apiErr.ErrorCode < 500 && apiErr.ErrorCode != 429 {
			return 0, false
		}
	}

	return time.Duration(attempt) * 2 * time.Second, true
}
