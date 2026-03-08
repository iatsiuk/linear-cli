package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultEndpoint = "https://api.linear.app/graphql"

// Client sends GraphQL requests to the Linear API.
type Client struct {
	httpClient *http.Client
	apiKey     string
	endpoint   string
	sleep      func(context.Context, time.Duration) error
}

// Option configures a Client.
type Option func(*Client)

// WithEndpoint sets a custom API endpoint (useful for testing).
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// NewClient creates a new API client authenticated with apiKey.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		endpoint:   defaultEndpoint,
		sleep: func(ctx context.Context, d time.Duration) error {
			t := time.NewTimer(d)
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
				return nil
			}
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors GraphQLErrors   `json:"errors,omitempty"`
}

// Do executes a GraphQL query and decodes the response data into result.
// Retries on rate limit (HTTP 400 + RATELIMITED code) and server errors (5xx)
// with exponential/jitter backoff, up to 3 retries.
func (c *Client) Do(ctx context.Context, query string, variables map[string]any, result any) error {
	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		delay, err := c.do(ctx, attempt, query, variables, result)
		if err == nil {
			return nil
		}
		lastErr = err
		if delay == 0 || attempt == maxRetries {
			break
		}
		if err := c.sleep(ctx, delay); err != nil {
			return err
		}
	}
	return lastErr
}

func (c *Client) do(ctx context.Context, attempt int, query string, variables map[string]any, result any) (time.Duration, error) {
	body, err := json.Marshal(graphqlRequest{Query: query, Variables: variables})
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		return 0, fmt.Errorf("unauthorized: check your API key")
	}
	if resp.StatusCode >= 500 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return serverErrorDelay(attempt), fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var gqlResp graphqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		for _, e := range gqlResp.Errors {
			if e.Code() == CodeRateLimited {
				return rateLimitDelay(resp, attempt), gqlResp.Errors
			}
		}
		return 0, gqlResp.Errors
	}

	if result != nil && gqlResp.Data != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return 0, fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return 0, nil
}
