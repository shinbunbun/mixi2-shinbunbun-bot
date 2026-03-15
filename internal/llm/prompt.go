package llm

import (
	"fmt"
	"strings"

	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
)

const systemPrompt = `あなたはGitHub活動の要約を生成するアシスタントです。
以下のルールに従ってください:
- 149文字以内の日本語で要約を生成する
- 絵文字を使ってもOK
- 要約テキストのみを返す（説明や前置きは不要）`

func BuildPrompt(events []github.EventWithDetails) (system, user string) {
	var parts []string

	for _, ed := range events {
		ev := ed.Event
		switch ev.Type {
		case "PushEvent":
			var msgs []string
			for _, c := range ev.Payload.Commits {
				msgs = append(msgs, c.Message)
			}
			parts = append(parts, fmt.Sprintf("[Push] %s\nコミット: %s", ev.Repo.Name, strings.Join(msgs, "; ")))

		case "PullRequestEvent":
			section := fmt.Sprintf("[PR] %s #%d (%s)", ev.Repo.Name, ev.Payload.Number, ev.Payload.Action)
			if len(ed.PRFiles) > 0 {
				var files []string
				for _, f := range ed.PRFiles {
					entry := fmt.Sprintf("  %s %s (+%d/-%d)", f.Status, f.Filename, f.Additions, f.Deletions)
					if f.Patch != "" {
						entry += "\n" + f.Patch
					}
					files = append(files, entry)
				}
				section += "\n変更ファイル:\n" + strings.Join(files, "\n")
			}
			parts = append(parts, section)

		case "IssuesEvent":
			title := ""
			if ev.Payload.Issue != nil {
				title = ev.Payload.Issue.Title
			}
			parts = append(parts, fmt.Sprintf("[Issue] %s #%d: %s (%s)", ev.Repo.Name, ev.Payload.Number, title, ev.Payload.Action))
		}
	}

	if len(parts) == 0 {
		return systemPrompt, "今日はGitHub活動がありませんでした。"
	}

	return systemPrompt, strings.Join(parts, "\n\n")
}
