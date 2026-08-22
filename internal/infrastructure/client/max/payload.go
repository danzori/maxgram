package max

type sendMessagePayload struct {
	ChatID  int64        `json:"chatId"`
	Message outboundBody `json:"message"`
	Notify  bool         `json:"notify"`
}

type outboundBody struct {
	Text     string `json:"text"`
	CID      int64  `json:"cid"`
	Elements []any  `json:"elements"`
}

type getMessagesPayload struct {
	ChatID      int64 `json:"chatId"`
	From        int64 `json:"from"`
	Forward     int   `json:"forward"`
	Backward    int   `json:"backward"`
	GetMessages bool  `json:"getMessages"`
}

type chatMediaPayload struct {
	ChatID      int64    `json:"chatId"`
	MessageID   int64    `json:"messageId"`
	AttachTypes []string `json:"attachTypes"`
	Forward     int      `json:"forward"`
	Backward    int      `json:"backward"`
}

type mediaURLPayload struct {
	ChatID      int64  `json:"chatId"`
	MessageID   string `json:"messageId"`
	VideoID     int64  `json:"videoId"`
	AudioID     int64  `json:"audioId"`
	ContentType int    `json:"contentType"`
}

type fileURLPayload struct {
	ChatID    int64  `json:"chatId"`
	MessageID string `json:"messageId"`
	FileID    int64  `json:"fileId"`
	Token     string `json:"token"`
}

type userAgent struct {
	DeviceName      string `json:"deviceName"`
	DeviceType      string `json:"deviceType"`
	PushDeviceType  string `json:"pushDeviceType"`
	DeviceLocale    string `json:"deviceLocale"`
	OSVersion       string `json:"osVersion"`
	HeaderUserAgent string `json:"headerUserAgent"`
	IsPwa           bool   `json:"isPwa"`
	AppVersion      string `json:"appVersion"`
	Screen          string `json:"screen"`
	TimeZone        string `json:"timezone"`
}

type handshakePayload struct {
	UserAgent userAgent `json:"userAgent"`
	DeviceID  string    `json:"deviceId"`
}

type authPayload struct {
	ChatsCount  int    `json:"chatsCount"`
	Interactive bool   `json:"interactive"`
	Token       string `json:"token"`
}

type heartbeatPayload struct {
	Interactive bool `json:"interactive"`
}

type contactGetPayload struct {
	ContactIDs []int64 `json:"contactIds"`
}
