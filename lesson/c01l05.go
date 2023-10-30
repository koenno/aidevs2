package lesson

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["c01l05"] = C01L05Creator{}
}

type C01L05Creator struct {
}

func (c C01L05Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C01L05{
		completor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "liar",
	}
}

type C01L05 struct {
	completor ChatCompletor
	moderator Moderator
	taskName  string
}

type C01L05Task struct {
	Task
}

func (t *C01L05Task) GetCode() int {
	return t.Code
}

func (t *C01L05Task) GetMsg() string {
	return t.Msg
}

func (t *C01L05Task) SetToken(token string) {
	t.Token = token
}

type C01L05Solution string

func (l C01L05) Solve(server TaskServer) error {
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

func (l C01L05) getSolution(task Lesson01Task) (C01L05Solution, error) {
	const question = "What is a capital of Poland?"
	system := fmt.Sprintf(`Keep answers simple - YES, NO without dot. Having a question "%s". Can you answer it in the following way `, question)

	answer, err := sendForm(task.Token, question)
	if err != nil {
		return "", fmt.Errorf("failed to send form request: %v", err)
	}
	moderationRequired, err := l.moderator.Moderate(context.Background(), system)
	if err != nil {
		return "", fmt.Errorf("failed to moderate entry: %v", err)
	}
	if moderationRequired {
		return "", fmt.Errorf("entry breaks openai usage policies: %s", system)
	}
	resp, err := l.completeChat(system, answer, "")
	if err != nil {
		return "", fmt.Errorf("failed to complete chat: %v", err)
	}
	log.Printf("%s | %s", answer, resp)
	return C01L05Solution(resp), nil
}

func sendForm(token, question string) (string, error) {
	data := url.Values{
		"question": {question},
	}
	rawURL := fmt.Sprintf("https://zadania.aidevs.pl/task/%s", token)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %v", err)
	}
	resp, err := http.PostForm(URL.String(), data)
	if err != nil {
		return "", fmt.Errorf("failed to send form to %s: %v", URL, err)
	}

	var res map[string]string
	json.NewDecoder(resp.Body).Decode(&res)
	return res["answer"], nil
}

func (l C01L05) completeChat(system, user, assistant string) (string, error) {
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
