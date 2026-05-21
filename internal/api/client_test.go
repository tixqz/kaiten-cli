package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/config"
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

// -------- Retry and rate limiting tests --------

func TestRetryOn429Then200(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	body, err := client.Get("/test")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 requests (initial + 1 retry), got %d", attempt)
	}
}

func TestRetryOn500Then200(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("server error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	body, err := client.Get("/test")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 requests (initial + 1 retry), got %d", attempt)
	}
}

func TestRetryOn503Then200(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("service unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	body, err := client.Get("/test")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 requests (initial + 1 retry), got %d", attempt)
	}
}

func TestNoRetryOn404(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	_, err := client.Get("/test")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	var apiErr *APIError
	if !as(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if attempt != 1 {
		t.Fatalf("expected only 1 request (no retry on 404), got %d", attempt)
	}
}

func TestNoRetryOn401(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	_, err := client.Get("/test")
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if attempt != 1 {
		t.Fatalf("expected only 1 request (no retry on 401), got %d", attempt)
	}
}

func TestRetryExhaustionReturnsLastError(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	_, err := client.Get("/test")
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
	var apiErr *APIError
	if !as(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
	if attempt != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempt)
	}
}

func TestRetryAfterHeaderHonored(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	body, err := client.Get("/test")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 requests, got %d", attempt)
	}
}

func TestRetryAfterWithPositiveSeconds(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.SetRetryConfig(3, time.Millisecond, 10*time.Millisecond)

	start := time.Now()
	_, err := client.Get("/test")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 requests, got %d", attempt)
	}
	// Should have waited at least ~1s for Retry-After
	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected Retry-After delay of ~1s, took %v", elapsed)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"", 0},
		{"0", 0},
		{"5", 5 * time.Second},
		{"120", 120 * time.Second},
		{"abc", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseRetryAfter(tt.input)
			if got != tt.expected {
				t.Fatalf("parseRetryAfter(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRateLimiterRequestsSpaced(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	// 1000 req/s = 1ms interval
	client.SetRateLimit(1000)

	start := time.Now()
	_, err1 := client.Get("/test1")
	_, err2 := client.Get("/test2")
	elapsed := time.Since(start)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected error: %v, %v", err1, err2)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls, got %d", callCount)
	}
	// With 1ms interval, 2 requests should take at least ~1ms
	if elapsed < 500*time.Microsecond {
		t.Fatalf("requests completed suspiciously fast (%v), rate limiter may not be working", elapsed)
	}
}

func TestRateLimiterDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestServerClient(server)
	client.DisableRateLimit()

	start := time.Now()
	_, err1 := client.Get("/test1")
	_, err2 := client.Get("/test2")
	elapsed := time.Since(start)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected error: %v, %v", err1, err2)
	}
	// Should be very fast when rate limiter is disabled
	if elapsed > 100*time.Millisecond {
		t.Fatalf("requests took too long (%v) with rate limiter disabled", elapsed)
	}
}

func TestSetHTTPClient(t *testing.T) {
	client := &Client{}
	custom := &http.Client{Timeout: 5 * time.Second}
	client.SetHTTPClient(custom)
	if client.httpClient != custom {
		t.Fatal("SetHTTPClient did not set the client")
	}
}

func TestNewClientDefaults(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}
	// Minimal config; NewClient only needs cfg.BaseURL() and cfg.Token.
	// We use a test helper to avoid needing a real config, but here we just
	// verify the defaults are set on the client.
	client := NewClient(cfg)
	if client.maxRetries != defaultMaxRetries {
		t.Fatalf("maxRetries = %d, want %d", client.maxRetries, defaultMaxRetries)
	}
	if client.retryMinWait != defaultRetryMinWait {
		t.Fatalf("retryMinWait = %v, want %v", client.retryMinWait, defaultRetryMinWait)
	}
	if client.retryMaxWait != defaultRetryMaxWait {
		t.Fatalf("retryMaxWait = %v, want %v", client.retryMaxWait, defaultRetryMaxWait)
	}
	if client.rateLimitDur != time.Second/defaultRateLimit {
		t.Fatalf("rateLimitDur = %v, want %v", client.rateLimitDur, time.Second/defaultRateLimit)
	}
}

// as is a typed wrapper for errors.As to avoid import cycle with go1.23 style.
func as(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	switch t := target.(type) {
	case **APIError:
		e, ok := err.(*APIError)
		if ok {
			*t = e
		}
		return ok
	}
	return false
}
