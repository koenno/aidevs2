package currency

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type Knowledge struct {
}

func NewKnowledge() *Knowledge {
	return &Knowledge{}
}

type CurrencyInfo struct {
	Table    string `json:"table"`
	Currency string `json:"currency"`
	Code     string `json:"code"`
	Rates    []struct {
		No            string  `json:"no"`
		EffectiveDate string  `json:"effectiveDate"`
		Mid           float64 `json:"mid"`
	} `json:"rates"`
}

func (k *Knowledge) TodaysCurrency(code string) (float64, error) {
	rawURL := fmt.Sprintf("http://api.nbp.pl/api/exchangerates/rates/a/%s/today/", code)
	URL, err := url.Parse(rawURL)
	if err != nil {
		return 0, fmt.Errorf("failed to create URL: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request to %s: %v", URL, err)
	}
	req.Header.Add("Host", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request to %s: %v", URL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("request to %s failed with status %d: %v", URL, resp.StatusCode, err)
	}
	var info CurrencyInfo
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return 0, fmt.Errorf("failed to decode body: %v", err)
	}
	if len(info.Rates) != 1 {
		return 0, fmt.Errorf("ambiguous info number %v", info)
	}
	return info.Rates[0].Mid, nil
}
