package topic

import (
	"strings"
	"time"
	"unicode/utf8"
)

const MaxNameLength = 128

type Topic struct {
	ChatID   int64
	ThreadID int

	Name      string
	CreatedAt time.Time
}

func NormalizeName(raw string) string {
	name := strings.Join(strings.Fields(raw), " ")

	if utf8.RuneCountInString(name) > MaxNameLength {
		name = string([]rune(name)[:MaxNameLength])
	}

	return name
}
