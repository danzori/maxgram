package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

func Load(envFile string) (Config, error) {
	if err := godotenv.Load(envFile); err != nil {
		return Config{}, fmt.Errorf("load %s: %w", envFile, err)
	}

	var (
		cfg Config
		p   parser
	)

	cfg.Max = Max{
		Token:    p.required("MAX_TOKEN"),
		DeviceID: p.required("MAX_DEVICE_ID"),

		WebSocketURL:      p.str("MAX_WS_URL", DefaultWebSocketURL),
		ReconnectInterval: p.duration("MAX_RECONNECT_INTERVAL", DefaultReconnectInterval),
		HeartbeatInterval: p.duration("MAX_HEARTBEAT_INTERVAL", DefaultHeartbeatInterval),
		RequestTimeout:    p.duration("MAX_REQUEST_TIMEOUT", DefaultRequestTimeout),

		ProtocolVersion: p.integer("MAX_PROTOCOL_VERSION", DefaultProtocolVersion),
		AppVersion:      p.str("MAX_APP_VERSION", DefaultAppVersion),
		DeviceName:      p.str("MAX_DEVICE_NAME", DefaultDeviceName),
		UserAgent:       p.str("MAX_USR_AGENT", DefaultUserAgent),
		ClientHints:     p.str("MAX_CLIENT_HINTS", DefaultClientHints),
		Platform:        p.str("MAX_PLATFORM", DefaultPlatform),
		Locale:          p.str("MAX_LOCALE", DefaultLocale),
		Screen:          p.str("MAX_SCREEN", DefaultScreen),
		TimeZone:        p.str("MAX_TIMEZONE", DefaultTimeZone),

		ChatsSnapshot:   p.integer("MAX_CHATS_SNAPSHOT", DefaultChatsSnapshot),
		ExcludedChatIDs: p.int64Slice("MAX_EXCLUDED_CHAT_IDS"),

		CatchupEnabled: p.bool("MAX_CATCHUP_ENABLED", DefaultCatchupEnabled),
		CatchupLimit:   p.integer("MAX_CATCHUP_LIMIT", DefaultCatchupLimit),
	}

	cfg.Telegram = Telegram{
		BotToken: p.required("TG_BOT_TOKEN"),
		ChatID:   p.requiredInt64("TG_CHAT_ID"),

		SendTimeout: p.duration("TG_SEND_TIMEOUT", DefaultSendTimeout),
		RateLimit:   p.float("TG_RATE_LIMIT", DefaultRateLimit),
		MaxRetries:  p.integer("TG_MAX_RETRIES", DefaultMaxRetries),
	}

	cfg.Topics = Topics{
		Mode:       TopicMode(p.integer("TOPIC_MODE", int(DefaultTopicMode))),
		ActiveDays: p.integer("TOPIC_ACTIVE_DAYS", DefaultActiveDays),
		SelfMode:   SelfMode(p.integer("TOPIC_SELF_MODE", int(DefaultSelfMode))),
	}

	cfg.Storage = Storage{
		SQLitePath: p.str("SQLITE_PATH", DefaultSQLitePath),
		MediaTmp:   p.str("MEDIA_TMP_DIR", DefaultMediaTmp),

		DedupTTL:    p.duration("DEDUP_TTL", DefaultDedupTTL),
		DeliveryTTL: p.duration("DELIVERY_TTL", DefaultDeliveryTTL),
	}

	cfg.Log = Log{
		Level:  p.str("LOG_LEVEL", DefaultLogLevel),
		Format: p.str("LOG_FORMAT", DefaultLogFormat),
	}

	if err := p.err(); err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.Max.ChatsSnapshot <= 0 || c.Max.ChatsSnapshot > MaxChatsSnapshot {
		return fmt.Errorf("%w: MAX_CHATS_SNAPSHOT must be between 1 ad %d", ErrInvalidValue, MaxChatsSnapshot)
	}

	if c.Telegram.RateLimit <= 0 {
		return fmt.Errorf("%w: TG_RATE_LIMIT must be > 0", ErrInvalidValue)
	}

	switch c.Topics.Mode {
	case TopicModeLazy, TopicModeAll:
	case TopicModeActive:
		if c.Topics.ActiveDays <= 0 {
			return fmt.Errorf("%w: TOPIC_ACTIVE_DAYS must be > 0 when TOPIC_MODE=3", ErrInvalidValue)
		}
	default:
		return fmt.Errorf("%w: TOPIC_MODE must be 1, 2 or 3", ErrInvalidValue)
	}

	switch c.Topics.SelfMode {
	case SelfModeSkip, SelfModeMirror:
	default:
		return fmt.Errorf("%w: TOPIC_SELF_MODE must be 1 or 2", ErrInvalidValue)
	}

	return nil
}

func (c Config) ExcludedSet() map[int64]struct{} {
	return idSet(c.Max.ExcludedChatIDs)
}

func idSet(ids []int64) map[int64]struct{} {
	set := make(map[int64]struct{}, len(ids))

	for _, id := range ids {
		set[id] = struct{}{}
	}

	return set
}
