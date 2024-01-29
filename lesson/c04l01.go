package lesson

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/koenno/aidevs2/ai"
	"github.com/koenno/aidevs2/knowledge/country"
	"github.com/koenno/aidevs2/knowledge/currency"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const (
	C04L01CollectionName = "aidevs2_c04l01"
)

func init() {
	registry["c04l01"] = C04L01Creator{}
}

type C04L01Creator struct {
}

func (c C04L01Creator) Create(openaiKey string) TaskSolver {
	client := ai.NewChat(openaiKey, ai.WithModel(openai.GPT40613))
	return C04L01{
		chat:         client,
		funCaller:    client,
		currencyInfo: currency.NewKnowledge(),
		countryInfo:  country.NewKnowledge(),
		taskName:     "knowledge",
	}
}

type CurrencyKnowledge interface {
	TodaysCurrency(code string) (float64, error)
}

type CountryKnowledge interface {
	Info(name string, opts ...country.Option) (country.CountryInfo, error)
}

type AIFunctionCaller interface {
	ModeratedFunctionCalling(system, user, assistant string, funcDefs []openai.FunctionDefinition) (*openai.FunctionCall, error)
}

type C04L01 struct {
	currencyInfo CurrencyKnowledge
	countryInfo  CountryKnowledge
	chat         AIChat
	funCaller    AIFunctionCaller
	taskName     string
}

type C04L01Task struct {
	Task
	Question string `json:"question"`
}

func (t *C04L01Task) GetCode() int {
	return t.Code
}

func (t *C04L01Task) GetMsg() string {
	return t.Msg
}

func (t *C04L01Task) SetToken(token string) {
	t.Token = token
}

type C04L01Solution string

func (l C04L01) Solve(server TaskServer) error {
	var task C04L01Task
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
	FuncGetPopulation    = "GetPopulation"
	FuncGetCurrency      = "GetCurrency"
	FuncGetGeneralAnswer = "GetGeneralAnswer"
)

func newFunctionsDefinitions() []openai.FunctionDefinition {
	return []openai.FunctionDefinition{
		{
			Name:        FuncGetPopulation,
			Description: "Get population of a country",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"country": {
						Type:        jsonschema.String,
						Description: "The country name in english, e.g. Germany, USA",
					},
				},
				Required: []string{"country"},
			},
		},
		{
			Name:        FuncGetCurrency,
			Description: "Get actual currency",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"code": {
						Type:        jsonschema.String,
						Description: "The ISO4217 alpha code for the currency, e.g. EUR for euro, USD for United States Dollar",
					},
				},
				Required: []string{"code"},
			},
		},
		{
			Name:        FuncGetGeneralAnswer,
			Description: "Get answer for general knowledge",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"question": {
						Type:        jsonschema.String,
						Description: "The question you were asked",
					},
				},
				Required: []string{"question"},
			},
		},
	}
}

type GetPopulationParams struct {
	Country string `json:"country"`
}

type GetCurrencyParams struct {
	Code string `json:"code"`
}

type GetGeneralAnswerParams struct {
	Question string `json:"question"`
}

func (l C04L01) callFunction(name, paramsJSON string) (string, error) {
	log.Printf("calling function %s with params %#v", name, paramsJSON)
	switch name {
	case FuncGetCurrency:
		var params GetCurrencyParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", fmt.Errorf("failed to decode params json '%s': %v", paramsJSON, err)
		}
		return l.GetCurrency(params)
	case FuncGetPopulation:
		var params GetPopulationParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", fmt.Errorf("failed to decode params json '%s': %v", paramsJSON, err)
		}
		return l.GetPopulation(params)
	case FuncGetGeneralAnswer:
		var params GetGeneralAnswerParams
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", fmt.Errorf("failed to decode params json '%s': %v", paramsJSON, err)
		}
		return l.GetGeneralAnswer(params)
	default:
		return "", fmt.Errorf("unsupported function %s", name)
	}
}

func (l C04L01) GetCurrency(params GetCurrencyParams) (string, error) {
	curr, err := l.currencyInfo.TodaysCurrency(params.Code)
	if err != nil {
		return "", fmt.Errorf("failed to get todays currency: %v", err)
	}
	return fmt.Sprintf("%f", curr), nil
}

func (l C04L01) GetPopulation(params GetPopulationParams) (string, error) {
	info, err := l.countryInfo.Info(params.Country, country.WithPopulation())
	if err != nil {
		return "", fmt.Errorf("failed to get country info for %s: %v", params.Country, err)
	}
	return fmt.Sprintf("%d", info.Population), nil
}

func (l C04L01) GetGeneralAnswer(params GetGeneralAnswerParams) (string, error) {
	user := params.Question
	resp, err := l.chat.ModeratedChat("", user)
	if err != nil {
		return "", fmt.Errorf("failed to get answer: %v", err)
	}
	return resp, nil
}

func (l C04L01) getSolution(task C04L01Task) (C04L01Solution, error) {
	functionsDefinitions := newFunctionsDefinitions()
	user := task.Question
	function, err := l.funCaller.ModeratedFunctionCalling("", user, "", functionsDefinitions)
	if err != nil {
		return "", fmt.Errorf("failed to make function calling: %v", err)
	}

	answer, err := l.callFunction(function.Name, function.Arguments)
	if err != nil {
		return "", fmt.Errorf("failed to call function: %v", err)
	}

	log.Printf("Question: %s", task.Question)
	log.Printf("Answer: %s", answer)

	return C04L01Solution(answer), nil
}
