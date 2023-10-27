package moderation

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIModerator interface {
	Moderations(ctx context.Context, request openai.ModerationRequest) (openai.ModerationResponse, error)
}

type Moderator struct {
	OpenAIMod OpenAIModerator
}

// Moderate returns true is moderation is required and the given entry does not fullfil openai usage policy
func (m Moderator) Moderate(ctx context.Context, entry string) (bool, error) {
	req := openai.ModerationRequest{
		Input: entry,
		Model: openai.ModerationTextLatest,
	}
	resp, err := m.OpenAIMod.Moderations(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to call moderation openai entry: %v", err)
	}
	if len(resp.Results) == 0 {
		return false, fmt.Errorf("no results in moderation response")
	}
	return resp.Results[0].Flagged, nil
}
