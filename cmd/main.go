package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type Expense struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading environment variables directly")
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if telegramToken == "" || openaiKey == "" {
		log.Fatal("TELEGRAM_TOKEN or OPENAI_API_KEY not set")
	}

	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false
	log.Printf("authorized on account %s", bot.Self.UserName)

	openaiClient := openai.NewClient(openaiKey)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.From.UserName != "iamoxyrus" {
			continue
		}

		text := update.Message.Text
		log.Printf("[%s] %s", update.Message.From.UserName, text)

		expense, err := extractExpense(openaiClient, text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error: %v", err)))
			continue
		}

		reply := fmt.Sprintf("Recorded\nDescription: %s\nCategory %s\nAmount: $%.2f",
			expense.Description, expense.Category, expense.Amount)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, reply))
	}
}

func extractExpense(client *openai.Client, text string) (*Expense, error) {
	prompt := fmt.Sprintf(`Extract structured data from this expense description:
		"%s"

		Return a JSON object like this:
		{
			"category": "string",
			"amount": number,
			"description": "string"
		}`, text)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You extract structured expense data from text and always respond ONLY with valid JSON."},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	var exp Expense
	if err := json.Unmarshal([]byte(content), &exp); err != nil {
		return nil, fmt.Errorf("failed to parse GPT response: %v\nResponse: %s", err, content)
	}
	return &exp, nil
}
