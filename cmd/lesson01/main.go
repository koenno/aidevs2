package main

import (
	"flag"
	"log"

	"github.com/koenno/aidevs2/client/aidevs"
	"github.com/koenno/aidevs2/request"
	"github.com/koenno/aidevs2/task"
)

const (
	taskName = "helloapi"
)

type Task struct {
	Token  string `json:"-"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Cookie string `json:"cookie"`
}

func (t *Task) GetCode() int {
	return t.Code
}

func (t *Task) GetMsg() string {
	return t.Msg
}

func (t *Task) SetToken(token string) {
	t.Token = token
}

func main() {
	apiKey := flag.String("apikey", "", "your API key")
	flag.Parse()
	if *apiKey == "" {
		log.Fatalf("API key is required")
	}

	c := aidevs.Client{}
	reqFactory := request.Factory{}

	t := fetchTask(c, reqFactory, *apiKey)
	solution := solveTask(t)
	sendSolution(c, reqFactory, t.Token, solution)
}

func fetchTask(c aidevs.Client, reqFactory request.Factory, apiKey string) Task {
	taskFetcher := task.Fetcher{
		ApiKey:  apiKey,
		Client:  c,
		Creator: reqFactory,
	}

	var taskData Task
	err := taskFetcher.Fetch(taskName, &taskData)
	if err != nil {
		log.Fatalf("failed to solve task %s: %v", taskName, err)
	}
	return taskData
}

func solveTask(t Task) string {
	return t.Cookie
}

func sendSolution(c aidevs.Client, reqFactory request.Factory, token, solution string) {
	taskAnswerer := task.Answerer{
		Client:  c,
		Creator: reqFactory,
	}

	err := taskAnswerer.Answer(token, solution)
	if err != nil {
		log.Fatalf("failed to send an answer to task %s: %v", taskName, err)
	}
}
