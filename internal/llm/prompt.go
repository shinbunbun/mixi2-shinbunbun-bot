package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
)

const systemPrompt = `あなたはGitHub活動の要約を生成するアシスタントです。
- 140〜149文字の日本語で要約する。短すぎず長すぎず、できるだけ多くの情報を詰め込む
- 「今日はこんなことをしたよ！」という語り口で、主な活動内容をハイライトする
- 全てを網羅する必要はない。特に印象的な活動をピックアップする
- PR番号やリポジトリ名は出力に含めず、何をしたかの内容を具体的に書く
- 抽象的な感想や意気込み（「チームの進捗を加速」「引き続き頑張る」など）は書かない。具体的な活動内容だけで埋める
- 絵文字を効果的に使う
- 要約テキストのみを返す（前置きや説明は不要）`

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

func BuildPrompt(events []github.Event) []Message {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		// Few-shot example
		{Role: "user", Content: "日付: 2026/03/14\n活動内容:\n[Push] CI設定を修正; READMEを更新\n\n[PR] Nix flake開発環境の追加 (マージ済み)"},
		{Role: "assistant", Content: "今日はCIのビルド設定を修正してパイプラインを安定させたよ！READMEにセットアップ手順も追記して、Nix flakeで開発環境を一発構築できるPRもマージ完了✨ これでnix developだけで依存関係が全部揃うようになった🚀 ドキュメントも環境構築もバッチリ整備できた一日💪"},
	}

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

	var userPrompt string
	if len(parts) == 0 {
		userPrompt = "今日はGitHub活動がありませんでした。"
	} else {
		userPrompt = fmt.Sprintf("日付: %s\n活動内容:\n%s", dateStr, strings.Join(parts, "\n\n"))
	}

	messages = append(messages, Message{Role: "user", Content: userPrompt})
	return messages
}
