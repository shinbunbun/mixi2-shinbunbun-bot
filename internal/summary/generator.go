package summary

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/llm"
)

// 禁止パターン: 未来志向フィラー・抽象化・メタテキスト
var forbiddenPatterns = regexp.MustCompile(
	`目指[すそせ]|していく|していこう|加速|推進中|引き続き|明日も|これからも|していきたい|していかないと|〜?\d+文字[）)]?$`,
)

func hasForbiddenPattern(text string) bool {
	return forbiddenPatterns.MatchString(text)
}

func removeForbiddenPatterns(text string) string {
	cleaned := forbiddenPatterns.ReplaceAllString(text, "")
	// 連続する句読点・空白をクリーンアップ
	cleaned = regexp.MustCompile(`[、。！]{2,}`).ReplaceAllString(cleaned, "。")
	cleaned = regexp.MustCompile(`\s{2,}`).ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(cleaned)
}

const maxPostLength = 149

type Generator struct {
	llmClient *llm.Client
}

func NewGenerator(llmClient *llm.Client) *Generator {
	return &Generator{llmClient: llmClient}
}

func (g *Generator) Generate(ctx context.Context, events []github.Event) string {
	messages := llm.BuildPrompt(events)
	result, err := g.llmClient.GenerateSummary(ctx, messages)
	if err != nil {
		slog.Warn("LLM summary failed, falling back to template", slog.String("error", err.Error()))
		return g.fallback(events)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		slog.Warn("LLM returned empty response, falling back to template")
		return g.fallback(events)
	}

	// リトライ: 短すぎる場合
	if len([]rune(result)) < 100 {
		slog.Info("LLM output too short, retrying", slog.Int("length", len([]rune(result))))
		retryMessages := append(messages,
			llm.Message{Role: "assistant", Content: result},
			llm.Message{Role: "user", Content: "短すぎます。140〜149文字になるよう、もっと詳細を追加して書き直してください。"},
		)
		retry, retryErr := g.llmClient.GenerateSummary(ctx, retryMessages)
		if retryErr == nil && strings.TrimSpace(retry) != "" {
			result = strings.TrimSpace(retry)
		}
	}

	// リトライ: 禁止パターン検出
	if hasForbiddenPattern(result) {
		matched := forbiddenPatterns.FindAllString(result, -1)
		slog.Info("LLM output contains forbidden patterns, retrying",
			slog.String("patterns", strings.Join(matched, ", ")))
		retryMessages := append(messages,
			llm.Message{Role: "assistant", Content: result},
			llm.Message{Role: "user", Content: fmt.Sprintf(
				"「%s」は使用禁止です。全て過去形で、やったことだけを書き直してください。",
				strings.Join(matched, "」「"),
			)},
		)
		retry, retryErr := g.llmClient.GenerateSummary(ctx, retryMessages)
		if retryErr == nil && strings.TrimSpace(retry) != "" {
			result = strings.TrimSpace(retry)
		}
	}

	// 最終クリーンアップ: リトライ後もまだ禁止パターンが残っていたら削除
	if hasForbiddenPattern(result) {
		slog.Warn("forbidden patterns still present after retry, removing them")
		result = removeForbiddenPatterns(result)
	}

	return truncate(result)
}

func (g *Generator) fallback(events []github.Event) string {
	return truncate(generateTemplate(events))
}

func truncate(text string) string {
	runes := []rune(text)
	if len(runes) <= maxPostLength {
		return text
	}
	return string(runes[:maxPostLength-1]) + "…"
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

	return strings.Join(parts, "\n")
}
