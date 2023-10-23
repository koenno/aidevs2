package aidevs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	httpClient = http.Client{
		Timeout: 10 * time.Second,
	}
)

type Client[T any] struct {
}

func (c Client[T]) Send(r *http.Request) (T, error) {
	log.Printf("sending request %s to %s", r.Method, r.URL)
	var respData T
	resp, err := httpClient.Do(r)
	if err != nil {
		return respData, fmt.Errorf("failed to send request %s to %s", r.Method, r.URL)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body from %s", r.URL)
		}
	}()
	contentType := resp.Header.Get("content-type")
	if contentType != "application/json" {
		return respData, fmt.Errorf("unsupported response content type from %s", r.URL)
	}

	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return respData, fmt.Errorf("failed to convert response payload: %v", err)
	}
	return respData, nil
}
