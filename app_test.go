//go:build unit

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.False(t, data[0])
}
func TestBuildDataReturnsEmptyStringForEmptyDataMap(t *testing.T) {
	dataMap := map[int]map[string]string{}
	expected := ""
	result := buildData(dataMap)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestBuildRequestWithEmptyData(t *testing.T) {
	data := ""
	req, err := buildRequest(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("Expected POST method, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded; charset=UTF-8" {
		t.Errorf("Expected Content-Type to be application/x-www-form-urlencoded; charset=UTF-8, got %s", req.Header.Get("Content-Type"))
	}
}

func TestDoRequestReturnsEmptySliceForEmptyResponse(t *testing.T) {
	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL, strings.NewReader("data"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	result := doRequest(req, client)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %v", result)
	}
}

func TestBuildRequestReturnsErrorForInvalidURL(t *testing.T) {
	data := "data"
	originalURL := "https://russia.tmembassy.gov.tm/ru/appointment/available"
	defer func() { baseURL = originalURL }()
	baseURL = "http://[::1]:namedport"

	_, err := buildRequest(data)
	if err == nil {
		t.Fatalf("Expected error for invalid URL, got nil")
	}
}

func TestDoRequestReturnsErrorForInvalidURL(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://invalid-url", strings.NewReader("data"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	result := doRequest(req, client)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for invalid URL, got %v", result)
	}
}

func TestDoRequestHandlesNon200StatusCode(t *testing.T) {
	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL, strings.NewReader("data"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	result := doRequest(req, client)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for non-200 status code, got %v", result)
	}
}

func TestDoRequestHandlesInvalidJSONResponse(t *testing.T) {
	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL, strings.NewReader("data"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	result := doRequest(req, client)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for invalid JSON response, got %v", result)
	}
}
