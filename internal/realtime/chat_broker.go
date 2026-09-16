// Package realtime holds in-process pub/sub brokers that bridge a mutation
// (a message being written) to any active GraphQL subscriptions watching
// for it, over the WebSocket transport. See BACKEND_HANDOFF.md's "Real-time
// chat" section for the full design and its single-instance limitation.
package realtime

import (
	"log"
	"sync"

	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

// ChatBroker fans newly created chat messages out to every active
// subscription for that chat. It is in-memory only: subscribers must be
// connected to this same process to receive anything. That is fine for a
// single API instance; it stops working the moment the API runs as more
// than one replica behind a load balancer, because a message published on
// instance A never reaches a subscriber whose WebSocket landed on instance
// B. Swapping this for a Redis-backed (or similar) broker is the fix if
// that's ever needed - see BACKEND_HANDOFF.md.
type ChatBroker struct {
	mu   sync.RWMutex
	subs map[string]map[chan database.ChatMessage]struct{}
}

// NewChatBroker creates an empty broker. One instance is shared by the
// whole process (wired in internal/app.New, held by ChatService).
func NewChatBroker() *ChatBroker {
	return &ChatBroker{
		subs: make(map[string]map[chan database.ChatMessage]struct{}),
	}
}

// Subscribe registers a listener for a single chat's messages. The
// returned channel receives every message Published for that chat from
// this point on; call the returned unsubscribe func (always via defer)
// when the caller is done listening, typically when the subscription's
// context is cancelled (client disconnected or unsubscribed).
func (b *ChatBroker) Subscribe(chatID pgtype.UUID) (<-chan database.ChatMessage, func()) {
	key := chatID.String()
	ch := make(chan database.ChatMessage, 8)

	b.mu.Lock()
	if b.subs[key] == nil {
		b.subs[key] = make(map[chan database.ChatMessage]struct{})
	}
	b.subs[key][ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs[key], ch)
			if len(b.subs[key]) == 0 {
				delete(b.subs, key)
			}
			b.mu.Unlock()
			close(ch)
		})
	}

	return ch, unsubscribe
}

// Publish fans a newly created message out to every active subscriber for
// its chat. Non-blocking: a subscriber whose buffered channel is full
// (i.e. a slow/stalled client) has this message dropped for it rather
// than blocking the publisher - and every other subscriber - on one slow
// reader. The publisher call site (ChatService) never waits on this.
func (b *ChatBroker) Publish(message database.ChatMessage) {
	key := message.ChatID.String()

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subs[key] {
		select {
		case ch <- message:
		default:
			log.Printf(
				"realtime: dropped chat message %s for a slow subscriber on chat %s",
				message.ID.String(),
				key,
			)
		}
	}
}
