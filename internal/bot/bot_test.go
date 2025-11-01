package bot

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Oxyrus/financebot/internal/expense"
	"github.com/Oxyrus/financebot/internal/storage"
)

type fakeExtractor struct {
	item     expense.Item
	err      error
	requests []string
}

func (f *fakeExtractor) Extract(_ context.Context, text string) (expense.Item, error) {
	f.requests = append(f.requests, text)
	if f.err != nil {
		return expense.Item{}, f.err
	}
	return f.item, nil
}

type fakeStore struct {
	items    []expense.Item
	err      error
	stats    storage.Summary
	statsErr error
}

func (f *fakeStore) SaveExpense(_ context.Context, item expense.Item) error {
	if f.err != nil {
		return f.err
	}
	f.items = append(f.items, item)
	return nil
}

func (f *fakeStore) Close() error { return nil }

func (f *fakeStore) Stats(_ context.Context, _ time.Time) (storage.Summary, error) {
	if f.statsErr != nil {
		return storage.Summary{}, f.statsErr
	}
	return f.stats, nil
}

type fakeAPI struct {
	messages []string
}

func (f *fakeAPI) GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel { return nil }

func (f *fakeAPI) StopReceivingUpdates() {}

func (f *fakeAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	msg, ok := c.(tgbotapi.MessageConfig)
	if !ok {
		return tgbotapi.Message{}, errors.New("unexpected chattable type")
	}
	f.messages = append(f.messages, msg.Text)
	return tgbotapi.Message{}, nil
}

type allowAllAuthorizer struct{}

func (allowAllAuthorizer) IsUserAllowed(string) bool { return true }

type denyAuthorizer struct{}

func (denyAuthorizer) IsUserAllowed(string) bool { return false }

func TestHandleUpdateSuccess(t *testing.T) {
	api := &fakeAPI{}
	extract := &fakeExtractor{
		item: expense.Item{Category: "Food", Amount: 12.34, Description: "Lunch"},
	}
	store := &fakeStore{}

	b := New(api, allowAllAuthorizer{}, extract, store)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "iamoxyrus"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "Bought lunch for $12.34",
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(extract.requests) != 1 {
		t.Fatalf("expected extractor to be called once, got %d", len(extract.requests))
	}
	if len(store.items) != 1 {
		t.Fatalf("expected store to persist one item, got %d", len(store.items))
	}
	if len(api.messages) != 1 {
		t.Fatalf("expected bot to send one message, got %d", len(api.messages))
	}
	if want := store.items[0]; want.Category != "Food" || want.Amount != 12.34 {
		t.Fatalf("unexpected stored item %#v", want)
	}
}

func TestHandleUpdateExtractorError(t *testing.T) {
	api := &fakeAPI{}
	extract := &fakeExtractor{err: errors.New("extract failed")}
	store := &fakeStore{}
	b := New(api, allowAllAuthorizer{}, extract, store)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "iamoxyrus"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "Lunch",
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(store.items) != 0 {
		t.Fatalf("expected no items stored, got %d", len(store.items))
	}
	if len(api.messages) != 1 || api.messages[0] != "Error: extract failed" {
		t.Fatalf("unexpected messages %#v", api.messages)
	}
}

func TestHandleUpdateStoreError(t *testing.T) {
	api := &fakeAPI{}
	extract := &fakeExtractor{
		item: expense.Item{Category: "Travel", Amount: 42, Description: "Taxi"},
	}
	store := &fakeStore{err: errors.New("db error")}
	b := New(api, allowAllAuthorizer{}, extract, store)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "iamoxyrus"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "Taxi",
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(store.items) != 0 {
		t.Fatalf("expected no items stored, got %d", len(store.items))
	}
	if len(api.messages) != 1 || api.messages[0] != "Failed to store expense: db error" {
		t.Fatalf("unexpected messages %#v", api.messages)
	}
}

func TestHandleUpdateUnauthorizedUser(t *testing.T) {
	api := &fakeAPI{}
	extract := &fakeExtractor{}
	store := &fakeStore{}
	b := New(api, denyAuthorizer{}, extract, store)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "intruder"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "Sneaky expense",
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(extract.requests) != 0 {
		t.Fatalf("expected extractor not to be called, got %d", len(extract.requests))
	}
	if len(store.items) != 0 {
		t.Fatalf("expected no items stored, got %d", len(store.items))
	}
	if len(api.messages) != 0 {
		t.Fatalf("expected no messages sent, got %d", len(api.messages))
	}
}

func TestHandleCommandAddWithArgs(t *testing.T) {
	api := &fakeAPI{}
	extract := &fakeExtractor{item: expense.Item{Category: "Coffee", Amount: 3.5, Description: "Morning brew"}}
	store := &fakeStore{}
	b := New(api, allowAllAuthorizer{}, extract, store)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "iamoxyrus"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "/add Morning coffee $3.50",
			Entities: []tgbotapi.MessageEntity{
				{Offset: 0, Length: 4, Type: "bot_command"},
			},
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(store.items) != 1 {
		t.Fatalf("expected store to save item from /add command, got %d", len(store.items))
	}
	if len(api.messages) != 1 {
		t.Fatalf("expected confirmation message, got %d", len(api.messages))
	}
}

func TestHandleCommandAddWithoutArgs(t *testing.T) {
	api := &fakeAPI{}
	b := New(api, allowAllAuthorizer{}, &fakeExtractor{}, &fakeStore{})

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "iamoxyrus"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "/add",
			Entities: []tgbotapi.MessageEntity{
				{Offset: 0, Length: 4, Type: "bot_command"},
			},
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(api.messages) != 1 || api.messages[0] == "" {
		t.Fatalf("expected guidance message for missing args, got %#v", api.messages)
	}
}

func TestHandleCommandStats(t *testing.T) {
	api := &fakeAPI{}
	b := New(api, allowAllAuthorizer{}, &fakeExtractor{}, &fakeStore{
		stats: storage.Summary{
			TotalCount:  2,
			TotalAmount: 25.5,
			CategoryTotals: map[string]float64{
				"Travel": 15.5,
				"Food":   10,
			},
		},
	})

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{UserName: "iamoxyrus"},
			Chat: &tgbotapi.Chat{ID: 1},
			Text: "/stats",
			Entities: []tgbotapi.MessageEntity{
				{Offset: 0, Length: 6, Type: "bot_command"},
			},
		},
	}

	b.handleUpdate(context.Background(), update)

	if len(api.messages) != 1 {
		t.Fatalf("expected stats message, got %d", len(api.messages))
	}
	if !strings.Contains(api.messages[0], "Total: $25.50") {
		t.Fatalf("expected total in stats message, got %q", api.messages[0])
	}
	if !strings.Contains(api.messages[0], "Travel: $15.50") {
		t.Fatalf("expected category breakdown, got %q", api.messages[0])
	}
}
