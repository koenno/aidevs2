package task

import "net/http"

type Response struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Token string `json:"token"`
}

type Client[T any] interface {
	Send(r *http.Request) (T, error)
}

type RequestCreator interface {
	Create(name string) (http.Request, error)
}

type ResolutionSender[T any] struct {
	ApiKey string
	Client Client[T]
}

func (s ResolutionSender[T]) Send(taskName string) {

}
