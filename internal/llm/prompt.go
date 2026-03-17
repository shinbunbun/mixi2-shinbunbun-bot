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
- 入力のPRタイトルやコミットメッセージから具体的なキーワードを拾って要約に反映する。「新機能のPR」「既存の改善」のような抽象化はしない
- 「〜を目指す」「〜を強化」「〜していく」「〜を加速」など未来志向や意気込みは禁止。過去形で実際にやったことだけ書く
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
		// Few-shot example 1: Push+PR混合パターン
		{Role: "user", Content: "日付: 2026/03/14\n活動内容:\n[Push] GitHub ActionsのGoバージョンを1.24に更新; READMEにローカル開発手順を追記\n\n[PR] Nix flakeによる開発環境の統一 (マージ済み)\n\n[PR] LLM要約の文字数制御を改善 (作成)"},
		{Role: "assistant", Content: "GitHub ActionsのGoを1.24に上げてCIを最新化、READMEにローカル開発手順も追記したよ！Nix flakeで開発環境を統一するPRをマージして、LLM要約の文字数制御を改善するPRも新規作成✨ CI・ドキュメント・LLMと幅広く手を動かした一日🚀💪"},
		// Few-shot example 2: PR中心・取捨選択パターン
		{Role: "user", Content: "日付: 2026/03/15\n活動内容:\n[Push] テストのタイムアウトを30秒に延長\n\n[PR] Webhook受信時のエラーハンドリング強化 (マージ済み)\n\n[PR] mixi2投稿APIのリトライ処理追加 (マージ済み)\n\n[PR] dependabot: go.uber.org/zapを1.27.0に更新 (マージ済み)"},
		{Role: "assistant", Content: "Webhookのエラーハンドリングを強化して、mixi2投稿APIにリトライ処理も追加したよ！どちらもマージ完了で耐障害性がアップ🛡️ テストのタイムアウト調整やzapのバージョン更新など細かいメンテもこなした一日🔧✨"},
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
