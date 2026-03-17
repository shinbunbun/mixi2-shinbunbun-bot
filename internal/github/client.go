package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

	c.enrichPRTitles(ctx, filtered)
	c.enrichPushCommits(ctx, filtered)

	prTitleCount := 0
	pushCommitCount := 0
	for _, ev := range filtered {
		if ev.Type == "PullRequestEvent" && ev.Payload.PullRequest != nil && ev.Payload.PullRequest.Title != "" {
			prTitleCount++
		}
		if ev.Type == "PushEvent" && len(ev.Payload.Commits) > 0 {
			pushCommitCount++
		}
	}
	slog.Info("fetched github events",
		slog.Int("total", len(events)),
		slog.Int("filtered", len(filtered)),
		slog.Int("prs_with_title", prTitleCount),
		slog.Int("pushes_with_commits", pushCommitCount),
	)

	return filtered, nil
}

func (c *Client) enrichPRTitles(ctx context.Context, events []Event) {
	for i, ev := range events {
		if ev.Type == "PullRequestEvent" && ev.Payload.PullRequest != nil && ev.Payload.PullRequest.Title == "" && ev.Payload.PullRequest.URL != "" {
			pr, err := c.fetchPRDetail(ctx, ev.Payload.PullRequest.URL)
			if err != nil {
				slog.Warn("failed to enrich PR title",
					slog.String("url", ev.Payload.PullRequest.URL),
					slog.String("error", err.Error()),
				)
				continue
			}
			events[i].Payload.PullRequest.Title = pr.Title
		}
	}
}

func (c *Client) fetchPRDetail(ctx context.Context, url string) (*PullRequest, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("fetching PR detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decoding PR detail: %w", err)
	}
	return &pr, nil
}

func (c *Client) enrichPushCommits(ctx context.Context, events []Event) {
	for i, ev := range events {
		if ev.Type == "PushEvent" && len(ev.Payload.Commits) == 0 && ev.Payload.Before != "" && ev.Payload.Head != "" {
			compareURL := fmt.Sprintf("https://api.github.com/repos/%s/compare/%s...%s", ev.Repo.Name, ev.Payload.Before, ev.Payload.Head)
			commits, err := c.fetchCompareCommits(ctx, compareURL)
			if err != nil {
				slog.Warn("failed to enrich push commits",
					slog.String("repo", ev.Repo.Name),
					slog.String("error", err.Error()),
				)
				continue
			}
			for _, cc := range commits {
				events[i].Payload.Commits = append(events[i].Payload.Commits, Commit{
					SHA:     cc.SHA,
					Message: cc.Commit.Message,
				})
			}
			if events[i].Payload.Size == 0 {
				events[i].Payload.Size = len(events[i].Payload.Commits)
			}
		}
	}
}

func (c *Client) fetchCompareCommits(ctx context.Context, url string) ([]CompareCommit, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("fetching compare: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cr CompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("decoding compare: %w", err)
	}
	return cr.Commits, nil
}
