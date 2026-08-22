package config

import "time"

const (
	DefaultWebSocketURL      = "wss://ws-api.oneme.ru/websocket"
	DefaultReconnectInterval = 5 * time.Second
	DefaultHeartbeatInterval = 30 * time.Second
	DefaultRequestTimeout    = 15 * time.Second

	DefaultProtocolVersion = 10
	DefaultAppVersion      = "26.8.4"
	DefaultDeviceName      = "Chrome"
	DefaultUserAgent       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"
	DefaultClientHints     = `"Not=A?Brand";v="99", "Google Chrome";v="151", "Chromium";v="151"`
	DefaultPlatform        = "macOS"
	DefaultLocale          = "ru"
	DefaultScreen          = "1080x1920 2.0x"
	DefaultTimeZone        = "Europe/Moscow"

	DefaultChatsSnapshot = 100
	MaxChatsSnapshot     = 127

	DefaultCatchupEnabled = true
	DefaultCatchupLimit   = 50
)

const (
	DefaultSendTimeout = 120 * time.Second
	DefaultRateLimit   = 18.0
	DefaultMaxRetries  = 3
)

const (
	DefaultTopicMode  = TopicModeAll
	DefaultActiveDays = 0
	DefaultSelfMode   = SelfModeSkip
)

const (
	DefaultSQLitePath = "./data/maxgram.db"
	DefaultMediaTmp   = ""

	DefaultDedupTTL    = time.Hour
	DefaultDeliveryTTL = 7 * 24 * time.Hour
)

const (
	DefaultLogLevel  = "info"
	DefaultLogFormat = "text"
)
