package bot

import (
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

// HandleProject handles /project command
func (b *Bot) HandleProject(c telebot.Context) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return b.projectHelp(c)
	}

	switch args[1] {
	case "list":
		return b.projectList(c)
	case "status":
		return b.projectStatus(c)
	case "switch":
		if len(args) < 3 {
			return c.Send("사용법: /project switch <project-id>")
		}
		return b.projectSwitch(c, args[2])
	case "info":
		return b.projectInfo(c)
	default:
		return b.projectHelp(c)
	}
}

func (b *Bot) projectHelp(c telebot.Context) error {
	return c.Send(`📁 /project 명령어

/project list    - 프로젝트 목록
/project status  - 현재 프로젝트 상태
/project switch <id> - 프로젝트 전환
/project info    - 프로젝트 정보`)
}

func (b *Bot) projectList(c telebot.Context) error {
	projects, err := b.svc.ListProjects()
	if err != nil {
		return c.Send("⚠️ 프로젝트 목록을 가져올 수 없습니다.")
	}

	if len(projects) == 0 {
		return c.Send("📁 등록된 프로젝트가 없습니다.")
	}

	currentID := b.getCurrentProject(c.Sender().ID)

	msg := "📁 프로젝트 목록\n\n"
	var buttons []telebot.Row
	markup := &telebot.ReplyMarkup{}

	for i, p := range projects {
		status, _ := b.svc.GetProjectStatus(p.ID)
		current := ""
		if p.ID == currentID {
			current = " ⭐"
		}

		msg += fmt.Sprintf("%d. %s%s\n   진행률: %.0f%% | 태스크: %d/%d\n\n",
			i+1, p.Name, current, status.Progress,
			status.CompletedTasks, status.TotalTasks)

		if p.ID != currentID {
			buttons = append(buttons, markup.Row(
				markup.Data(fmt.Sprintf("전환: %s", p.Name), fmt.Sprintf("project:switch:%s", p.ID)),
			))
		}
	}

	if len(buttons) > 0 {
		markup.Inline(buttons...)
		return c.Send(msg, markup)
	}

	return c.Send(msg)
}

func (b *Bot) projectStatus(c telebot.Context) error {
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

	// Get recent and in-progress tasks
	completedTasks, _ := b.svc.ListTasks(projectID, "done", 3)
	inProgressTasks, _ := b.svc.ListTasks(projectID, "doing", 5)

	msg := fmt.Sprintf("📊 %s 상태\n\n"+
		"진행률: %s %.0f%%\n\n"+
		"태스크 현황:\n"+
		"  ✅ 완료: %d\n"+
		"  🔄 진행중: %d\n"+
		"  ⏳ 대기: %d\n"+
		"  ❌ 실패: %d\n",
		project.Name, progressBar, status.Progress,
		status.CompletedTasks, status.InProgressTasks,
		status.PendingTasks, status.FailedTasks)

	if len(completedTasks) > 0 {
		msg += "\n최근 완료:\n"
		for _, t := range completedTasks {
			msg += fmt.Sprintf("  • %s\n", truncate(t.Title, 35))
		}
	}

	if len(inProgressTasks) > 0 {
		msg += "\n현재 진행:\n"
		for _, t := range inProgressTasks {
			msg += fmt.Sprintf("  • %s\n", truncate(t.Title, 35))
		}
	}

	return c.Send(msg)
}

func (b *Bot) projectSwitch(c telebot.Context, projectID string) error {
	project, err := b.svc.GetProject(projectID)
	if err != nil || project == nil {
		return c.Send("❌ 프로젝트를 찾을 수 없습니다: " + projectID)
	}

	// Update state
	b.state.SetCurrentProject(c.Sender().ID, projectID)

	// Also update DB
	if err := b.svc.SetCurrentProject(projectID); err != nil {
		b.logger.Error().Err(err).Msg("failed to set current project in DB")
	}

	return c.Send(fmt.Sprintf("✅ 프로젝트가 전환되었습니다.\n현재 프로젝트: %s", project.Name))
}

func (b *Bot) projectInfo(c telebot.Context) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	project, err := b.svc.GetProject(projectID)
	if err != nil || project == nil {
		return c.Send("⚠️ 프로젝트를 찾을 수 없습니다.")
	}

	msg := fmt.Sprintf("📁 프로젝트 정보\n\n"+
		"ID: %s\n"+
		"이름: %s\n"+
		"설명: %s\n"+
		"상태: %s\n"+
		"생성일: %s",
		project.ID, project.Name,
		project.Description, project.Status,
		project.CreatedAt.Format("2006-01-02"))

	return c.Send(msg)
}

func (b *Bot) handleProjectCallback(c telebot.Context, parts []string) error {
	if len(parts) < 3 {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 요청"})
	}

	switch parts[1] {
	case "switch":
		err := b.projectSwitch(c, parts[2])
		if err == nil {
			c.Respond(&telebot.CallbackResponse{Text: "프로젝트 전환됨"})
		}
		return err
	case "status":
		return b.projectStatus(c)
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "알 수 없는 명령"})
	}
}
