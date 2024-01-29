package lesson

import (
	"fmt"
)

func init() {
	registry["c04l05"] = C04L05Creator{}
}

type C04L05Creator struct {
}

func (c C04L05Creator) Create(openaiKey string) TaskSolver {
	return C04L05{
		taskName: "ownapipro",
	}
}

type C04L05 struct {
	taskName string
}

type C04L05Task struct {
	Task
}

func (t *C04L05Task) GetCode() int {
	return t.Code
}

func (t *C04L05Task) GetMsg() string {
	return t.Msg
}

func (t *C04L05Task) SetToken(token string) {
	t.Token = token
}

type C04L05Solution string

func (l C04L05) Solve(server TaskServer) error {
	var task C04L05Task
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

func (l C04L05) getSolution(task C04L05Task) (C04L05Solution, error) {
	return "https://gocraft.it/answer", nil
}
