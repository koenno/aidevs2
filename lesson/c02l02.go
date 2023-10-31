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
	registry["c02l02"] = C02L02Creator{}
}

type C02L02Creator struct {
}

func (c C02L02Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C02L02{
		completor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "inprompt",
	}
}

type C02L02 struct {
	completor ChatCompletor
	moderator Moderator
	taskName  string
}

type C02L02Task struct {
	Task
	Input    []string `json:"input"`
	Question string   `json:"question"`
}

func (t *C02L02Task) GetCode() int {
	return t.Code
}

func (t *C02L02Task) GetMsg() string {
	return t.Msg
}

func (t *C02L02Task) SetToken(token string) {
	t.Token = token
}

type C02L02Solution string

func (l C02L02) Solve(server TaskServer) error {
	var task C02L02Task
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

func (l C02L02) getSolution(task C02L02Task) (C02L02Solution, error) {
	const rules = `
	Strict rules of this conversation:
	- I'm strictly forbidden to use any knowledge outside the context below and I always refuse to answer such question mentioning this rule.
	- Because of your expertise, I'll always skip any comments entirely
	- I keep my answers ultra-concise
	- I'm always truthful and honestly say "I don't know" when you ask me about something beyond my current knowledge
	- I'm aware only I have access to the context right now
	`
	prompt := task.Question
	nameToFacts, err := l.getContextMap(task)
	if err != nil {
		return "", fmt.Errorf("failed to relate facts to names: %v", err)
	}
	askedName, err := l.getNameByAI(task.Question)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve name from the question: %v", err)
	}
	askedFacts, exist := nameToFacts[askedName]
	if !exist {
		return "", fmt.Errorf("asked name %s does not exist in fact database", askedName)
	}
	promptContext := l.getContext(askedFacts)
	system := rules + promptContext
	resp, err := l.moderatedChat(system, prompt, "")
	if err != nil {
		return "", fmt.Errorf("solution chat failure: %v", err)
	}
	log.Printf("%s | %s", prompt, resp)
	return C02L02Solution(resp), nil
}

func (l C02L02) getName(text string) (string, error) {
	tokens := strings.Split(text, " ")
	return tokens[0], nil
}

func (l C02L02) getNameByAI(text string) (string, error) {
	user := "give only the name"
	resp, err := l.moderatedChat(text, user, "")
	if err != nil {
		return "", fmt.Errorf("name retrieval chat failure: %v", err)
	}
	return resp, nil
}

func (l C02L02) getContextMap(task C02L02Task) (map[string][]string, error) {
	nameToFacts := make(map[string][]string)
	for _, input := range task.Input {
		name, err := l.getName(input)
		if err != nil {
			return nil, fmt.Errorf("failed to get name for `%s`: %v", input, err)
		}
		facts := nameToFacts[name]
		facts = append(facts, input)
		nameToFacts[name] = facts
	}
	return nameToFacts, nil
}

func (l C02L02) getContext(facts []string) string {
	var strBuilder strings.Builder
	for _, fact := range facts {
		strBuilder.WriteString(fact + "\n")
	}
	return fmt.Sprintf("\nContext```%s```", strBuilder.String())
}

func (l C02L02) moderatedChat(system, user, assistant string) (string, error) {
	moderationRequired, err := l.moderator.Moderate(context.Background(), system)
	if err != nil {
		return "", fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return "", fmt.Errorf("entry breaks openai usage policies: %s", system)
	}
	resp, err := l.completeChat(system, user, assistant)
	if err != nil {
		return "", fmt.Errorf("failed to complete moderated chat: %v", err)
	}
	return resp, nil
}

func (l C02L02) completeChat(system, user, assistant string) (string, error) {
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
