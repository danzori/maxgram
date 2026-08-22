package config

import "time"

type TopicMode int

const (
	TopicModeLazy TopicMode = iota + 1
	TopicModeAll
	TopicModeActive
)

type SelfMode int

const (
	SelfModeSkip SelfMode = iota + 1
	SelfModeMirror
)

type Config struct {
	Max      Max
	Telegram Telegram
	Topics   Topics
	Storage  Storage
	Log      Log
}

type Max struct {
	Token    string
	DeviceID string

	WebSocketURL      string
	ReconnectInterval time.Duration
	HeartbeatInterval time.Duration
	RequestTimeout    time.Duration

	ProtocolVersion int
	AppVersion      string
	DeviceName      string
	UserAgent       string
	ClientHints     string
	Platform        string
	Locale          string
	Screen          string
	TimeZone        string

	ChatsSnapshot   int
	ExcludedChatIDs []int64

	CatchupEnabled bool
	CatchupLimit   int
}

type Telegram struct {
	BotToken string
	ChatID   int64

	SendTimeout time.Duration
	RateLimit   float64
	MaxRetries  int
}

type Topics struct {
	Mode       TopicMode
	ActiveDays int
	SelfMode   SelfMode
}

type Storage struct {
	SQLitePath string
	MediaTmp   string

	DedupTTL    time.Duration
	DeliveryTTL time.Duration
}

type Log struct {
	Level  string
	Format string
}
