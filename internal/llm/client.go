package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

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

func (c *Client) GenerateSummary(ctx context.Context, messages []Message) (string, error) {
	var params []openai.ChatCompletionMessageParamUnion
	for _, m := range messages {
		switch m.Role {
		case "system":
			params = append(params, openai.SystemMessage(m.Content))
		case "user":
			params = append(params, openai.UserMessage(m.Content))
		case "assistant":
			params = append(params, openai.AssistantMessage(m.Content))
		}
	}
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       c.model,
		Messages:    params,
		MaxTokens:   openai.Int(256),
		Temperature: openai.Float(0.5),
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	return resp.Choices[0].Message.Content, nil
}
