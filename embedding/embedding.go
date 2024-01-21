package embedding

import (
	"context"
	"fmt"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/sashabaranov/go-openai"
)

type AIClient interface {
	CreateEmbeddings(context.Context, openai.EmbeddingRequestConverter) (openai.EmbeddingResponse, error)
}

type Moderator interface {
	Moderate(ctx context.Context, entry string) (bool, error)
}

type Embeddor struct {
	Client    AIClient
	Moderator Moderator
}

func (e Embeddor) Embedding(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	}
	resp, err := e.Client.CreateEmbeddings(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("response failure for embeddings: %v", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response received")
	}
	return resp.Data[0].Embedding, nil
}

func (e Embeddor) ModeratedEmbedding(ctx context.Context, text string) ([]float32, error) {
	invalid, err := e.Moderator.Moderate(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to moderate embedding: %v", err)
	}
	if invalid {
		return nil, fmt.Errorf("text does not fullfil usage policy")
	}

	var embedding []float32
	r := retrier.New(retrier.ExponentialBackoff(10, 100*time.Millisecond), nil)
	err = r.Run(func() error {
		embedding, err = e.Embedding(ctx, text)
		return err
	})
	return embedding, err
}
