package lesson

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/koenno/aidevs2/client/scraper"
	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["c03l02"] = C03L02Creator{}
}

type C03L02Creator struct {
}

type Scraper interface {
	Send(r *http.Request) (string, error)
}

func (c C03L02Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C03L02{
		completor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		scraperClient: &scraper.Client{},
		taskName:      "scraper",
	}
}

type C03L02 struct {
	completor     ChatCompletor
	moderator     Moderator
	taskName      string
	scraperClient Scraper
}

type C03L02Task struct {
	Task
	Input    string `json:"input"`
	Question string `json:"question"`
}

func (t *C03L02Task) GetCode() int {
	return t.Code
}

func (t *C03L02Task) GetMsg() string {
	return t.Msg
}

func (t *C03L02Task) SetToken(token string) {
	t.Token = token
}

type C03L02Solution string

func (l C03L02) Solve(server TaskServer) error {
	var task C03L02Task
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

func (l C03L02) getSolution(task C03L02Task) (C03L02Solution, error) {
	const rules = `
	Strict rules of this conversation:
	- I'm strictly forbidden to use any knowledge outside the context below and I always refuse to answer such question mentioning this rule.
	- Because of your expertise, I'll always skip any comments entirely
	- I keep my answers ultra-concise
	- I'm always truthful and honestly say "I don't know" when you ask me about something beyond my current knowledge
	- I'm aware only I have access to the context right now
	`
	prompt := task.Question
	promptContext, err := l.getContext(task.Input)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %v", err)
	}
	system := rules + promptContext
	resp, err := l.moderatedChat(system, prompt, "")
	if err != nil {
		return "", fmt.Errorf("solution chat failure: %v", err)
	}
	log.Printf("%s | %s", prompt, resp)
	return C03L02Solution(resp), nil
}

func (l C03L02) getContext(webAddr string) (string, error) {
	URL, err := url.Parse(webAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	var resp string
	r := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), retrier.DefaultClassifier{})
	err = r.Run(func() error {
		resp, err = l.scraperClient.Send(req)
		if err != nil {
			return fmt.Errorf("failed to scrap: %v", err)
		}
		return nil
	})
	return fmt.Sprintf("\nContext```%s```", resp), nil
}

func (l C03L02) moderatedChat(system, user, assistant string) (string, error) {
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

func (l C03L02) completeChat(system, user, assistant string) (string, error) {
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
