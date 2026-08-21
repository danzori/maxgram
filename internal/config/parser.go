package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type parser struct {
	problems []error
}

func (p *parser) err() error {
	return errors.Join(p.problems...)
}

func (p *parser) fail(key string, err error) {
	p.problems = append(p.problems, fmt.Errorf("%s: %w", key, err))
}

func (p *parser) invalid(key, typ, value string) {
	p.fail(key, fmt.Errorf("%w: %s", ErrInvalidValue, fmt.Sprintf("expected %s, got %q", typ, value)))
}

func (p *parser) value(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func (p *parser) str(key, fallback string) string {
	if v := p.value(key); v != "" {
		return v
	}

	return fallback
}

func (p *parser) required(key string) string {
	v := p.value(key)
	if v == "" {
		p.fail(key, ErrMissingValue)
	}

	return v
}

func (p *parser) requiredInt64(key string) int64 {
	raw := p.required(key)
	if raw == "" {
		return 0
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		p.invalid(key, "integer", raw)

		return 0
	}

	return v
}

func (p *parser) integer(key string, fallback int) int {
	raw := p.value(key)
	if raw == "" {
		return fallback
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		p.invalid(key, "integer", raw)

		return fallback
	}

	return v
}

func (p *parser) float(key string, fallback float64) float64 {
	raw := p.value(key)
	if raw == "" {
		return fallback
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		p.invalid(key, "number", raw)

		return fallback
	}

	return v
}

func (p *parser) bool(key string, fallback bool) bool {
	raw := p.value(key)
	if raw == "" {
		return fallback
	}

	v, err := strconv.ParseBool(raw)
	if err != nil {
		p.invalid(key, "boolean", raw)
	}

	return v
}

func (p *parser) duration(key string, fallback time.Duration) time.Duration {
	raw := p.value(key)
	if raw == "" {
		return fallback
	}

	v, err := time.ParseDuration(raw)
	if err != nil {
		p.invalid(key, "duration", raw)

		return fallback
	}

	return v
}

func (p *parser) int64Slice(key string) []int64 {
	raw := p.value(key)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]int64, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		v, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			p.invalid(key, "integer", raw)

			return nil
		}

		out = append(out, v)
	}

	return out
}
