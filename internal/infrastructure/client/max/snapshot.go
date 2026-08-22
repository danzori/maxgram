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
	Names   []name `json:"names"`
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

func (p profile) displayNames() []name {
	if len(p.Names) > 0 {
		return p.Names
	}

	if p.Contact != nil {
		return p.Contact.Names
	}

	return nil
}

type lastMessagePreview struct {
	Time int64 `json:"time"`
}

type chatEntry struct {
	ID            int64               `json:"id"`
	Type          string              `json:"type"`
	Title         string              `json:"title"`
	Participants  map[string]int64    `json:"participants"`
	LastEventTime int64               `json:"lastEventTime"`
	Modified      int64               `json:"modified"`
	Created       int64               `json:"created"`
	LastMessage   *lastMessagePreview `json:"lastMessage"`
}

func (c chatEntry) lastActivityMillis() int64 {
	best := c.LastEventTime

	for _, v := range []int64{c.Modified, c.Created} {
		if v > best {
			best = v
		}
	}

	if c.LastMessage != nil && c.LastMessage.Time > best {
		best = c.LastMessage.Time
	}

	return best
}

type snapshot struct {
	Profile profile     `json:"profile"`
	Chats   []chatEntry `json:"chats"`
}

type contactEntry struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Names     []name `json:"names"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Name      string `json:"name"`
}

func (c contactEntry) id() int64 {
	if c.ID != 0 {
		return c.ID
	}

	return c.UserID
}

func (c contactEntry) display() string {
	for _, n := range c.Names {
		if d := n.display(); d != "" {
			return d
		}
	}

	return name{
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Name:      c.Name,
	}.display()
}

type contactsResponse struct {
	Contacts []contactEntry `json:"contacts"`
	Users    []contactEntry `json:"users"`
}
