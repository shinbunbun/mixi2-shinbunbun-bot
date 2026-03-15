package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/sync/errgroup"
)

const maxPatchLength = 1000

func (c *Client) FetchPRFiles(ctx context.Context, owner, repo string, number int) ([]PRFile, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", owner, repo, number)
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("fetching PR files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var files []PRFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("decoding PR files: %w", err)
	}

	for i := range files {
		files[i].Patch = trimPatch(files[i].Patch)
	}

	return files, nil
}

func (c *Client) EnrichEvents(ctx context.Context, events []Event) ([]EventWithDetails, error) {
	results := make([]EventWithDetails, len(events))
	for i, ev := range events {
		results[i] = EventWithDetails{Event: ev}
	}

	g, gctx := errgroup.WithContext(ctx)

	for i, ev := range events {
		if ev.Type != "PullRequestEvent" {
			continue
		}

		owner, repo, err := parseRepoName(ev.Repo.Name)
		if err != nil {
			slog.Warn("skipping PR diff: invalid repo name", slog.String("repo", ev.Repo.Name))
			continue
		}

		i, ev := i, ev
		g.Go(func() error {
			files, err := c.FetchPRFiles(gctx, owner, repo, ev.Payload.Number)
			if err != nil {
				slog.Warn("failed to fetch PR files, skipping",
					slog.String("repo", ev.Repo.Name),
					slog.Int("number", ev.Payload.Number),
					slog.String("error", err.Error()),
				)
				return nil
			}
			results[i].PRFiles = files
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, fmt.Errorf("enriching events: %w", err)
	}

	return results, nil
}

func parseRepoName(name string) (owner, repo string, err error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo name: %s", name)
	}
	return parts[0], parts[1], nil
}

func trimPatch(patch string) string {
	runes := []rune(patch)
	if len(runes) <= maxPatchLength {
		return patch
	}
	return string(runes[:maxPatchLength])
}
