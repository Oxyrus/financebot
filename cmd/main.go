package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"

	"github.com/Oxyrus/financebot/internal/bot"
	"github.com/Oxyrus/financebot/internal/config"
	"github.com/Oxyrus/financebot/internal/extractor"
	"github.com/Oxyrus/financebot/internal/storage/memory"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal(err)
	}

	botAPI.Debug = false
	log.Printf("authorized on account %s", botAPI.Self.UserName)

	openaiClient := openai.NewClient(cfg.OpenAIKey)
	extractorSvc := extractor.NewOpenAI(openaiClient)
	store := memory.NewStore()

	expenseBot := bot.New(botAPI, cfg, extractorSvc, store)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := expenseBot.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped: %v", err)
	}
}
