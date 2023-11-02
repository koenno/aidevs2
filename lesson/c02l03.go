package lesson

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["c02l03"] = C02L03Creator{}
}

type C02L03Creator struct {
}

func (c C02L03Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C02L03{
		embeddor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "embedding",
	}
}

type Embeddor interface {
	CreateEmbeddings(context.Context, openai.EmbeddingRequestConverter) (openai.EmbeddingResponse, error)
}

type C02L03 struct {
	embeddor  Embeddor
	moderator Moderator
	taskName  string
}

type C02L03Task struct {
	Task
}

func (t *C02L03Task) GetCode() int {
	return t.Code
}

func (t *C02L03Task) GetMsg() string {
	return t.Msg
}

func (t *C02L03Task) SetToken(token string) {
	t.Token = token
}

type C02L03Solution []float32

func (l C02L03) Solve(server TaskServer) error {
	var task C02L03Task
	err := server.FetchTask(l.taskName, &task)
	if err != nil {
		return fmt.Errorf("failed to fetch task: %v", err)
	}
	solution, err := l.getSolution(task)
	if err != nil {
		return fmt.Errorf("failed to find solution: %v", err)
	}
	err = server.SendSolution(task.Token, solution)
	if err != nil {
		return fmt.Errorf("failed to send solution: %v", err)
	}
	return nil
}

func (l C02L03) getSolution(task C02L03Task) (C02L03Solution, error) {
	const phrase = "Send me just array of params: "
	parts := strings.SplitAfter(task.Msg, phrase)
	text := parts[len(parts)-1]
	resp, err := l.moderatedEmbedding(text)
	if err != nil {
		return nil, fmt.Errorf("solution %s failure: %v", l.taskName, err)
	}
	log.Printf("%s | %v", text, resp)
	return resp, nil
}

func (l C02L03) moderatedEmbedding(text string) ([]float32, error) {
	moderationRequired, err := l.moderator.Moderate(context.Background(), text)
	if err != nil {
		return nil, fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return nil, fmt.Errorf("entry breaks openai usage policies: %s", text)
	}
	resp, err := l.embedding(text)
	if err != nil {
		return nil, fmt.Errorf("failed to create moderated embedding: %v", err)
	}
	return resp, nil
}

func (l C02L03) embedding(text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	}
	resp, err := l.embeddor.CreateEmbeddings(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("response failure for embeddings: %v", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response received")
	}
	return resp.Data[0].Embedding, nil
}
