package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/config"
)

// Client is the Kaiten API HTTP client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client from config.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL(),
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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

// doRequest executes an HTTP request with authentication and error handling.
func (c *Client) doRequest(method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
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
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	return respBody, nil
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
	return c.doRequest(http.MethodPost, path, strings.NewReader(string(data)))
}

// Patch performs an HTTP PATCH request with a JSON body.
func (c *Client) Patch(path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return c.doRequest(http.MethodPatch, path, strings.NewReader(string(data)))
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}
