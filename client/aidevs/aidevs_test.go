package aidevs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldReturnErrorWhenResponseHasUnsupportedType(t *testing.T) {
	// given
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain")
	}))
	sut := Client[string]{}
	req, err := http.NewRequest(http.MethodGet, fakeServer.URL, nil)
	assert.NoError(t, err)

	// when
	_, err = sut.Send(req)

	// then
	assert.Error(t, err)
}

type TestPayload struct {
	Name string
}

func TestShouldReturnDecodedPauload(t *testing.T) {
	// given
	expectedPayload := TestPayload{
		Name: "ene due",
	}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(expectedPayload)
	}))
	sut := Client[TestPayload]{}
	req, err := http.NewRequest(http.MethodGet, fakeServer.URL, nil)
	assert.NoError(t, err)

	// when
	respData, err := sut.Send(req)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedPayload, respData)
}
