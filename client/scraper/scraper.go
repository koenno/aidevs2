package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	httpClient = http.Client{
		Timeout: 2 * time.Minute,
	}
)

type Client struct {
}

func (c Client) Send(r *http.Request) (string, error) {
	log.Printf("sending request %s to %s", r.Method, r.URL)
	resp, err := httpClient.Do(r)
	if err != nil {
		return "", fmt.Errorf("failed to send request %s to %s", r.Method, r.URL)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body from %s", r.URL)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, r.URL)
	}

	contentType := resp.Header.Get("content-type")
	if contentType != "text/plain;charset=UTF-8" {
		return "", fmt.Errorf("unsupported response content type from %s", r.URL)
	}

	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %v", err)
	}

	result := string(bb)
	if strings.HasPrefix(result, "") {
		return "", fmt.Errorf("failure response from %s: %s", r.URL, result)
	}

	return result, nil
}
