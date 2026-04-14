package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServerClient(server *httptest.Server) *Client {
	return &Client{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: server.Client(),
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Fatalf("Authorization header = %q", auth)
		}
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Fatalf("Accept header = %q", accept)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)

	body, err := client.Get("/test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("Get() body = %s", body)
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type header = %q", ct)
		}
		data, _ := io.ReadAll(r.Body)
		if string(data) != `{"name":"value"}` {
			t.Fatalf("body = %s", data)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)

	payload := struct {
		Name string `json:"name"`
	}{Name: "value"}

	body, err := client.Post("/test", payload)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if string(body) != `{"created":true}` {
		t.Fatalf("Post() body = %s", body)
	}
}

func TestClientPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type header = %q", ct)
		}
		data, _ := io.ReadAll(r.Body)
		if string(data) != `{"name":"patched"}` {
			t.Fatalf("body = %s", data)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"patched":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	payload := struct {
		Name string `json:"name"`
	}{Name: "patched"}

	body, err := client.Patch("/test", payload)
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	if string(body) != `{"patched":true}` {
		t.Fatalf("Patch() body = %s", body)
	}
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestServerClient(server)

	_, err := client.Delete("/test")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestClientErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		expectedSubstr string
	}{
		{"unauthorized", http.StatusUnauthorized, "authentication failed"},
		{"forbidden", http.StatusForbidden, "forbidden"},
		{"not_found", http.StatusNotFound, "not found"},
		{"rate_limit", http.StatusTooManyRequests, "rate limit"},
		{"server_error", http.StatusInternalServerError, "API error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.name))
			}))
			defer server.Close()

			client := newTestServerClient(server)

			_, err := client.Get("/test")
			if err == nil {
				t.Fatalf("expected error for status %d", tt.status)
			}

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("expected *APIError, got %T", err)
			}
			if apiErr.StatusCode != tt.status {
				t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if !strings.Contains(apiErr.Error(), tt.expectedSubstr) {
				t.Fatalf("error message %q missing substring %q", apiErr.Error(), tt.expectedSubstr)
			}
		})
	}
}

func TestAPIErrorErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		expected string
	}{
		{"401", 401, "detail", "authentication failed (HTTP 401): invalid or missing token. detail"},
		{"402", 402, "detail", "feature unavailable (HTTP 402): blocked by tariff/plan. detail"},
		{"403", 403, "detail", "forbidden (HTTP 403): insufficient permissions. detail"},
		{"404", 404, "detail", "not found (HTTP 404): resource does not exist. detail"},
		{"429", 429, "detail", "rate limit exceeded (HTTP 429): too many requests. detail"},
		{"500", 500, "detail", "API error (HTTP 500): detail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&APIError{StatusCode: tt.status, Body: tt.body}).Error()
			if err != tt.expected {
				t.Fatalf("Error() = %q, want %q", err, tt.expected)
			}
		})
	}
}
