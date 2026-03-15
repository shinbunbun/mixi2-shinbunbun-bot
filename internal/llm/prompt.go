package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
)

const systemPrompt = `あなたはGitHub活動の要約を生成するアシスタントです。
- 必ず149文字以内の日本語で要約する
- 「今日はこんなことをしたよ！」という語り口で、主な活動内容をハイライトする
- 全てを網羅する必要はない。特に印象的な活動をピックアップする
- 絵文字を効果的に使う
- 要約テキストのみを返す（前置きや説明は不要）

出力例: 「今日はmixi2 botにLLM要約機能を追加したり、dotfilesでローカルLLM推論環境を構築したりした！Apple Silicon GPUで動くようになって快適 🚀」`

func BuildPrompt(events []github.Event) (system, user string) {
	var parts []string

	now := time.Now()
	dateStr := now.Format("2006/01/02")

	for _, ev := range events {
		switch ev.Type {
		case "PushEvent":
			var msgs []string
			for _, c := range ev.Payload.Commits {
				msgs = append(msgs, fmt.Sprintf("%q", c.Message))
			}
			parts = append(parts, fmt.Sprintf("[Push] %s\nコミット: %s", ev.Repo.Name, strings.Join(msgs, "; ")))

		case "PullRequestEvent":
			title := ""
			if ev.Payload.PullRequest != nil {
				title = ev.Payload.PullRequest.Title
			}
			parts = append(parts, fmt.Sprintf("[PR] %s #%d %q (%s)", ev.Repo.Name, ev.Payload.Number, title, ev.Payload.Action))

		case "IssuesEvent":
			title := ""
			if ev.Payload.Issue != nil {
				title = ev.Payload.Issue.Title
			}
			parts = append(parts, fmt.Sprintf("[Issue] %s #%d: %q (%s)", ev.Repo.Name, ev.Payload.Number, title, ev.Payload.Action))
		}
	}

	if len(parts) == 0 {
		return systemPrompt, "今日はGitHub活動がありませんでした。"
	}

	userPrompt := fmt.Sprintf("日付: %s\n活動内容:\n%s", dateStr, strings.Join(parts, "\n\n"))
	return systemPrompt, userPrompt
}
