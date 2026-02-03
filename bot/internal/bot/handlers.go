package bot

import (
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

// HandleStart handles /start command
func (b *Bot) HandleStart(c telebot.Context) error {
	userID := c.Sender().ID

	// Get current project
	project, err := b.svc.GetCurrentProject()
	if err != nil {
		return c.Send("⚠️ 오류가 발생했습니다.")
	}

	var msg string
	if project == nil {
		msg = fmt.Sprintf("👋 Claribot에 오신 것을 환영합니다!\n\n"+
			"현재 설정된 프로젝트가 없습니다.\n"+
			"/project list로 프로젝트를 확인하세요.\n\n"+
			"명령어 도움말: /help")
	} else {
		// Cache current project
		b.state.SetCurrentProject(userID, project.ID)

		status, _ := b.svc.GetProjectStatus(project.ID)
		progressBar := makeProgressBar(status.Progress, 15)

		msg = fmt.Sprintf("👋 Claribot에 오신 것을 환영합니다!\n\n"+
			"현재 프로젝트: %s\n"+
			"진행률: %s %.0f%%\n\n"+
			"명령어 도움말: /help",
			project.Name, progressBar, status.Progress)
	}

	return c.Send(msg)
}

// HandleHelp handles /help command
func (b *Bot) HandleHelp(c telebot.Context) error {
	args := strings.Fields(c.Text())

	if len(args) > 1 {
		return b.helpDetail(c, args[1])
	}

	msg := `📖 Claribot 명령어

프로젝트
  /project list    - 프로젝트 목록
  /project status  - 현재 프로젝트 상태
  /project switch  - 프로젝트 전환

태스크
  /task list       - 태스크 목록
  /task add        - 태스크 추가
  /task start <id> - 태스크 시작
  /task done <id>  - 태스크 완료

메시지
  /msg send        - 메시지 전송
  /msg list        - 메시지 목록

Expert
  /expert list     - Expert 목록
  /expert ask      - Expert에게 질문

상태
  /status          - 전체 상태 요약

설정
  /settings        - 설정 확인/변경`

	return c.Send(msg)
}

// helpDetail shows detailed help for a command
func (b *Bot) helpDetail(c telebot.Context, command string) error {
	var msg string

	switch command {
	case "project":
		msg = `📁 /project 명령어

/project list
  모든 프로젝트 목록 표시

/project status
  현재 프로젝트의 상세 상태

/project switch <id>
  다른 프로젝트로 전환

/project info
  현재 프로젝트 정보`

	case "task":
		msg = `📋 /task 명령어

/task list [상태]
  태스크 목록 (pending/doing/done)

/task add
  새 태스크 추가 (대화형)

/task get <id>
  태스크 상세 정보

/task start <id>
  태스크 시작

/task done <id>
  태스크 완료

/task fail <id> [이유]
  태스크 실패 처리`

	case "msg", "message":
		msg = `💬 /msg 명령어

/msg list
  최근 메시지 목록

/msg send
  새 메시지 전송 (대화형)

/msg get <id>
  메시지 상세 내용`

	case "expert":
		msg = `👥 /expert 명령어

/expert list
  Expert 목록

/expert status
  Expert별 태스크 현황

/expert ask <name>
  Expert에게 질문`

	default:
		msg = fmt.Sprintf("❓ '%s' 명령어에 대한 도움말이 없습니다.", command)
	}

	return c.Send(msg)
}

// HandleStatus handles /status command
func (b *Bot) HandleStatus(c telebot.Context) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.\n/project list로 확인하세요.")
	}

	project, err := b.svc.GetProject(projectID)
	if err != nil || project == nil {
		return c.Send("⚠️ 프로젝트를 찾을 수 없습니다.")
	}

	status, err := b.svc.GetProjectStatus(projectID)
	if err != nil {
		return c.Send("⚠️ 상태를 가져올 수 없습니다.")
	}

	progressBar := makeProgressBar(status.Progress, 15)

	// Get recent tasks
	recentTasks, _ := b.svc.ListTasks(projectID, "done", 3)
	inProgressTasks, _ := b.svc.ListTasks(projectID, "doing", 3)

	msg := fmt.Sprintf("📊 Claritask 대시보드\n\n"+
		"프로젝트: %s\n"+
		"진행률: %s %.0f%%\n\n"+
		"태스크 요약:\n"+
		"  ✅ 완료: %d\n"+
		"  🔄 진행: %d\n"+
		"  ⏳ 대기: %d\n"+
		"  ❌ 실패: %d\n",
		project.Name, progressBar, status.Progress,
		status.CompletedTasks, status.InProgressTasks,
		status.PendingTasks, status.FailedTasks)

	if len(inProgressTasks) > 0 {
		msg += "\n현재 진행 중:\n"
		for _, t := range inProgressTasks {
			msg += fmt.Sprintf("  🔄 %d. %s\n", t.ID, truncate(t.Title, 30))
		}
	}

	if len(recentTasks) > 0 {
		msg += "\n최근 완료:\n"
		for _, t := range recentTasks {
			msg += fmt.Sprintf("  ✅ %d. %s\n", t.ID, truncate(t.Title, 30))
		}
	}

	// Add inline buttons
	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(
			markup.Data("프로젝트", "project:status"),
			markup.Data("태스크", "task:list"),
			markup.Data("메시지", "msg:list"),
		),
	)

	return c.Send(msg, markup)
}

// HandleText handles regular text messages (conversational input)
func (b *Bot) HandleText(c telebot.Context) error {
	userID := c.Sender().ID
	state := b.state.Get(userID)

	switch state.WaitingFor {
	case WaitingTaskTitle:
		return b.handleTaskTitleInput(c, state)
	case WaitingTaskDescription:
		return b.handleTaskDescriptionInput(c, state)
	case WaitingMessageContent:
		return b.handleMessageContentInput(c, state)
	case WaitingExpertQuestion:
		return b.handleExpertQuestionInput(c, state)
	case WaitingNone:
		return c.Send("❓ 알 수 없는 입력입니다. /help를 확인하세요.")
	default:
		b.state.Clear(userID)
		return c.Send("❓ 입력이 취소되었습니다. /help를 확인하세요.")
	}
}

// HandleCallback handles inline button callbacks
func (b *Bot) HandleCallback(c telebot.Context) error {
	data := c.Callback().Data
	parts := strings.Split(data, ":")

	if len(parts) < 2 {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 요청"})
	}

	switch parts[0] {
	case "project":
		return b.handleProjectCallback(c, parts)
	case "task":
		return b.handleTaskCallback(c, parts)
	case "msg":
		return b.handleMessageCallback(c, parts)
	case "expert":
		return b.handleExpertCallback(c, parts)
	case "settings":
		return b.handleSettingsCallback(c, parts)
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "알 수 없는 명령"})
	}
}

// makeProgressBar creates a progress bar string
func makeProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
