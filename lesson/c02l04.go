package lesson

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/koenno/aidevs2/moderation"
	"github.com/sashabaranov/go-openai"
)

func init() {
	registry["c02l04"] = C02L04Creator{}
}

type C02L04Creator struct {
}

func (c C02L04Creator) Create(openaiKey string) TaskSolver {
	client := openai.NewClient(openaiKey)
	return C02L04{
		moderator: moderation.Moderator{
			OpenAIMod: client,
		},
		taskName: "whisper",
	}
}

type C02L04 struct {
	moderator Moderator
	taskName  string
}

type C02L04Task struct {
	Task
}

func (t *C02L04Task) GetCode() int {
	return t.Code
}

func (t *C02L04Task) GetMsg() string {
	return t.Msg
}

func (t *C02L04Task) SetToken(token string) {
	t.Token = token
}

type C02L04Solution string

func (l C02L04) Solve(server TaskServer) error {
	var task C02L04Task
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

func (l C02L04) getSolution(task C02L04Task) (C02L04Solution, error) {
	const phrase = "please return transcription of this file: "
	parts := strings.SplitAfter(task.Msg, phrase)
	fileURL := parts[len(parts)-1]
	fileName := getFilePath(fileURL)
	err := DownloadFile(fileName, fileURL)
	if err != nil {
		return "", fmt.Errorf("solution %s failure: %v", l.taskName, err)
	}
	_, err = l.whisper(fileName, "pl", ".", "txt")
	if err != nil {
		return "", fmt.Errorf("failed to transcript file %s: %v", fileName, err)
	}
	newFileName := newExtension(fileName, "txt")
	f, err := os.OpenFile(newFileName, os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %v", newFileName, err)
	}
	bb, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", newFileName, err)
	}
	transcript := string(bb)
	// transcript = strings.ReplaceAll(transcript, "\n", " ")
	// transcript = strings.Trim(transcript, " ")
	log.Printf("%s | %v", fileURL, transcript)
	return C02L04Solution(transcript), nil
}

func getFilePath(fileURL string) string {
	parts := strings.Split(fileURL, "/")
	fileName := parts[len(parts)-1]
	return fileName
}

func newExtension(fileName, extension string) string {
	parts := strings.Split(fileName, ".")
	parts[len(parts)-1] = extension
	return strings.Join(parts, ".")
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (l C02L04) whisper(path, lang, out, format string) (string, error) {
	// whisper mateusz.mp3 --language pl -f all -o whisper_out
	args := []string{
		path,
		"--language", lang,
		"-f", format,
		"-o", out,
		"--model", "base",
	}
	cmd := exec.Command("whisper", args...)
	bb, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("whisper failure: %v", err)
	}
	result := string(bb)
	log.Printf("whisper output: %s", result)
	return result, nil
}
