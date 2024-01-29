package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Chat struct {
	client    *openai.Client
	moderator *Moderator
	model     string
}

type Option func(*Chat)

func WithModel(m string) Option {
	return func(c *Chat) {
		c.model = m
	}
}

func NewChat(openaiKey string, opts ...Option) *Chat {
	client := openai.NewClient(openaiKey)
	chat := &Chat{
		client: client,
		moderator: &Moderator{
			client: client,
		},
		model: openai.GPT3Dot5Turbo,
	}
	for _, o := range opts {
		o(chat)
	}
	return chat
}

func (c *Chat) ModeratedChat(system string, userMsgs ...string) (string, error) {
	moderationRequired, err := c.moderator.Moderate(context.Background(), system)
	if err != nil {
		return "", fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return "", fmt.Errorf("entry breaks openai usage policies: %s", system)
	}
	resp, err := c.CompleteChat(system, userMsgs...)
	if err != nil {
		return "", fmt.Errorf("failed to complete moderated chat: %v", err)
	}
	return resp, nil
}

func (c *Chat) ModeratedFunctionCalling(system, user, assistant string, funcDefs []openai.FunctionDefinition) (*openai.FunctionCall, error) {
	msg := system + user + assistant
	moderationRequired, err := c.moderator.Moderate(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return nil, fmt.Errorf("entry breaks openai usage policies: %s", system)
	}
	resp, err := c.FunctionCalling(system, user, assistant, funcDefs)
	if err != nil {
		return nil, fmt.Errorf("failed to execute moderated function calling: %v", err)
	}
	return resp, nil
}

func (c *Chat) CompleteChat(system string, userMsgs ...string) (string, error) {
	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		},
	}
	for _, userMsg := range userMsgs {
		msg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userMsg,
		}
		msgs = append(msgs, msg)
	}
	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    msgs,
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

func (c *Chat) FunctionCalling(system, user, assistant string, funcDefs []openai.FunctionDefinition) (*openai.FunctionCall, error) {
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
		Functions:   funcDefs,
		MaxTokens:   250,
		Temperature: 0,
		TopP:        1,
		N:           1,
		Stream:      false,
	}
	resp, err := c.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("response failure for chat completion: %v", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response received")
	}
	if resp.Choices[0].Message.FunctionCall == nil {
		return nil, fmt.Errorf("failed to call function: %#v", resp)
	}
	return resp.Choices[0].Message.FunctionCall, nil
}
