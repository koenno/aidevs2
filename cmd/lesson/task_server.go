package main

import (
	"fmt"
	"log"

	"github.com/koenno/aidevs2/client/aidevs"
	"github.com/koenno/aidevs2/request"
	"github.com/koenno/aidevs2/task"
)

var (
	client     = aidevs.Client{}
	reqFactory = request.Factory{}
)

type TaskServer struct {
	ApiKey string
}

func (s TaskServer) FetchTask(name string, taskData task.AIDevsTask) error {
	taskFetcher := task.Fetcher{
		ApiKey:  s.ApiKey,
		Client:  client,
		Creator: reqFactory,
	}

	err := taskFetcher.Fetch(name, taskData)
	if err != nil {
		return fmt.Errorf("failed to solve the task %s: %v", name, err)
	}
	log.Printf("fetched following task: %#v", taskData)
	return nil
}

func (s TaskServer) SendSolution(token string, solution any) error {
	taskAnswerer := task.Answerer{
		Client:  client,
		Creator: reqFactory,
	}
	log.Printf("sending following solution: %#v", solution)
	err := taskAnswerer.Answer(token, solution)
	if err != nil {
		return fmt.Errorf("failed to send an answers: %v", err)
	}
	return nil
}
