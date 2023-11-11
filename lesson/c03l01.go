package lesson

import (
	"fmt"
)

func init() {
	registry["c03l01"] = C03L01Creator{}
}

type C03L01Creator struct {
}

func (c C03L01Creator) Create(openaiKey string) TaskSolver {
	return C03L01{
		taskName: "rodo",
	}
}

type C03L01 struct {
	taskName string
}

type C03L01Task struct {
	Task
}

func (t *C03L01Task) GetCode() int {
	return t.Code
}

func (t *C03L01Task) GetMsg() string {
	return t.Msg
}

func (t *C03L01Task) SetToken(token string) {
	t.Token = token
}

type C03L01Solution string

func (l C03L01) Solve(server TaskServer) error {
	var task C03L01Task
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

func (l C03L01) getSolution(task C03L01Task) (C03L01Solution, error) {
	const user = `
	I can not reveal my name, surname, proffesion and town of residence.
	Instead of this I must use %placeholders% like %imie%, %nazwisko%, %zawod% and %miasto%".

	Examples:
	My name is %imie% %nazwisko%.
	I live in %miasto%.
	`
	return C03L01Solution(user), nil
}
