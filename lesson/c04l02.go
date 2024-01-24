package lesson

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/koenno/aidevs2/ai"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const (
	C04L02CollectionName = "aidevs2_c04l02"
)

func init() {
	registry["c04l02"] = C04L02Creator{}
}

type C04L02Creator struct {
}

func (c C04L02Creator) Create(openaiKey string) TaskSolver {
	client := ai.NewChat(openaiKey, ai.WithModel(openai.GPT40613))
	return C04L02{
		chat:      client,
		funCaller: client,
		taskName:  "tools",
	}
}

type C04L02 struct {
	chat      AIChat
	funCaller AIFunctionCaller
	taskName  string
}

type C04L02Task struct {
	Task
	Question string `json:"question"`
}

func (t *C04L02Task) GetCode() int {
	return t.Code
}

func (t *C04L02Task) GetMsg() string {
	return t.Msg
}

func (t *C04L02Task) SetToken(token string) {
	t.Token = token
}

type C04L02Solution ToDoAndCalendar

func (l C04L02) Solve(server TaskServer) error {
	var task C04L02Task
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

const (
	FuncToDo     = "ToDo"
	FuncCalendar = "Calendar"
)

func newFunctionsDefinitionsC04L02() []openai.FunctionDefinition {
	return []openai.FunctionDefinition{
		{
			Name:        FuncToDo,
			Description: "Reminder that I need to do something without any date nor time given",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"tool": {
						Type:        jsonschema.String,
						Description: "Must always have following value: ToDo",
					},
					"desc": {
						Type:        jsonschema.String,
						Description: "Something I need to do",
					},
				},
				Required: []string{"tool", "desc"},
			},
		},
		{
			Name:        FuncCalendar,
			Description: "A meeting or event that I am supposed to have in a given date",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"tool": {
						Type:        jsonschema.String,
						Description: "Must always have following value: Calendar",
					},
					"desc": {
						Type:        jsonschema.String,
						Description: "The meeting or event description",
					},
					"date": {
						Type:        jsonschema.String,
						Description: "The date in format in format YYYY-MM-DD",
					},
				},
				Required: []string{"tool", "desc", "date"},
			},
		},
	}
}

type ToDoAndCalendar struct {
	Tool string `json:"tool"`
	Desc string `json:"desc"`
	Date string `json:"date,omitempty"`
}

func (l C04L02) callFunction(name, paramsJSON string) (ToDoAndCalendar, error) {
	log.Printf("calling function %s with params %#v", name, paramsJSON)
	var data ToDoAndCalendar
	err := json.Unmarshal([]byte(paramsJSON), &data)
	return data, err
}

func (l C04L02) getSolution(task C04L02Task) (C04L02Solution, error) {
	functionsDefinitions := newFunctionsDefinitionsC04L02()
	system := fmt.Sprintf("today is %s", time.Now().Format("Monday, 02 January 2006"))
	log.Println(system)
	user := task.Question
	function, err := l.funCaller.ModeratedFunctionCalling(system, user, "", functionsDefinitions)
	if err != nil {
		return C04L02Solution{}, fmt.Errorf("failed to make function calling: %v", err)
	}

	answer, err := l.callFunction(function.Name, function.Arguments)
	if err != nil {
		return C04L02Solution{}, fmt.Errorf("failed to call function: %v", err)
	}

	log.Printf("Question: %s", task.Question)
	log.Printf("Answer: %s", answer)

	return C04L02Solution(answer), nil
}
