package bot

import (
	"context"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Oxyrus/financebot/internal/expense"
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
	items []expense.Item
	err   error
}

func (f *fakeStore) SaveExpense(_ context.Context, item expense.Item) error {
	if f.err != nil {
		return f.err
	}
	f.items = append(f.items, item)
	return nil
}

func (f *fakeStore) Close() error { return nil }

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
