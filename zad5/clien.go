package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)
 
type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
 
type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
 
type Client struct {
	baseURL    string
	httpClient *http.Client
}
 
type UserAgentTransport struct {
	Base      http.RoundTripper
	UserAgent string
}
 
func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) { 
	newReq := req.Clone(req.Context())
	 
	newReq.Header.Set("User-Agent", t.UserAgent)
	 
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	
	return base.RoundTrip(newReq)
}
 
func NewClient(baseURL string) *Client { 
	transport := &UserAgentTransport{
		Base:      http.DefaultTransport,
		UserAgent: "go-http-client-pv",
	} 

	httpClient := &http.Client{
		Transport: transport,
	}
	
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}
 
func NewClientWithHTTPClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}
 
func (c *Client) GetItems(ctx context.Context) ([]Item, error) { 
	url := fmt.Sprintf("%s/items", c.baseURL)
	 
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	 
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	 
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d (expected %d)", resp.StatusCode, http.StatusOK)
	}
	 
	var items []Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return items, nil
}
 
func (c *Client) CreateItem(ctx context.Context, input CreateItemRequest) (*Item, error) { 
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	 
	url := fmt.Sprintf("%s/items", c.baseURL)
	 
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	 
	req.Header.Set("Content-Type", "application/json")
	 
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	 
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d (expected %d)", resp.StatusCode, http.StatusCreated)
	}
	 
	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &item, nil
}