package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(baseURL, model string) *Client {
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey("not-needed"),
	)
	return &Client{
		client: &client,
		model:  model,
	}
}

func (c *Client) GenerateSummary(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		MaxTokens: openai.Int(128),
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	return resp.Choices[0].Message.Content, nil
}
