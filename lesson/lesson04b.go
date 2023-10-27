package lesson

import (
	"context"
	"fmt"
	"log"

	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["4b"] = Lesson04bCreator{}
}

type ChatCompletor interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (response openai.ChatCompletionResponse, err error)
}

type Lesson04bCreator struct {
}

func (c Lesson04bCreator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return Lesson04b{
		completor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "blogger",
	}
}

type Lesson04b struct {
	completor ChatCompletor
	moderator Moderator
	taskName  string
}

type Lesson04bTask struct {
	Task
	Blog []string `json:"blog"`
}

func (t *Lesson04bTask) GetCode() int {
	return t.Code
}

func (t *Lesson04bTask) GetMsg() string {
	return t.Msg
}

func (t *Lesson04bTask) SetToken(token string) {
	t.Token = token
}

type Lesson04bSolution []string

func (l Lesson04b) Solve(server TaskServer) error {
	var task Lesson04bTask
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

func (l Lesson04b) getSolution(task Lesson04bTask) (Lesson04bSolution, error) {
	const system = `As a cuisine blogger I want to create a blog post in polish about pizza Margarita.
The blog post is divided on chapters. The chapter must describe only one topic which is`
	solution := make(Lesson04bSolution, len(task.Blog))
	for i, user := range task.Blog {
		entry := system + user
		moderationRequired, err := l.moderator.Moderate(context.Background(), entry)
		if err != nil {
			return nil, fmt.Errorf("failed to moderate entry: %v", err)
		}
		if moderationRequired {
			return nil, fmt.Errorf("entry breaks openai usage policies: %s", entry)
		}
		resp, err := l.completeChat(system, user, "")
		if err != nil {
			return nil, fmt.Errorf("failed to complete chat: %v", err)
		}
		solution[i] = resp
		log.Printf("%s | %s", user, solution[i])
	}
	return solution, nil
}

func (l Lesson04b) completeChat(system, user, assistant string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: user,
			},
		},
		MaxTokens:   250,
		Temperature: 0,
		TopP:        1,
		N:           1,
		Stream:      false,
	}
	resp, err := l.completor.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("response failure for chat completion: %v", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response received")
	}
	return resp.Choices[0].Message.Content, nil
}
