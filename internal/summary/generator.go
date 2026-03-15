package summary

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/llm"
)

const maxPostLength = 149

type Generator struct {
	llmClient *llm.Client
}

func NewGenerator(llmClient *llm.Client) *Generator {
	return &Generator{llmClient: llmClient}
}

func (g *Generator) Generate(ctx context.Context, events []github.EventWithDetails) string {
	systemPrompt, userPrompt := llm.BuildPrompt(events)
	result, err := g.llmClient.GenerateSummary(ctx, systemPrompt, userPrompt)
	if err != nil {
		slog.Warn("LLM summary failed, falling back to template", slog.String("error", err.Error()))
		return g.fallback(events)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		slog.Warn("LLM returned empty response, falling back to template")
		return g.fallback(events)
	}

	return truncate(result)
}

func (g *Generator) fallback(enriched []github.EventWithDetails) string {
	var events []github.Event
	for _, ed := range enriched {
		events = append(events, ed.Event)
	}
	return generateTemplate(events)
}

type repoStats struct {
	commits int
}

type prInfo struct {
	repo   string
	number int
	action string
}

type issueInfo struct {
	repo   string
	number int
	title  string
	action string
}

func generateTemplate(events []github.Event) string {
	now := time.Now()
	dateStr := now.Format("2006/01/02")

	if len(events) == 0 {
		return fmt.Sprintf("📊 shinbunbun の GitHub 活動レポート (%s)\n\n今日は活動がありませんでした。ゆっくり休んでね！", dateStr)
	}

	pushRepos := make(map[string]*repoStats)
	totalCommits := 0
	var prs []prInfo
	var issues []issueInfo

	for _, ev := range events {
		switch ev.Type {
		case "PushEvent":
			size := ev.Payload.Size
			if size == 0 {
				size = len(ev.Payload.Commits)
			}
			if _, ok := pushRepos[ev.Repo.Name]; !ok {
				pushRepos[ev.Repo.Name] = &repoStats{}
			}
			pushRepos[ev.Repo.Name].commits += size
			totalCommits += size

		case "PullRequestEvent":
			prs = append(prs, prInfo{
				repo:   ev.Repo.Name,
				number: ev.Payload.Number,
				action: ev.Payload.Action,
			})

		case "IssuesEvent":
			if ev.Payload.Issue != nil {
				issues = append(issues, issueInfo{
					repo:   ev.Repo.Name,
					number: ev.Payload.Issue.Number,
					title:  ev.Payload.Issue.Title,
					action: ev.Payload.Action,
				})
			}
		}
	}

	var parts []string
	header := fmt.Sprintf("📊 shinbunbun の GitHub 活動レポート (%s)", dateStr)
	parts = append(parts, header)

	if totalCommits > 0 {
		pushLine := fmt.Sprintf("🔨 Push: %d commits", totalCommits)
		parts = append(parts, pushLine)
		for repo, stats := range pushRepos {
			parts = append(parts, fmt.Sprintf("  - %s: %d commits", repo, stats.commits))
		}
	}

	if len(prs) > 0 {
		parts = append(parts, fmt.Sprintf("🔀 PR: %d件", len(prs)))
		for _, pr := range prs {
			parts = append(parts, fmt.Sprintf("  - %s #%d (%s)", pr.repo, pr.number, pr.action))
		}
	}

	if len(issues) > 0 {
		parts = append(parts, fmt.Sprintf("📝 Issue: %d件", len(issues)))
		for _, issue := range issues {
			parts = append(parts, fmt.Sprintf("  - %s #%d: \"%s\" (%s)", issue.repo, issue.number, issue.title, issue.action))
		}
	}

	parts = append(parts, "今日も開発お疲れ様でした！")

	result := strings.Join(parts, "\n")
	return truncate(result)
}

func truncate(s string) string {
	runes := []rune(s)
	if len(runes) <= maxPostLength {
		return s
	}
	return string(runes[:maxPostLength-1]) + "…"
}
