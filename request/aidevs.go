package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	endpoint = "https://zadania.aidevs.pl"
)

type Factory struct {
}

type AuthenticateRequest struct {
	ApiKey string `json:"apikey"`
}

func (f Factory) Authenticate(apiKey, taskName string) (*http.Request, error) {
	rawURL := fmt.Sprintf("%s/token/%s", endpoint, taskName)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse an authentication url: %v", err)
	}
	payload := AuthenticateRequest{
		ApiKey: apiKey,
	}
	bb, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode authentication payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, URL.String(), bytes.NewReader(bb))
	if err != nil {
		return nil, fmt.Errorf("failed to create an authentication request: %v", err)
	}
	return req, nil
}

func (f Factory) Task(token string) (*http.Request, error) {
	rawURL := fmt.Sprintf("%s/task/%s", endpoint, token)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a task url: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a task request: %v", err)
	}
	return req, nil
}

type AnswerRequest struct {
	Answer string `json:"answer"`
}

func (f Factory) Answer(token, answerStr string) (*http.Request, error) {
	rawURL := fmt.Sprintf("%s/answer/%s", endpoint, token)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a task url: %v", err)
	}
	payload := AnswerRequest{
		Answer: answerStr,
	}
	bb, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode task payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, URL.String(), bytes.NewReader(bb))
	if err != nil {
		return nil, fmt.Errorf("failed to create a task request: %v", err)
	}
	return req, nil
}
