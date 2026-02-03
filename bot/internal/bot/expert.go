package bot

import (
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

// HandleExpert handles /expert command
func (b *Bot) HandleExpert(c telebot.Context) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return b.expertHelp(c)
	}

	switch args[1] {
	case "list":
		return b.expertList(c)
	case "status":
		return b.expertStatus(c)
	case "ask":
		if len(args) < 3 {
			return c.Send("사용법: /expert ask <expert-name>")
		}
		return b.expertAskStart(c, args[2])
	default:
		return b.expertHelp(c)
	}
}

func (b *Bot) expertHelp(c telebot.Context) error {
	return c.Send(`👥 /expert 명령어

/expert list       - Expert 목록
/expert status     - Expert별 태스크 현황
/expert ask <name> - Expert에게 질문`)
}

func (b *Bot) expertList(c telebot.Context) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	experts, err := b.svc.ListExperts(projectID)
	if err != nil {
		return c.Send("⚠️ Expert 목록을 가져올 수 없습니다.")
	}

	if len(experts) == 0 {
		return c.Send("👥 등록된 Expert가 없습니다.")
	}

	msg := "👥 Expert 목록\n\n"

	for _, e := range experts {
		statusEmoji := "🟢"
		if e.Status != "active" {
			statusEmoji = "⚪"
		}

		msg += fmt.Sprintf("%s %s\n", statusEmoji, e.Name)
		if e.Domain != "" {
			msg += fmt.Sprintf("   도메인: %s\n", e.Domain)
		}
		if e.Description != "" {
			msg += fmt.Sprintf("   설명: %s\n", truncate(e.Description, 40))
		}
		msg += "\n"
	}

	// Add buttons
	markup := &telebot.ReplyMarkup{}
	var buttons []telebot.Btn
	for _, e := range experts {
		if e.Status == "active" {
			buttons = append(buttons, markup.Data(e.Name, fmt.Sprintf("expert:ask:%s", e.ID)))
		}
	}
	if len(buttons) > 0 {
		markup.Inline(markup.Row(buttons...))
		return c.Send(msg, markup)
	}

	return c.Send(msg)
}

func (b *Bot) expertStatus(c telebot.Context) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	experts, err := b.svc.ListExperts(projectID)
	if err != nil {
		return c.Send("⚠️ Expert 정보를 가져올 수 없습니다.")
	}

	if len(experts) == 0 {
		return c.Send("👥 등록된 Expert가 없습니다.")
	}

	msg := "👥 Expert 현황\n\n"

	for _, e := range experts {
		statusEmoji := "🟢"
		statusText := "활성"
		if e.Status != "active" {
			statusEmoji = "⚪"
			statusText = "비활성"
		}

		msg += fmt.Sprintf("%s %s (%s)\n", statusEmoji, e.Name, statusText)
		if e.Domain != "" {
			msg += fmt.Sprintf("   담당: %s\n", e.Domain)
		}
		msg += "\n"
	}

	return c.Send(msg)
}

func (b *Bot) expertAskStart(c telebot.Context, expertName string) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	// Verify expert exists
	experts, err := b.svc.ListExperts(projectID)
	if err != nil {
		return c.Send("⚠️ Expert 정보를 가져올 수 없습니다.")
	}

	var found bool
	for _, e := range experts {
		if e.ID == expertName || e.Name == expertName {
			found = true
			expertName = e.Name
			break
		}
	}

	if !found {
		return c.Send(fmt.Sprintf("❌ Expert를 찾을 수 없습니다: %s", expertName))
	}

	userID := c.Sender().ID
	b.state.SetTempData(userID, "expert", expertName)
	b.state.SetWaiting(userID, WaitingExpertQuestion)

	return c.Send(fmt.Sprintf("%s에게 질문할 내용을 입력하세요:", expertName))
}

func (b *Bot) handleExpertQuestionInput(c telebot.Context, state *UserState) error {
	userID := c.Sender().ID
	question := strings.TrimSpace(c.Text())

	if question == "" {
		return c.Send("질문을 입력해주세요:")
	}

	expertName := b.state.GetTempData(userID, "expert")
	if expertName == nil {
		b.state.Clear(userID)
		return c.Send("⚠️ 오류가 발생했습니다. 다시 시도해주세요.")
	}

	projectID := b.getCurrentProject(userID)

	// Send as message
	content := fmt.Sprintf("[Expert 질문: %s]\n\n%s", expertName, question)
	_, err := b.svc.SendMessage(projectID, content, nil)
	if err != nil {
		b.state.Clear(userID)
		return c.Send("⚠️ 질문 전송에 실패했습니다.")
	}

	b.state.Clear(userID)
	return c.Send(fmt.Sprintf("✅ %s에게 질문이 전송되었습니다.", expertName))
}

func (b *Bot) handleExpertCallback(c telebot.Context, parts []string) error {
	if len(parts) < 3 {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 요청"})
	}

	switch parts[1] {
	case "ask":
		c.Respond(&telebot.CallbackResponse{})
		return b.expertAskStart(c, parts[2])
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "알 수 없는 명령"})
	}
}
