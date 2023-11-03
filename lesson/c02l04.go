package lesson

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["c02l04"] = C02L04Creator{}
}

type C02L04Creator struct {
}

func (c C02L04Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C02L04{
		transcriptor: client,
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "whisper",
	}
}

type Transcriptor interface {
	CreateTranscription(context.Context, openai.AudioRequest) (openai.AudioResponse, error)
}

type C02L04 struct {
	transcriptor Transcriptor
	moderator    Moderator
	taskName     string
}

type C02L04Task struct {
	Task
}

func (t *C02L04Task) GetCode() int {
	return t.Code
}

func (t *C02L04Task) GetMsg() string {
	return t.Msg
}

func (t *C02L04Task) SetToken(token string) {
	t.Token = token
}

type C02L04Solution string

func (l C02L04) Solve(server TaskServer) error {
	var task C02L04Task
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

func (l C02L04) getSolution(task C02L04Task) (C02L04Solution, error) {
	const phrase = "please return transcription of this file: "
	parts := strings.SplitAfter(task.Msg, phrase)
	fileURL := parts[len(parts)-1]
	text, err := l.transcript(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe: %v", err)
	}
	log.Printf("%s | %v", fileURL, text)
	return C02L04Solution(text), nil
}

func (l C02L04) transcript(URL string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", fmt.Errorf("download failure: %v", err)
	}
	defer resp.Body.Close()

	fileName := getFilePath(URL)
	log.Printf("file name: %s", fileName)

	req := openai.AudioRequest{
		Model:       openai.Whisper1,
		FilePath:    fileName,
		Reader:      resp.Body,
		Prompt:      "",
		Temperature: 0,
		Language:    "pl",
		Format:      openai.AudioResponseFormatJSON,
	}
	transResp, err := l.transcriptor.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("transcription error: %v", err)
	}

	return transResp.Text, nil
}

func getFilePath(URL string) string {
	parts := strings.Split(URL, "/")
	return parts[len(parts)-1]
}
