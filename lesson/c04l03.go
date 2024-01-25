package lesson

import (
	"fmt"
	"log"

	"github.com/koenno/aidevs2/ai"
)

func init() {
	registry["c04l03"] = C04L03Creator{}
}

type C04L03Creator struct {
}

func (c C04L03Creator) Create(openaiKey string) TaskSolver {
	client := ai.NewVisioner(openaiKey)
	return C04L03{
		visioner: client,
		taskName: "gnome",
	}
}

type AIVisioner interface {
	ModeratedSee(system, user, assistant, imageURI string) (string, error)
}

type C04L03 struct {
	visioner AIVisioner
	taskName string
}

type C04L03Task struct {
	Task
	URL string `json:"url"`
}

func (t *C04L03Task) GetCode() int {
	return t.Code
}

func (t *C04L03Task) GetMsg() string {
	return t.Msg
}

func (t *C04L03Task) SetToken(token string) {
	t.Token = token
}

type C04L03Solution string

func (l C04L03) Solve(server TaskServer) error {
	var task C04L03Task
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

func (l C04L03) getSolution(task C04L03Task) (C04L03Solution, error) {
	system := `
I am supposed to watch only pictures with dwarfs.
If there is no dwarf nor gnome on the picture answer shortly: "error".
If there is a dwarf or gnome on the picture answer ultra-concise and in polish.
`
	user := "What color is the hat of a dwarf?"
	answer, err := l.visioner.ModeratedSee(system, user, "", task.URL)
	if err != nil {
		return "", fmt.Errorf("failed to describe following picture %s: %v", task.URL, err)
	}

	log.Printf("Question: %s", task.URL)
	log.Printf("Answer: %s", answer)

	return C04L03Solution(answer), nil
}
