package country

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Knowledge struct {
}

func NewKnowledge() *Knowledge {
	return &Knowledge{}
}

type options struct {
	fields []string
}

func WithPopulation() Option {
	return func(o *options) {
		o.fields = append(o.fields, "population")
	}
}

func WithCurrency() Option {
	return func(o *options) {
		o.fields = append(o.fields, "currencies")
	}
}

type Option func(*options)

type CountryInfo struct {
	Name struct {
		Common     string `json:"common"`
		Official   string `json:"official"`
		NativeName struct {
			Deu struct {
				Official string `json:"official"`
				Common   string `json:"common"`
			} `json:"deu"`
		} `json:"nativeName"`
	} `json:"name"`
	Currencies struct {
		Eur struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"EUR"`
	} `json:"currencies"`
	Population int `json:"population"`
}

func (k *Knowledge) Info(name string, opts ...Option) (CountryInfo, error) {
	cfg := &options{}
	for _, o := range opts {
		o(cfg)
	}
	rawURL := fmt.Sprintf("https://restcountries.com/v3.1/name/%s?fields=name,%s", name, strings.Join(cfg.fields, ","))
	URL, err := url.Parse(rawURL)
	if err != nil {
		return CountryInfo{}, fmt.Errorf("failed to create URL: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return CountryInfo{}, fmt.Errorf("failed to create request to %s: %v", URL, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CountryInfo{}, fmt.Errorf("failed to send request to %s: %v", URL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return CountryInfo{}, fmt.Errorf("request to %s failed with status %d: %v", URL, resp.StatusCode, err)
	}
	var info []CountryInfo
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return CountryInfo{}, fmt.Errorf("failed to decode body: %v", err)
	}
	if len(info) != 1 {
		return CountryInfo{}, fmt.Errorf("ambiguous info number %v", info)
	}
	return info[0], nil
}
