package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
)

const systemPrompt = `あなたはGitHub活動の要約を生成するアシスタントです。
- 必ず140〜149文字の日本語で要約する。120文字未満は絶対NG。文字数が足りなければ詳細を追加して埋める
- 「今日はこんなことをしたよ！」という語り口で、主な活動内容をハイライトする
- 全てを網羅する必要はない。特に印象的な活動をピックアップする
- PR番号やリポジトリ名は出力に含めず、何をしたかの内容を具体的に書く
- 絵文字を効果的に使う
- 要約テキストのみを返す（前置きや説明は不要）

出力例: 「今日はmixi2のbotにLLM要約機能を追加して、GitHubの活動を自動で要約投稿できるようにしたよ！さらにdotfilesリポジトリでローカルLLM推論環境も構築して、Apple Silicon GPUでの推論がサクサク動くようになった✨ Nixの開発環境設定も整えてかなり快適になった🚀」`

func translateAction(action string) string {
	switch action {
	case "opened":
		return "作成"
	case "closed":
		return "マージ済み"
	case "reopened":
		return "再オープン"
	default:
		return action
	}
}

func BuildPrompt(events []github.Event) (system, user string) {
	var parts []string

	now := time.Now()
	dateStr := now.Format("2006/01/02")

	for _, ev := range events {
		switch ev.Type {
		case "PushEvent":
			var msgs []string
			for _, c := range ev.Payload.Commits {
				msgs = append(msgs, c.Message)
			}
			parts = append(parts, fmt.Sprintf("[Push] %s", strings.Join(msgs, "; ")))

		case "PullRequestEvent":
			title := ""
			if ev.Payload.PullRequest != nil {
				title = ev.Payload.PullRequest.Title
			}
			action := translateAction(ev.Payload.Action)
			parts = append(parts, fmt.Sprintf("[PR] %s (%s)", title, action))

		case "IssuesEvent":
			title := ""
			if ev.Payload.Issue != nil {
				title = ev.Payload.Issue.Title
			}
			action := translateAction(ev.Payload.Action)
			parts = append(parts, fmt.Sprintf("[Issue] %s (%s)", title, action))
		}
	}

	if len(parts) == 0 {
		return systemPrompt, "今日はGitHub活動がありませんでした。"
	}

	userPrompt := fmt.Sprintf("日付: %s\n活動内容:\n%s", dateStr, strings.Join(parts, "\n\n"))
	return systemPrompt, userPrompt
}
