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

type Client struct {
}

func (c Client) Send(r *http.Request, respPayload any) error {
	log.Printf("sending request %s to %s", r.Method, r.URL)
	resp, err := httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to send request %s to %s", r.Method, r.URL)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body from %s", r.URL)
		}
	}()
	// TODO: check status code
	contentType := resp.Header.Get("content-type")
	if contentType != "application/json" {
		return fmt.Errorf("unsupported response content type from %s", r.URL)
	}

	err = json.NewDecoder(resp.Body).Decode(respPayload)
	if err != nil {
		return fmt.Errorf("failed to convert response payload: %v", err)
	}
	return nil
}
