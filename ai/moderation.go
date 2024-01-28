package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Moderator struct {
	client *openai.Client
}

func NewModerator(client *openai.Client) *Moderator {
	return &Moderator{
		client: client,
	}
}

// Moderate returns true if moderation is required and the given entry does not fullfil openai usage policy
func (m *Moderator) Moderate(ctx context.Context, entry string) (bool, error) {
	req := openai.ModerationRequest{
		Input: entry,
		Model: openai.ModerationTextLatest,
	}
	resp, err := m.client.Moderations(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to call moderation openai entry: %v", err)
	}
	if len(resp.Results) == 0 {
		return false, fmt.Errorf("no results in moderation response")
	}
	return resp.Results[0].Flagged, nil
}
