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
	registry["c03l03"] = C03L03Creator{}
}

type C03L03Creator struct {
}

func (c C03L03Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C03L03{
		completor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "whoami",
	}
}

type C03L03 struct {
	completor ChatCompletor
	moderator Moderator
	taskName  string
}

type C03L03Task struct {
	Task
	Hint string `json:"hint"`
}

func (t *C03L03Task) GetCode() int {
	return t.Code
}

func (t *C03L03Task) GetMsg() string {
	return t.Msg
}

func (t *C03L03Task) SetToken(token string) {
	t.Token = token
}

type C03L03Solution string

func (l C03L03) Solve(server TaskServer) error {
	var task C03L03Task
	err := server.FetchTask(l.taskName, &task)
	if err != nil {
		return fmt.Errorf("failed to fetch task: %v", err)
	}
	solution, err := l.getSolution(server, task)
	if err != nil {
		return fmt.Errorf("failed to find solution: %v", err)
	}
	err = server.SendSolution(task.Token, solution)
	if err != nil {
		return fmt.Errorf("failed to send solution: %v", err)
	}
	return nil
}

func (l C03L03) getSolution(server TaskServer, task C03L03Task) (C03L03Solution, error) {
	const rules = `
	Strict rules of this conversation:
	- I guess a person name based on facts you give me
	- I'll always skip any comments entirely
	- I keep my answers ultra-concise
	- I'm always truthful and honestly say "I don't know" when you ask me about something beyond my current knowledge
	- If I am not sure the anser I say "I don't know"
	- I answer only with a name of the person
	- I answer only in polish
	- I answer only if I am 100% sure
	`

	var resp string
	var facts []string
	if task.Hint != "" {
		facts = append(facts, task.Hint)
	}
	for {
		if len(facts) != 0 {
			prompt := fmt.Sprintf("\nFacts:\n%s", strings.Join(facts, "\n"))
			system := rules
			var err error
			resp, err = l.moderatedChat(system, prompt, "")
			if err != nil {
				return "", fmt.Errorf("solution chat failure: %v", err)
			}
			log.Printf("%s | %s", prompt, resp)
		}

		if resp == "I don't know" || resp == "" {
			var moreInfo C03L03Task
			err := server.FetchTask(l.taskName, &moreInfo)
			if err != nil {
				return "", fmt.Errorf("failed to fetch more info: %v", err)
			}
			facts = append(facts, moreInfo.Hint)
		} else {
			break
		}
	}
	return C03L03Solution(resp), nil
}

func (l C03L03) moderatedChat(system, user, assistant string) (string, error) {
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

func (l C03L03) completeChat(system, user, assistant string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
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
