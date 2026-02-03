package bot

import (
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

// HandleSettings handles /settings command
func (b *Bot) HandleSettings(c telebot.Context) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return b.settingsShow(c)
	}

	switch args[1] {
	case "notify":
		return b.settingsNotify(c)
	default:
		return b.settingsShow(c)
	}
}

func (b *Bot) settingsShow(c telebot.Context) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	projectName := "없음"
	if projectID != "" {
		project, _ := b.svc.GetProject(projectID)
		if project != nil {
			projectName = project.Name
		}
	}

	isAdmin := b.cfg.IsAdmin(c.Sender().ID)
	adminStatus := "일반 사용자"
	if isAdmin {
		adminStatus = "관리자"
	}

	msg := fmt.Sprintf("⚙️ 설정\n\n"+
		"현재 프로젝트: %s\n"+
		"권한: %s\n\n"+
		"알림 설정:\n"+
		"  태스크 완료: %s\n"+
		"  태스크 실패: %s\n",
		projectName, adminStatus,
		boolToOnOff(b.cfg.NotifyOnComplete),
		boolToOnOff(b.cfg.NotifyOnFail))

	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(
			markup.Data("알림 설정", "settings:notify"),
		),
	)

	return c.Send(msg, markup)
}

func (b *Bot) settingsNotify(c telebot.Context) error {
	msg := "🔔 알림 설정\n\n" +
		"현재 설정은 서버 환경변수로 관리됩니다.\n\n" +
		"NOTIFY_ON_COMPLETE: " + boolToOnOff(b.cfg.NotifyOnComplete) + "\n" +
		"NOTIFY_ON_FAIL: " + boolToOnOff(b.cfg.NotifyOnFail)

	return c.Send(msg)
}

func (b *Bot) handleSettingsCallback(c telebot.Context, parts []string) error {
	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 요청"})
	}

	switch parts[1] {
	case "notify":
		c.Respond(&telebot.CallbackResponse{})
		return b.settingsNotify(c)
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "알 수 없는 명령"})
	}
}

func boolToOnOff(b bool) string {
	if b {
		return "켜짐"
	}
	return "꺼짐"
}
