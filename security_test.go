package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"groupie-tracker/internal/handlers"
)

func TestSecurityMeasures(t *testing.T) {
	// Test directory traversal protection
	testCases := []struct {
		path     string
		expected int
		desc     string
	}{
		{"/static/css/app.css", 200, "Valid CSS file"},
		{"/static/js/search.js", 200, "Valid JS file"},
		{"/static/../internal/templates/index.html", 404, "Directory traversal attempt"},
		{"/static/../../cmd/main.go", 404, "Directory traversal attempt"},
		{"/static/test.txt", 404, "Disallowed file extension"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handlers.StaticHandler(rr, req)

			if status := rr.Code; status != tc.expected {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tc.expected)
			}
		})
	}
}

func TestAPIEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/artists", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlers.APIArtistsHandler(rr, req)

	// Should return JSON content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
