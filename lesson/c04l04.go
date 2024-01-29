package lesson

import (
	"fmt"
)

func init() {
	registry["c04l04"] = C04L04Creator{}
}

type C04L04Creator struct {
}

func (c C04L04Creator) Create(openaiKey string) TaskSolver {
	return C04L04{
		taskName: "ownapi",
	}
}

type C04L04 struct {
	taskName string
}

type C04L04Task struct {
	Task
}

func (t *C04L04Task) GetCode() int {
	return t.Code
}

func (t *C04L04Task) GetMsg() string {
	return t.Msg
}

func (t *C04L04Task) SetToken(token string) {
	t.Token = token
}

type C04L04Solution string

func (l C04L04) Solve(server TaskServer) error {
	var task C04L04Task
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

func (l C04L04) getSolution(task C04L04Task) (C04L04Solution, error) {
	return "https://gocraft.it/answer", nil
}
