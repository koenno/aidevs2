package lesson

import (
	"fmt"
)

func init() {
	registry["1"] = Lesson01Creator{}
}

type Lesson01Creator struct {
}

func (c Lesson01Creator) Create(openaiKey string) TaskSolver {
	return Lesson01{
		taskName: "helloapi",
	}
}

type Lesson01 struct {
	taskName string
}

type Lesson01Task struct {
	Task
	Cookie string `json:"cookie"`
}

func (t *Lesson01Task) GetCode() int {
	return t.Code
}

func (t *Lesson01Task) GetMsg() string {
	return t.Msg
}

func (t *Lesson01Task) SetToken(token string) {
	t.Token = token
}

type Lesson01Solution string

func (l Lesson01) Solve(server TaskServer) error {
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

func (l Lesson01) getSolution(task Lesson01Task) (Lesson01Solution, error) {
	return Lesson01Solution(task.Cookie), nil
}
