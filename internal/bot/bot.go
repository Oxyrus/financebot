package bot

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Oxyrus/financebot/internal/extractor"
	"github.com/Oxyrus/financebot/internal/storage"
)

// Authorizer determines whether a Telegram username may interact with the bot.
type Authorizer interface {
	IsUserAllowed(username string) bool
}

// TelegramAPI abstracts sending and receiving Telegram updates.
type TelegramAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	StopReceivingUpdates()
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// Bot wraps Telegram update handling with expense extraction and persistence.
type Bot struct {
	api        TelegramAPI
	extractor  extractor.Service
	store      storage.ExpenseStore
	authorizer Authorizer
}

// New constructs a bot ready to process updates.
func New(api TelegramAPI, authorizer Authorizer, extractor extractor.Service, store storage.ExpenseStore) *Bot {
	return &Bot{
		api:        api,
		extractor:  extractor,
		store:      store,
		authorizer: authorizer,
	}
}

// Start begins consuming telegram updates until the context is canceled.
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	defer b.api.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if update.Message.From == nil {
		log.Println("skipping message without sender")
		return
	}

	username := update.Message.From.UserName
	if username == "" || !b.authorizer.IsUserAllowed(username) {
		return
	}

	text := update.Message.Text
	log.Printf("[%s] %s", username, text)

	item, err := b.extractor.Extract(ctx, text)
	if err != nil {
		b.reply(update.Message.Chat.ID, fmt.Sprintf("Error: %v", err))
		return
	}

	if err := b.store.SaveExpense(ctx, item); err != nil {
		b.reply(update.Message.Chat.ID, fmt.Sprintf("Failed to store expense: %v", err))
		return
	}

	b.reply(update.Message.Chat.ID, item.ReplyMessage())
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("failed to send message: %v", err)
	}
}
