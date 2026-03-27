package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPageSize = 250
	apiBasePath     = "/api/v2"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string
}

func New(ctx context.Context, baseURL string, apiToken string) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	if baseURL == "" {
		return nil, fmt.Errorf("baton-retool: base URL is required")
	}

	if apiToken == "" {
		return nil, fmt.Errorf("baton-retool: API token is required")
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		apiToken: apiToken,
	}

	return c, nil
}

func (c *Client) ValidateConnection(ctx context.Context) error {
	_, err := c.GetOrganization(ctx)
	return err
}

func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body interface{}) ([]byte, error) {
	u := fmt.Sprintf("%s%s%s", c.baseURL, apiBasePath, path)
	if query != nil {
		u = fmt.Sprintf("%s?%s", u, query.Encode())
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("baton-retool: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, fmt.Errorf("baton-retool: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	return respBody, nil
}

func listQuery(offset int, limit int) url.Values {
	q := url.Values{}
	if limit <= 0 {
		limit = defaultPageSize
	}
	q.Set("limit", strconv.Itoa(limit))
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	return q
}

// ListResponse is the common paginated response shape from Retool API v2.
type ListResponse[T any] struct {
	Data       []T  `json:"data"`
	HasMore    bool `json:"has_more"`
	TotalCount int  `json:"total_count"`
}

// SingleResponse wraps a single object response.
type SingleResponse[T any] struct {
	Data T `json:"data"`
}

// APIError represents an error response from the Retool API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("baton-retool: API returned status %d: %s", e.StatusCode, e.Body)
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func IsConflict(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// Pager holds pagination state for list requests.
type Pager struct {
	Token string
	Size  int
}

func (p *Pager) Parse() (int, int, error) {
	var offset int
	var err error
	limit := p.Size
	if limit <= 0 {
		limit = defaultPageSize
	}

	if p.Token != "" {
		offset, err = strconv.Atoi(p.Token)
		if err != nil {
			return 0, 0, err
		}
	}

	return offset, limit, nil
}
