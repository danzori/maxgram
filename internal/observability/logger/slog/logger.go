package slog

import (
	stdslog "log/slog"
	"os"
	"strings"

	"github.com/danzori/maxgram/internal/config"
)

func New(cfg config.Log) *stdslog.Logger {
	opts := &stdslog.HandlerOptions{
		Level: level(cfg.Level),
	}

	var handler stdslog.Handler

	if strings.EqualFold(cfg.Format, "json") {
		handler = stdslog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = stdslog.NewTextHandler(os.Stdout, opts)
	}

	return stdslog.New(handler)
}

func level(name string) stdslog.Level {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return stdslog.LevelDebug
	case "warn", "warning":
		return stdslog.LevelWarn
	case "error", "err":
		return stdslog.LevelError
	default:
		return stdslog.LevelInfo
	}
}
