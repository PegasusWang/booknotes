package main_test

import (
	"net/http"
	"testing"
)

func TestGetoOriginalURL(t *testing.T) {
	response, err := http.Get("http://localhost:8000/v1/short/1")
	if http.StatusOK != response.StatusCode {
		t.Errorf("Expected response code %d. Got %d\n", http.StatusOK)
	}
	if err != nil {
		t.Errorf("encoutered an error:", err)
	}
}
