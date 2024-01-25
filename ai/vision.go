package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Vision struct {
	client    *openai.Client
	moderator *Moderator
	model     string
}

func NewVisioner(openaiKey string) *Vision {
	client := openai.NewClient(openaiKey)
	return &Vision{
		client: client,
		moderator: &Moderator{
			client: client,
		},
		model: openai.GPT4VisionPreview,
	}
}

func (v *Vision) ModeratedSee(system, user, assistant, imageURI string) (string, error) {
	moderationRequired, err := v.moderator.Moderate(context.Background(), system)
	if err != nil {
		return "", fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return "", fmt.Errorf("entry breaks openai usage policies: %s", system)
	}
	resp, err := v.See(system, user, assistant, imageURI)
	if err != nil {
		return "", fmt.Errorf("failed to complete moderated see: %v", err)
	}
	return resp, nil
}

func (v *Vision) See(system, user, assistant, imageURI string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: v.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: user,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    imageURI,
							Detail: openai.ImageURLDetailAuto,
						},
					},
				},
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
	resp, err := v.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("response failure for seeing: %v", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response received")
	}
	return resp.Choices[0].Message.Content, nil
}
