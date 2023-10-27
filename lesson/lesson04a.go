package lesson

import (
	"context"
	"fmt"
	"log"

	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["4a"] = Lesson04aCreator{}
}

type Moderator interface {
	Moderate(ctx context.Context, entry string) (bool, error)
}

type Lesson04aCreator struct {
}

func (c Lesson04aCreator) Create(openaiKey string) TaskSolver {
	return Lesson04a{
		moderator: moderation.Moderator{
			OpenAIMod: openai.NewClient(openaiKey),
		},
		taskName: "moderation",
	}
}

type Lesson04a struct {
	moderator Moderator
	taskName  string
}

type Lesson04aTask struct {
	Task
	Input []string `json:"input"`
}

func (t *Lesson04aTask) GetCode() int {
	return t.Code
}

func (t *Lesson04aTask) GetMsg() string {
	return t.Msg
}

func (t *Lesson04aTask) SetToken(token string) {
	t.Token = token
}

type Lesson04aSolution []int

func (l Lesson04a) Solve(server TaskServer) error {
	var task Lesson04aTask
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

func (l Lesson04a) getSolution(task Lesson04aTask) (Lesson04aSolution, error) {
	solution := make(Lesson04aSolution, len(task.Input))
	for i, input := range task.Input {
		moderationRequired, err := l.moderator.Moderate(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("failed to moderate entry: %v", err)
		}
		if moderationRequired {
			solution[i] = 1
		}
		log.Printf("%s | %d", input, solution[i])
	}
	return solution, nil
}
