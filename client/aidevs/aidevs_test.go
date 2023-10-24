package aidevs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestPayload struct {
	Name string
}

func TestShouldReturnErrorWhenResponseHasUnsupportedType(t *testing.T) {
	// given
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain")
	}))
	sut := Client{}
	req, err := http.NewRequest(http.MethodGet, fakeServer.URL, nil)
	assert.NoError(t, err)
	var respPayload TestPayload

	// when
	err = sut.Send(req, &respPayload)

	// then
	assert.Error(t, err)
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
	sut := Client{}
	req, err := http.NewRequest(http.MethodGet, fakeServer.URL, nil)
	assert.NoError(t, err)
	var respPayload TestPayload

	// when
	err = sut.Send(req, &respPayload)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedPayload, respPayload)
}
