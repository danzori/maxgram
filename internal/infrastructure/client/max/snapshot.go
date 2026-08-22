package max

import (
	"strings"
)

type name struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Name      string `json:"name"`
}

func (n name) display() string {
	full := strings.TrimSpace(n.FirstName + " " + n.LastName)
	if full != "" {
		return full
	}

	return strings.TrimSpace(n.Name)
}

type profile struct {
	Contact *struct {
		ID    int64  `json:"id"`
		Names []name `json:"names"`
	} `json:"contact"`
}

func (p profile) id() int64 {
	if p.Contact != nil {
		return p.Contact.ID
	}

	return 0
}

type lastMessagePreview struct {
	Time int64 `json:"time"`
}

type chatEntry struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type snapshot struct {
	Profile profile     `json:"profile"`
	Chats   []chatEntry `json:"chats"`
}
