package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/config"
)

// Default retry and rate limit settings.
const (
	defaultMaxRetries   = 3
	defaultRetryMinWait = 500 * time.Millisecond
	defaultRetryMaxWait = 5 * time.Second
	defaultRateLimit    = 5 // requests per second
)

// Client is the Kaiten API HTTP client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	// retry settings
	maxRetries   int
	retryMinWait time.Duration
	retryMaxWait time.Duration

	// rate limiting
	rateLimitDur time.Duration
	lastReqTime  time.Time
	mu           sync.Mutex
}

// NewClient creates a new API client from config with default retry and rate limit settings.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL(),
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries:   defaultMaxRetries,
		retryMinWait: defaultRetryMinWait,
		retryMaxWait: defaultRetryMaxWait,
		rateLimitDur: time.Second / defaultRateLimit,
	}
}

// SetRetryConfig sets the retry configuration.
// maxAttempts is the total number of attempts (including the initial request).
// Set to 1 to disable retries.
func (c *Client) SetRetryConfig(maxAttempts int, minWait, maxWait time.Duration) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	c.maxRetries = maxAttempts
	c.retryMinWait = minWait
	c.retryMaxWait = maxWait
}

// SetRateLimit sets the rate limit to the given number of requests per second.
// Set to 0 to disable rate limiting.
func (c *Client) SetRateLimit(requestsPerSecond float64) {
	if requestsPerSecond <= 0 {
		c.rateLimitDur = 0
		return
	}
	c.rateLimitDur = time.Duration(float64(time.Second) / requestsPerSecond)
}

// DisableRateLimit disables rate limiting.
func (c *Client) DisableRateLimit() {
	c.rateLimitDur = 0
}

// SetHTTPClient sets the underlying HTTP client.
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// SetBaseURL sets the API base URL (for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetToken sets the API token (for testing).
func (c *Client) SetToken(token string) {
	c.token = token
}

// APIError represents an error response from the Kaiten API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	switch e.StatusCode {
	case 401:
		return fmt.Sprintf("authentication failed (HTTP 401): invalid or missing token. %s", e.Body)
	case 402:
		return fmt.Sprintf("feature unavailable (HTTP 402): blocked by tariff/plan. %s", e.Body)
	case 403:
		return fmt.Sprintf("forbidden (HTTP 403): insufficient permissions. %s", e.Body)
	case 404:
		return fmt.Sprintf("not found (HTTP 404): resource does not exist. %s", e.Body)
	case 429:
		return fmt.Sprintf("rate limit exceeded (HTTP 429): too many requests. %s", e.Body)
	default:
		return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Body)
	}
}

// doRequest executes an HTTP request with retry and rate limiting.
func (c *Client) doRequest(method, path string, body []byte) ([]byte, error) {
	var lastErr error
	maxAttempts := c.maxRetries
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Backoff before retry
			c.backoffSleep(attempt)
		}

		// Rate limiting applies to every HTTP attempt, including retries.
		c.rateLimit()

		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		url := c.baseURL + path
		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/json")
		if body != nil && (method == http.MethodPost || method == http.MethodPatch) {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading response: %w", readErr)
		}

		if resp.StatusCode < 400 {
			return respBody, nil
		}

		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}

		if !isRetryable(resp.StatusCode) {
			return nil, apiErr
		}

		lastErr = apiErr

		if attempt >= maxAttempts-1 {
			return nil, apiErr
		}

		// For 429, check Retry-After header
		if resp.StatusCode == 429 {
			if retryAfter := parseRetryAfter(resp.Header.Get("Retry-After")); retryAfter > 0 {
				time.Sleep(retryAfter)
				continue
			}
		}
	}

	return nil, lastErr
}

// backoffSleep sleeps for the computed backoff duration for the given attempt.
func (c *Client) backoffSleep(attempt int) {
	wait := c.backoffDuration(attempt)
	time.Sleep(wait)
}

// backoffDuration returns the exponential backoff duration for the given attempt (1-based).
func (c *Client) backoffDuration(attempt int) time.Duration {
	wait := c.retryMinWait * time.Duration(math.Pow(2, float64(attempt-1)))
	if wait > c.retryMaxWait {
		wait = c.retryMaxWait
	}
	return wait
}

// rateLimit enforces the rate limit by sleeping if necessary.
func (c *Client) rateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rateLimitDur == 0 {
		return
	}
	elapsed := time.Since(c.lastReqTime)
	if elapsed < c.rateLimitDur {
		time.Sleep(c.rateLimitDur - elapsed)
	}
	c.lastReqTime = time.Now()
}

// isRetryable returns true if the status code should trigger a retry.
func isRetryable(status int) bool {
	if status == 429 {
		return true
	}
	return status >= 500 && status < 600
}

// parseRetryAfter parses the Retry-After header value and returns the duration to wait.
// Supports seconds (integer) and HTTP-date formats.
func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 0
	}
	// Try seconds first
	seconds, err := strconv.Atoi(val)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}
	// Try HTTP-date
	t, err := time.Parse(time.RFC1123, val)
	if err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

// Get performs an HTTP GET request.
func (c *Client) Get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post performs an HTTP POST request with a JSON body.
func (c *Client) Post(path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return c.doRequest(http.MethodPost, path, data)
}

// Patch performs an HTTP PATCH request with a JSON body.
func (c *Client) Patch(path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return c.doRequest(http.MethodPatch, path, data)
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}
