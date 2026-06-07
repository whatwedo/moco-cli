// Package client is a thin HTTP client for the MOCO API v1.
//
// It passes JSON through unchanged: requests are assembled from a path, query
// parameters and an optional JSON body; the response is returned as raw bytes.
// Formatting is handled by the CLI layer.
package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client performs authenticated requests against the MOCO API.
type Client struct {
	baseURL    string
	token      string
	userAgent  string
	httpClient *http.Client
}

// New creates a Client. baseURL is the base URL including /api/v1.
func New(baseURL, token, userAgent string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Request describes a single API call. Path starts with "/" and already
// contains resolved path parameters (e.g. "/projects/123").
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Body   []byte // raw JSON body, or nil
}

// APIError describes a failed request (HTTP status >= 400).
type APIError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *APIError) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("MOCO API responded with %s: %s", e.Status, e.Body)
	}
	return fmt.Sprintf("MOCO API responded with %s", e.Status)
}

// Do executes the request and returns the response body. For HTTP status
// >= 400 it returns an *APIError.
func (c *Client) Do(ctx context.Context, req Request) ([]byte, error) {
	u := c.baseURL + req.Path
	if len(req.Query) > 0 {
		u += "?" + req.Query.Encode()
	}

	var body io.Reader
	if req.Body != nil {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u, body)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Token token="+c.token)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: respBody}
	}

	return respBody, nil
}
