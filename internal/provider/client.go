// HTTP client for the Savannaa API. Thin wrapper over net/http that
// stamps the right headers (x-api-key, x-region) on every request so
// resource code can stay short.
//
// Errors are returned as *APIError carrying the HTTP status + body so
// resource code can render useful diagnostics in `terraform plan/apply`.
package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Project string
	Region  string
	HTTP    *http.Client
}

type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("savannaa API returned %d: %s", e.Status, e.Body)
}

func NewClient(baseURL, apiKey, project, region string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Project: project,
		Region:  region,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

// Do issues an authenticated request. `path` is appended to BaseURL
// (start it with /). `body` is JSON-encoded; pass nil for GET/DELETE.
// `out` is JSON-decoded from the response body; pass nil to skip.
//
// Returns *APIError for non-2xx responses so callers can surface the
// upstream message in Terraform diagnostics.
func (c *Client) Do(method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("x-region", c.Region)
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("http %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return &APIError{Status: resp.StatusCode, Body: string(rawBody)}
	}
	if out != nil && len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, out); err != nil {
			return fmt.Errorf("decode response: %w (body: %s)", err, rawBody)
		}
	}
	return nil
}
