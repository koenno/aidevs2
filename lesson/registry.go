package lesson

import (
	"fmt"

	"github.com/koenno/aidevs2/task"
)

type Task struct {
	Token string `json:"-"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
}

type TaskServer interface {
	FetchTask(name string, task task.AIDevsTask) error
	SendSolution(token string, solution any) error
}

type TaskSolverFactory interface {
	Create(openaiKey string) TaskSolver
}

type TaskSolver interface {
	Solve(s TaskServer) error
}

var (
	registry = make(map[string]TaskSolverFactory)
)

func AllLessons() []string {
	var lessons []string
	for k := range registry {
		lessons = append(lessons, k)
	}
	return lessons
}

func CreateTaskSolver(lessonName, openaiKey string) TaskSolver {
	lessonCreator, exists := registry[lessonName]
	if !exists {
		return UnsupportedLessonSolver{
			name: lessonName,
		}
	}
	return lessonCreator.Create(openaiKey)
}

type UnsupportedLessonSolver struct {
	name string
}

func (s UnsupportedLessonSolver) Solve(_ TaskServer) error {
	return fmt.Errorf("unsupported lesson solver %s", s.name)
}
