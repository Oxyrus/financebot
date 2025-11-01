package bot

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

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

	if update.Message.IsCommand() {
		b.handleCommand(ctx, update)
		return
	}

	b.processExpense(ctx, update)
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("failed to send message: %v", err)
	}
}

func (b *Bot) handleCommand(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message
	switch msg.Command() {
	case "add":
		args := msg.CommandArguments()
		if args == "" {
			b.reply(msg.Chat.ID, "Send an expense description after /add, e.g. `/add Coffee $3.50`.")
			return
		}
		update.Message.Text = args
		b.processExpense(ctx, update)
	case "stats":
		since := time.Now().AddDate(0, 0, -7)
		summary, err := b.store.Stats(ctx, since)
		if err != nil {
			b.reply(msg.Chat.ID, fmt.Sprintf("Failed to load stats: %v", err))
			return
		}
		if summary.TotalCount == 0 {
			b.reply(msg.Chat.ID, "No expenses recorded in the last 7 days.")
			return
		}
		b.reply(msg.Chat.ID, formatSummary(summary, since))
	default:
		b.reply(msg.Chat.ID, fmt.Sprintf("Unknown command: /%s", msg.Command()))
	}
}

func (b *Bot) processExpense(ctx context.Context, update tgbotapi.Update) {
	text := update.Message.Text
	log.Printf("[%s] %s", update.Message.From.UserName, text)

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

func formatSummary(summary storage.Summary, since time.Time) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Last 7 days (since %s):\n", since.Format("2006-01-02")))
	builder.WriteString(fmt.Sprintf("Total: $%.2f across %d expenses\n", summary.TotalAmount, summary.TotalCount))

	type catTotal struct {
		name  string
		value float64
	}
	catTotals := make([]catTotal, 0, len(summary.CategoryTotals))
	for name, total := range summary.CategoryTotals {
		catTotals = append(catTotals, catTotal{name: name, value: total})
	}
	sort.Slice(catTotals, func(i, j int) bool {
		return catTotals[i].value > catTotals[j].value
	})

	if len(catTotals) > 0 {
		builder.WriteString("By category:\n")
		for _, ct := range catTotals {
			builder.WriteString(fmt.Sprintf("- %s: $%.2f\n", ct.name, ct.value))
		}
	}

	return strings.TrimRight(builder.String(), "\n")
}
