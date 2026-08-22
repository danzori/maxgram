package max

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/danzori/maxgram/internal/domain/chat"
)

const contactBatchSize = 100

type Directory struct {
	client *Client
	log    *slog.Logger

	selfID atomic.Int64

	mu    sync.RWMutex
	users map[int64]string
	chats map[int64]chat.Chat

	order  []int64
	failed map[int64]struct{}
}

func newDirectory(client *Client, log *slog.Logger) *Directory {
	return &Directory{
		client: client,
		log:    log.With("component", "max.directory"),
		users:  make(map[int64]string),
		chats:  make(map[int64]chat.Chat),
		failed: make(map[int64]struct{}),
	}
}

func (d *Directory) LoadSnapshot(snap snapshot) []int64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	if id := snap.Profile.id(); id != 0 {
		d.selfID.Store(id)

		for _, n := range snap.Profile.displayNames() {
			if display := n.display(); display != "" {
				d.users[id] = display
				break
			}
		}
	} else {
		d.log.Warn("snapshot carries no profile id, keeping the known one", "self_id", d.selfID.Load())
	}

	self := d.selfID.Load()
	unresolved := make(map[int64]struct{})
	d.order = make([]int64, 0, len(snap.Chats))

	for _, entry := range snap.Chats {
		if entry.ID == 0 {
			continue
		}

		d.order = append(d.order, entry.ID)

		c := chat.Chat{
			ID:    entry.ID,
			Title: entry.Title,
			Kind:  chat.Kind(entry.Type),
		}

		if millis := entry.lastActivityMillis(); millis > 0 {
			c.LastActive = time.UnixMilli(millis)
		}

		for raw := range entry.Participants {
			uid, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				continue
			}

			if _, known := d.users[uid]; !known {
				unresolved[uid] = struct{}{}
			}

			if self != 0 && uid != self && c.PeerID == 0 && c.Kind == chat.KindDialog {
				c.PeerID = uid
			}
		}

		d.chats[entry.ID] = c
	}

	return slices.Collect(maps.Keys(unresolved))
}

func (d *Directory) Prefetch(ctx context.Context, ids []int64) {
	pending := make([]int64, 0, len(ids))

	d.mu.RLock()
	for _, id := range ids {
		_, known := d.users[id]
		_, failed := d.failed[id]

		if !known && !failed {
			pending = append(pending, id)
		}
	}
	d.mu.RUnlock()

	for chunk := range slices.Chunk(pending, contactBatchSize) {
		if ctx.Err() != nil {
			return
		}

		d.fetch(ctx, chunk)
	}
}

func (d *Directory) UserName(ctx context.Context, id int64) string {
	if id == 0 {
		return ""
	}

	d.mu.RLock()
	name, known := d.users[id]
	_, failed := d.failed[id]
	d.mu.RUnlock()

	if known {
		return name
	}
	if failed {
		return strconv.FormatInt(id, 10)
	}

	d.fetch(ctx, []int64{id})

	d.mu.RLock()
	name, known = d.users[id]
	d.mu.RUnlock()

	if known {
		return name
	}

	d.mu.Lock()
	d.failed[id] = struct{}{}
	d.mu.Unlock()

	return strconv.FormatInt(id, 10)
}

func (d *Directory) Chat(id int64) (chat.Chat, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	c, ok := d.chats[id]
	if !ok {
		return chat.Chat{}, false
	}

	return d.withResolvedTitle(c), true
}

func (d *Directory) Chats() []chat.Chat {
	d.mu.RLock()
	defer d.mu.RUnlock()

	out := make([]chat.Chat, 0, len(d.order))
	for _, id := range d.order {
		if c, ok := d.chats[id]; ok {
			out = append(out, d.withResolvedTitle(c))
		}
	}

	return out
}

func (d *Directory) SelfID() int64 {
	return d.selfID.Load()
}

func (d *Directory) withResolvedTitle(c chat.Chat) chat.Chat {
	if c.Title != "" || c.PeerID == 0 || c.PeerID == d.selfID.Load() {
		return c
	}

	if name, ok := d.users[c.PeerID]; ok {
		c.Title = name
	}

	return c
}

func (d *Directory) fetch(ctx context.Context, ids []int64) {
	if len(ids) == 0 {
		return
	}

	contacts, err := d.client.contacts(ctx, ids)
	if err != nil {
		d.log.Warn("resolve contacts", "count", len(ids), "err", err)

		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, contact := range contacts {
		id := contact.id()
		display := contact.display()

		for id == 0 || display == "" {
			continue
		}

		d.users[id] = display
	}
}
