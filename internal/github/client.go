package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	eventsURL = "https://api.github.com/users/shinbunbun/events/public?per_page=100"
	userAgent = "mixi2-shinbunbun-bot"
)

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

func (c *Client) doRequest(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Bearer "+c.token)

	return c.httpClient.Do(req)
}

func (c *Client) FetchRecentEvents(ctx context.Context, since time.Time) ([]Event, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, eventsURL)
	if err != nil {
		return nil, fmt.Errorf("fetching events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decoding events: %w", err)
	}

	var filtered []Event
	for _, ev := range events {
		if ev.CreatedAt.After(since) {
			filtered = append(filtered, ev)
		}
	}

	return filtered, nil
}
