package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Chat struct {
	client *openai.Client
	model  string
}

type Option func(*Chat)

func WithModel(m string) Option {
	return func(c *Chat) {
		c.model = m
	}
}

func NewChat(openaiKey string, opts ...Option) *Chat {
	chat := &Chat{
		client: openai.NewClient(openaiKey),
		model:  openai.GPT3Dot5Turbo,
	}
	for _, o := range opts {
		o(chat)
	}
	return chat
}

func (c *Chat) ModeratedChat(system, user, assistant string) (string, error) {
	moderationRequired, err := c.Moderate(context.Background(), system)
	if err != nil {
		return "", fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return "", fmt.Errorf("entry breaks openai usage policies: %s", system)
	}
	resp, err := c.CompleteChat(system, user, assistant)
	if err != nil {
		return "", fmt.Errorf("failed to complete moderated chat: %v", err)
	}
	return resp, nil
}

// Moderate returns true if moderation is required and the given entry does not fullfil openai usage policy
func (c *Chat) Moderate(ctx context.Context, entry string) (bool, error) {
	req := openai.ModerationRequest{
		Input: entry,
		Model: openai.ModerationTextLatest,
	}
	resp, err := c.client.Moderations(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to call moderation openai entry: %v", err)
	}
	if len(resp.Results) == 0 {
		return false, fmt.Errorf("no results in moderation response")
	}
	return resp.Results[0].Flagged, nil
}

func (c *Chat) CompleteChat(system, user, assistant string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: user,
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: assistant,
			},
		},
		MaxTokens:   250,
		Temperature: 0,
		TopP:        1,
		N:           1,
		Stream:      false,
	}
	resp, err := c.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("response failure for chat completion: %v", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response received")
	}
	return resp.Choices[0].Message.Content, nil
}
