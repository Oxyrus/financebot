package extractor

import (
	"context"
	"errors"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

type stubClient struct {
	response openai.ChatCompletionResponse
	err      error
	request  openai.ChatCompletionRequest
}

func (s *stubClient) CreateChatCompletion(_ context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	s.request = request
	if s.err != nil {
		return openai.ChatCompletionResponse{}, s.err
	}
	return s.response, nil
}

func TestOpenAIExtractSuccess(t *testing.T) {
	client := &stubClient{
		response: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: `{"category":"Food","amount":12.5,"description":"Lunch burrito"}`,
					},
				},
			},
		},
	}

	extractor := &OpenAI{client: client, model: "test-model"}

	item, err := extractor.Extract(context.Background(), "Bought lunch for $12.50")
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	if item.Category != "Food" || item.Amount != 12.5 || item.Description != "Lunch burrito" {
		t.Fatalf("unexpected item %#v", item)
	}

	if client.request.Model != "test-model" {
		t.Fatalf("expected model %q, got %q", "test-model", client.request.Model)
	}
}

func TestOpenAIExtractPropagatesErrors(t *testing.T) {
	client := &stubClient{err: errors.New("openai error")}
	extractor := &OpenAI{client: client, model: "test-model"}

	_, err := extractor.Extract(context.Background(), "some text")
	if err == nil || err.Error() != "openai error" {
		t.Fatalf("expected openai error, got %v", err)
	}
}

func TestOpenAIExtractNoChoices(t *testing.T) {
	client := &stubClient{
		response: openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{}},
	}
	extractor := &OpenAI{client: client, model: "test-model"}

	if _, err := extractor.Extract(context.Background(), "text"); err == nil {
		t.Fatal("expected error when no choices returned")
	}
}

func TestOpenAIExtractInvalidJSON(t *testing.T) {
	client := &stubClient{
		response: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: `{"category":`,
					},
				},
			},
		},
	}
	extractor := &OpenAI{client: client, model: "test-model"}

	if _, err := extractor.Extract(context.Background(), "text"); err == nil {
		t.Fatal("expected JSON parsing error")
	}
}
