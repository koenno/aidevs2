package lesson

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/koenno/aidevs2/embedding"
	"github.com/koenno/aidevs2/moderation"
	"github.com/koenno/aidevs2/vectordb"
	"github.com/sashabaranov/go-openai"
)

const collectionName = "aidevs2_c03l04"

func init() {
	registry["c03l04"] = C03L04Creator{}
}

type C03L04Creator struct {
}

func (c C03L04Creator) Create(openaiKey string) TaskSolver {
	db, err := vectordb.New("localhost:6334")
	if err != nil {
		log.Fatalf("failed to create vector db: %v", err)
	}
	client := openai.NewClient(openaiKey)
	return C03L04{
		embeddor: embedding.Embeddor{
			Client: client,
			Moderator: moderation.Moderator{
				OpenAIMod: client,
			},
		},
		db:       db,
		taskName: "search",
	}
}

type ModeratedEmbeddor interface {
	ModeratedEmbedding(ctx context.Context, text string) ([]float32, error)
}

type VectorDB interface {
	CollectionExist(ctx context.Context, collectionName string) (bool, error)
	Search(ctx context.Context, collectionName string, vector []float32, items any, options ...vectordb.SearchOption) error
	UpsertMany(ctx context.Context, collectionName string, items []any) error
	CreateCollection(ctx context.Context, collectionName string) error
}

type C03L04 struct {
	embeddor ModeratedEmbeddor
	db       VectorDB
	taskName string
}

type C03L04Task struct {
	Task
	Question string `json:"question"`
}

func (t *C03L04Task) GetCode() int {
	return t.Code
}

func (t *C03L04Task) GetMsg() string {
	return t.Msg
}

func (t *C03L04Task) SetToken(token string) {
	t.Token = token
}

type C03L04Solution string

func (l C03L04) Solve(server TaskServer) error {
	var task C03L04Task
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

type ArchiveEntry struct {
	Title string `json:"title" qdrant:"title"`
	URL   string `json:"url" qdrant:"url"`
	Info  string `json:"info" qdrant:"info"`
	Date  string `json:"date" qdrant:"date"`
}

type ArchiveEntity struct {
	ID     string    `qdrant:"_id"`
	Vector []float32 `qdrant:"_vector"`
	Title  string    `qdrant:"title"`
	URL    string    `qdrant:"url"`
	Info   string    `qdrant:"info"`
	Date   string    `qdrant:"date"`
}

func (l C03L04) getSolution(task C03L04Task) (C03L04Solution, error) {
	const filePath = "data/c03l04/small_archiwum1.json"
	// const filePath = "data/c03l04/test.json"
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s': %v", filePath, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file '%s': %v\n", filePath, err)
		}
	}()

	ctx := context.Background()
	exist, err := l.db.CollectionExist(ctx, collectionName)
	if err != nil {
		return "", fmt.Errorf("failed to check collection presence: %v", err)
	}
	if !exist {
		if err := l.db.CreateCollection(ctx, collectionName); err != nil {
			return "", fmt.Errorf("failed to create collection: %v", err)
		}
		log.Printf("collection '%s' created", collectionName)
		var archive []ArchiveEntry
		if err := json.NewDecoder(f).Decode(&archive); err != nil {
			return "", fmt.Errorf("failed to decode file content '%s': %v", filePath, err)
		}
		if len(archive) == 0 {
			return "", fmt.Errorf("no archive entries found")
		}
		if err := l.storeEntries(ctx, archive); err != nil {
			return "", fmt.Errorf("failed to store entries: %v", err)
		}
		log.Printf("all entries stored")
	}

	answer, err := l.findAnswer(ctx, task.Question)
	if err != nil {
		return "", fmt.Errorf("failed to find answer for question '%s': %v", task.Question, err)
	}

	log.Printf("Question: %s", task.Question)
	log.Printf("Answer: %s", answer)

	return C03L04Solution(answer), nil
}

func (l C03L04) storeEntries(ctx context.Context, entries []ArchiveEntry) error {
	var entities []any
	for _, entry := range entries {
		embedding, err := l.embeddor.ModeratedEmbedding(ctx, entry.Title)
		if err != nil {
			return fmt.Errorf("failed to create embedding for entry '%s': %v", entry.Title, err)
		}
		if len(embedding) == 0 {
			return fmt.Errorf("no embedding for entry '%s'", entry.Title)
		}
		entity := ArchiveEntity{
			ID:     uuid.NewString(),
			Vector: embedding,
			Title:  entry.Title,
			URL:    entry.URL,
			Info:   entry.Info,
			Date:   entry.Date,
		}
		entities = append(entities, entity)
	}
	log.Printf("embeddings created")
	err := l.db.UpsertMany(context.Background(), collectionName, entities)
	if err != nil {
		return fmt.Errorf("failed to upsert archive entity: %v", err)
	}
	return nil
}

func (l C03L04) findAnswer(ctx context.Context, question string) (string, error) {
	log.Printf("finding answer")
	embedding, err := l.embeddor.ModeratedEmbedding(ctx, question)
	if err != nil {
		return "", fmt.Errorf("failed to create embedding for question '%s': %v", question, err)
	}
	if len(embedding) == 0 {
		return "", fmt.Errorf("no embedding for question '%s'", question)
	}
	var entities []ArchiveEntity
	err = l.db.Search(ctx, collectionName, embedding, &entities, vectordb.WithLimit(1))
	if err != nil {
		return "", fmt.Errorf("failed to find answer: %v", err)
	}
	if len(entities) == 0 {
		return "", fmt.Errorf("no answer found %#v", entities)
	}
	return entities[0].URL, nil
}
