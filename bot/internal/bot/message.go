package bot

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/telebot.v3"
)

// HandleMessage handles /msg command
func (b *Bot) HandleMessage(c telebot.Context) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return b.messageHelp(c)
	}

	switch args[1] {
	case "list":
		return b.messageList(c)
	case "send":
		return b.messageSendStart(c)
	case "get":
		return b.messageGet(c, args[2:])
	default:
		return b.messageHelp(c)
	}
}

func (b *Bot) messageHelp(c telebot.Context) error {
	return c.Send(`💬 /msg 명령어

/msg list    - 메시지 목록
/msg send    - 메시지 전송
/msg get <id> - 메시지 상세`)
}

func (b *Bot) messageList(c telebot.Context) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	messages, err := b.svc.ListMessages(projectID, 10)
	if err != nil {
		return c.Send("⚠️ 메시지 목록을 가져올 수 없습니다.")
	}

	if len(messages) == 0 {
		return c.Send("💬 메시지가 없습니다.")
	}

	msg := "💬 메시지 목록\n\n"

	for _, m := range messages {
		statusEmoji := "📩"
		if m.Status == "completed" {
			statusEmoji = "✅"
		} else if m.Status == "failed" {
			statusEmoji = "❌"
		} else if m.Status == "processing" {
			statusEmoji = "🔄"
		}

		content := truncate(m.Content, 30)
		timeAgo := formatTimeAgo(m.CreatedAt)

		msg += fmt.Sprintf("%s %d. %s\n     %s | %s\n",
			statusEmoji, m.ID, content, m.Status, timeAgo)
	}

	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(
			markup.Data("새 메시지", "msg:send"),
		),
	)

	return c.Send(msg, markup)
}

func (b *Bot) messageSendStart(c telebot.Context) error {
	b.state.SetWaiting(c.Sender().ID, WaitingMessageContent)
	return c.Send("메시지 내용을 입력하세요:")
}

func (b *Bot) handleMessageContentInput(c telebot.Context, state *UserState) error {
	userID := c.Sender().ID
	content := strings.TrimSpace(c.Text())

	if content == "" {
		return c.Send("내용을 입력해주세요:")
	}

	projectID := b.getCurrentProject(userID)
	if projectID == "" {
		b.state.Clear(userID)
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	// Send message
	msgID, err := b.svc.SendMessage(projectID, content, nil)
	if err != nil {
		b.state.Clear(userID)
		return c.Send("⚠️ 메시지 전송에 실패했습니다.")
	}

	b.state.Clear(userID)
	return c.Send(fmt.Sprintf("✅ 메시지가 전송되었습니다.\nMSG-%d", msgID))
}

func (b *Bot) messageGet(c telebot.Context, args []string) error {
	if len(args) < 1 {
		return c.Send("사용법: /msg get <id>")
	}

	// For now, just list messages since we don't have GetMessage by ID
	return b.messageList(c)
}

func (b *Bot) handleMessageCallback(c telebot.Context, parts []string) error {
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 요청"})
	}

	switch parts[1] {
	case "send":
		c.Respond(&telebot.CallbackResponse{})
		return b.messageSendStart(c)
	case "list":
		return b.messageList(c)
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "알 수 없는 명령"})
	}
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Minute {
		return "방금 전"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d분 전", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d시간 전", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d일 전", int(diff.Hours()/24))
	}
	return t.Format("01-02")
}
