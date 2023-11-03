package lesson

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func init() {
	registry["c02l05"] = C02L05Creator{}
}

type C02L05Creator struct {
}

func (c C02L05Creator) Create(openaiKey string) TaskSolver {
	return C02L05{
		taskName: "functions",
	}
}

type C02L05 struct {
	taskName string
}

type C02L05Task struct {
	Task
}

func (t *C02L05Task) GetCode() int {
	return t.Code
}

func (t *C02L05Task) GetMsg() string {
	return t.Msg
}

func (t *C02L05Task) SetToken(token string) {
	t.Token = token
}

type C02L05Solution openai.FunctionDefinition

func (l C02L05) Solve(server TaskServer) error {
	var task Lesson01Task
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

func (l C02L05) getSolution(task Lesson01Task) (C02L05Solution, error) {
	return C02L05Solution(openai.FunctionDefinition{
		Name:        "addUser",
		Description: "Add user to the system",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"name": {
					Type:        jsonschema.String,
					Description: "Name of the user",
				},
				"surname": {
					Type:        jsonschema.String,
					Description: "Surname of the user",
				},
				"year": {
					Type:        jsonschema.Integer,
					Description: "Year of birth of the user",
				},
			},
			Required: []string{"name", "surname", "year"},
		},
	}), nil
}
