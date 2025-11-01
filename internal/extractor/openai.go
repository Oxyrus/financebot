package extractor

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"github.com/Oxyrus/financebot/internal/expense"
)

// Service defines the contract for turning free-form text into an expense item.
type Service interface {
	Extract(ctx context.Context, text string) (expense.Item, error)
}

// OpenAI implements Service using the OpenAI Chat Completions API.
type OpenAI struct {
	client *openai.Client
	model  string
}

// NewOpenAI returns an extractor configured with the provided OpenAI client.
func NewOpenAI(client *openai.Client) *OpenAI {
	return &OpenAI{
		client: client,
		model:  "gpt-4o-mini",
	}
}

// Extract requests structured expense data from OpenAI and normalizes the result.
func (o *OpenAI) Extract(ctx context.Context, text string) (expense.Item, error) {
	prompt := fmt.Sprintf(`Extract structured data from this expense description:
"%s"

Return a JSON object like this:
{
  "category": "string",
  "amount": number,
  "description": "string"
}`, text)

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You extract structured expense data from text and always respond ONLY with valid JSON."},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return expense.Item{}, err
	}

	if len(resp.Choices) == 0 {
		return expense.Item{}, fmt.Errorf("no choices returned from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	var item expense.Item
	if err := json.Unmarshal([]byte(content), &item); err != nil {
		return expense.Item{}, fmt.Errorf("failed to parse GPT response: %v\nResponse: %s", err, content)
	}

	return item, nil
}
