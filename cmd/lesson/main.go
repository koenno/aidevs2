package main

import (
	"flag"
	"log"

	"github.com/koenno/aidevs2/lesson"
)

type Task struct {
	Token string   `json:"-"`
	Code  int      `json:"code"`
	Msg   string   `json:"msg"`
	Input []string `json:"input"`
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
	aidevsKey := flag.String("aidevsKey", "", "your AIDevs API key")
	openaiKey := flag.String("openaiKey", "", "your OpenAI API key")
	lessonName := flag.String("lesson", "", "lesson name")
	flag.Parse()
	if *aidevsKey == "" {
		log.Fatalf("AIDevs API key is required")
	}
	if *openaiKey == "" {
		log.Fatalf("OpenAI API key is required")
	}
	if *lessonName == "" {
		log.Fatalf("lesson name is required")
	}

	ts := TaskServer{
		ApiKey: *aidevsKey,
	}
	solver := lesson.CreateTaskSolver(*lessonName, *openaiKey)
	err := solver.Solve(ts)
	if err != nil {
		log.Fatalf("failed to solve task for lesson %s: %s", *lessonName, err)
	}
}
