package realtime

import (
	"testing"
	"time"

	"example/hello/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func newChatID(t *testing.T) pgtype.UUID {
	t.Helper()
	return pgtype.UUID{Bytes: uuid.New(), Valid: true}
}

func recvOrTimeout(t *testing.T, ch <-chan database.ChatMessage) (database.ChatMessage, bool) {
	t.Helper()
	select {
	case msg, ok := <-ch:
		return msg, ok
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
		return database.ChatMessage{}, false
	}
}

func TestChatBroker_DeliversToSubscriber(t *testing.T) {
	broker := NewChatBroker()
	chatID := newChatID(t)

	ch, unsubscribe := broker.Subscribe(chatID)
	defer unsubscribe()

	want := database.ChatMessage{ID: newChatID(t), ChatID: chatID, Content: "hello"}
	broker.Publish(want)

	got, ok := recvOrTimeout(t, ch)
	if !ok {
		t.Fatal("channel closed unexpectedly")
	}
	if got.Content != want.Content {
		t.Fatalf("got content %q, want %q", got.Content, want.Content)
	}
}

func TestChatBroker_MultipleSubscribersBothReceive(t *testing.T) {
	broker := NewChatBroker()
	chatID := newChatID(t)

	chA, unsubA := broker.Subscribe(chatID)
	defer unsubA()
	chB, unsubB := broker.Subscribe(chatID)
	defer unsubB()

	broker.Publish(database.ChatMessage{ChatID: chatID, Content: "fan-out"})

	if got, ok := recvOrTimeout(t, chA); !ok || got.Content != "fan-out" {
		t.Fatalf("subscriber A: got %+v, ok=%v", got, ok)
	}
	if got, ok := recvOrTimeout(t, chB); !ok || got.Content != "fan-out" {
		t.Fatalf("subscriber B: got %+v, ok=%v", got, ok)
	}
}

func TestChatBroker_DoesNotLeakAcrossChats(t *testing.T) {
	broker := NewChatBroker()
	chatA := newChatID(t)
	chatB := newChatID(t)

	chForA, unsubA := broker.Subscribe(chatA)
	defer unsubA()

	broker.Publish(database.ChatMessage{ChatID: chatB, Content: "not for A"})

	select {
	case msg := <-chForA:
		t.Fatalf("subscriber for chat A received a message meant for chat B: %+v", msg)
	case <-time.After(100 * time.Millisecond):
		// expected: nothing arrives
	}
}

func TestChatBroker_UnsubscribeClosesChannel(t *testing.T) {
	broker := NewChatBroker()
	chatID := newChatID(t)

	ch, unsubscribe := broker.Subscribe(chatID)
	unsubscribe()

	if _, ok := <-ch; ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}

	// Publishing after every subscriber has left must not panic (no
	// subscribers left registered for this chat).
	broker.Publish(database.ChatMessage{ChatID: chatID, Content: "after unsubscribe"})
}

func TestChatBroker_SlowSubscriberDoesNotBlockPublish(t *testing.T) {
	broker := NewChatBroker()
	chatID := newChatID(t)

	// Never read from this channel - simulates a stalled/slow client.
	_, unsubscribe := broker.Subscribe(chatID)
	defer unsubscribe()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 20; i++ {
			broker.Publish(database.ChatMessage{ChatID: chatID, Content: "spam"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on a slow subscriber instead of dropping the message")
	}
}
