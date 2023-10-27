package task

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

var (
	ErrFetch  = errors.New("failed to fetch a task")
	ErrAnswer = errors.New("failed to send answer")

	errAuth = errors.New("failed to authenticate")
	errTask = errors.New("failed to fetch task")
)

//go:generate mockery --name=Client --case underscore --with-expecter
type Client interface {
	Send(r *http.Request, respPayload any) error
}

//go:generate mockery --name=TaskRequestCreator --case underscore --with-expecter
type TaskRequestCreator interface {
	Authenticate(apiKey, taskName string) (*http.Request, error)
	Task(token string) (*http.Request, error)
}

type Fetcher struct {
	ApiKey  string
	Client  Client
	Creator TaskRequestCreator
}

type AIDevsTask interface {
	GetCode() int
	GetMsg() string
	SetToken(string)
}

func (s Fetcher) Fetch(taskName string, resp AIDevsTask) error {
	token, err := s.authenticate(taskName)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFetch, err)
	}

	err = s.fetchTask(token, resp)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFetch, err)
	}
	resp.SetToken(token)
	return nil
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type AuthorizationResponse struct {
	Response
	Token string `json:"token"`
}

func (s Fetcher) authenticate(taskName string) (string, error) {
	req, err := s.Creator.Authenticate(s.ApiKey, taskName)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errAuth, err)
	}
	var resp AuthorizationResponse
	err = s.Client.Send(req, &resp)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errAuth, err)
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("%w: error response: %d - %s", errAuth, resp.Code, resp.Msg)
	}
	if resp.Token == "" {
		return "", fmt.Errorf("%w: empty token", errAuth)
	}
	log.Printf("auth resp message: %s", resp.Msg)
	return resp.Token, nil
}

func (s Fetcher) fetchTask(token string, resp AIDevsTask) error {
	req, err := s.Creator.Task(token)
	if err != nil {
		return fmt.Errorf("%w: %v", errTask, err)
	}
	err = s.Client.Send(req, resp)
	if err != nil {
		return fmt.Errorf("%w: %v", errTask, err)
	}
	if resp.GetCode() != 0 {
		return fmt.Errorf("%w: error response: %d - %s", errTask, resp.GetCode(), resp.GetMsg())
	}
	log.Printf("task resp message: %s", resp.GetMsg())
	return nil
}

//go:generate mockery --name=AnswerRequestCreator --case underscore --with-expecter
type AnswerRequestCreator interface {
	Answer(token string, answerData any) (*http.Request, error)
}

type Answerer struct {
	Client  Client
	Creator AnswerRequestCreator
}

func (a Answerer) Answer(token string, answerData any) error {
	req, err := a.Creator.Answer(token, answerData)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAnswer, err)
	}
	var resp Response
	err = a.Client.Send(req, &resp)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAnswer, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("%w: error response: %d - %s", ErrAnswer, resp.Code, resp.Msg)
	}
	log.Printf("answer resp message: %s", resp.Msg)
	return nil
}
