//go:build unit

package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIsAvailable(t *testing.T) {
	now := time.Now()

	form := url.Values{}

	form.Add(fmt.Sprintf("data[%d][service]", 0), "requesting_information")
	form.Add(fmt.Sprintf("data[%d][datetime]", 0), time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC).Format(time.DateTime))

	req, err := buildRequest(form.Encode())
	assert.NoError(t, err)
	data := doRequest(req, &http.Client{})
	assert.NotEmpty(t, data)
}

func TestMainFunction(t *testing.T) {
	os.Setenv("TEST_GREEN", "true")
	defer os.Unsetenv("TEST_GREEN")

	// Capture the output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "Test is set to true") {
		t.Errorf("Expected output to contain 'Test is set to true'")
	}
}

func TestBuildData(t *testing.T) {
	dataMap := map[int]map[string]string{
		0: {"service": "requesting_information", "datetime": "2023-10-10T08:45:00"},
		1: {"service": "requesting_information", "datetime": "2023-10-10T09:00:00"},
	}
	expected := "data[0][service]=requesting_information&data[0][datetime]=2023-10-10T08%3A45%3A00&data[1][service]=requesting_information&data[1][datetime]=2023-10-10T09%3A00%3A00"
	result := buildData(dataMap)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestBuildMap(t *testing.T) {
	result := buildMap()
	if len(result) != 120 {
		t.Errorf("Expected 120 entries, got %d", len(result))
	}
}

func TestDoRequest(t *testing.T) {
	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[true, false, true]`))
	}))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL, strings.NewReader("data"))
	result := doRequest(req, client)
	expected := []bool{true, false, true}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	}
}

func TestBuildRequest(t *testing.T) {
	data := "data"
	req, err := buildRequest(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("Expected POST method, got %s", req.Method)
	}
	if req.URL.String() != "https://russia.tmembassy.gov.tm/ru/appointment/available" {
		t.Errorf("Expected URL to be https://russia.tmembassy.gov.tm/ru/appointment/available, got %s", req.URL.String())
	}
}
