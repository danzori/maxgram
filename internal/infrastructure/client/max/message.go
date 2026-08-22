package max

type Message struct {
	ID         string `json:"id"`
	Sender     int64  `json:"sender"`
	Text       string `json:"text"`
	Time       int64  `json:"time"`
	UpdateTime *int64 `json:"updateTime"`
}

type dispatchPayload struct {
	ChatID  int64    `json:"chatId"`
	Message *Message `json:"message"`
}

type sentMessageResponse struct {
	Message *Message `json:"message"`
}
