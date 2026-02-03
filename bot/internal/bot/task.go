package bot

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"
)

// HandleTask handles /task command
func (b *Bot) HandleTask(c telebot.Context) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return b.taskHelp(c)
	}

	switch args[1] {
	case "list":
		return b.taskList(c, args[2:])
	case "add":
		return b.taskAddStart(c)
	case "get":
		return b.taskGet(c, args[2:])
	case "start":
		return b.taskStart(c, args[2:])
	case "done":
		return b.taskDone(c, args[2:])
	case "fail":
		return b.taskFail(c, args[2:])
	default:
		return b.taskHelp(c)
	}
}

func (b *Bot) taskHelp(c telebot.Context) error {
	return c.Send(`📋 /task 명령어

/task list [상태]  - 태스크 목록
/task add          - 태스크 추가
/task get <id>     - 태스크 상세
/task start <id>   - 태스크 시작
/task done <id>    - 태스크 완료
/task fail <id>    - 태스크 실패`)
}

func (b *Bot) taskList(c telebot.Context, args []string) error {
	projectID := b.getCurrentProject(c.Sender().ID)
	if projectID == "" {
		return c.Send("❌ 현재 프로젝트가 설정되지 않았습니다.")
	}

	status := ""
	if len(args) > 0 {
		status = args[0]
	}

	tasks, err := b.svc.ListTasks(projectID, status, 20)
	if err != nil {
		return c.Send("⚠️ 태스크 목록을 가져올 수 없습니다.")
	}

	if len(tasks) == 0 {
		return c.Send("📋 태스크가 없습니다.")
	}

	project, _ := b.svc.GetProject(projectID)
	projectName := projectID
	if project != nil {
		projectName = project.Name
	}

	msg := fmt.Sprintf("📋 태스크 목록 (%s)\n\n", projectName)

	// Group by status
	var doing, pending, done []string

	for _, t := range tasks {
		line := fmt.Sprintf("%d. %s", t.ID, truncate(t.Title, 30))

		switch t.Status {
		case "doing":
			doing = append(doing, "🔄 "+line)
		case "pending":
			pending = append(pending, "⏳ "+line)
		case "done":
			done = append(done, "✅ "+line)
		}
	}

	if len(doing) > 0 {
		msg += "진행중:\n"
		for _, l := range doing {
			msg += "  " + l + "\n"
		}
		msg += "\n"
	}

	if len(pending) > 0 {
		msg += "대기중:\n"
		for _, l := range pending[:min(5, len(pending))] {
			msg += "  " + l + "\n"
		}
		if len(pending) > 5 {
			msg += fmt.Sprintf("  ... 외 %d개\n", len(pending)-5)
		}
		msg += "\n"
	}

	if len(done) > 0 && status == "done" {
		msg += "완료:\n"
		for _, l := range done[:min(5, len(done))] {
			msg += "  " + l + "\n"
		}
	}

	return c.Send(msg)
}

func (b *Bot) taskGet(c telebot.Context, args []string) error {
	if len(args) < 1 {
		return c.Send("사용법: /task get <id>")
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("❌ 잘못된 태스크 ID입니다.")
	}

	task, err := b.svc.GetTask(id)
	if err != nil || task == nil {
		return c.Send("❌ 태스크를 찾을 수 없습니다.")
	}

	statusEmoji := map[string]string{
		"pending": "⏳ 대기",
		"doing":   "🔄 진행중",
		"done":    "✅ 완료",
		"failed":  "❌ 실패",
	}

	featureName := b.svc.GetFeatureName(task.FeatureID)

	msg := fmt.Sprintf("📌 TASK-%d: %s\n\n"+
		"상태: %s\n"+
		"Feature: %s\n",
		task.ID, task.Title,
		statusEmoji[task.Status], featureName)

	if task.TargetFile != "" {
		msg += fmt.Sprintf("대상 파일: %s\n", task.TargetFile)
	}

	if task.Content != "" {
		content := task.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		msg += fmt.Sprintf("\n설명:\n%s\n", content)
	}

	msg += fmt.Sprintf("\n생성일: %s", task.CreatedAt.Format("2006-01-02"))

	if task.StartedAt != nil {
		msg += fmt.Sprintf("\n시작일: %s", task.StartedAt.Format("2006-01-02"))
	}

	// Add action buttons based on status
	markup := &telebot.ReplyMarkup{}
	var buttons []telebot.Btn

	switch task.Status {
	case "pending":
		buttons = append(buttons, markup.Data("시작", fmt.Sprintf("task:start:%d", task.ID)))
	case "doing":
		buttons = append(buttons, markup.Data("완료", fmt.Sprintf("task:done:%d", task.ID)))
		buttons = append(buttons, markup.Data("실패", fmt.Sprintf("task:fail:%d", task.ID)))
	}

	if len(buttons) > 0 {
		markup.Inline(markup.Row(buttons...))
		return c.Send(msg, markup)
	}

	return c.Send(msg)
}

func (b *Bot) taskAddStart(c telebot.Context) error {
	b.state.SetWaiting(c.Sender().ID, WaitingTaskTitle)
	return c.Send("태스크 제목을 입력하세요:")
}

func (b *Bot) handleTaskTitleInput(c telebot.Context, state *UserState) error {
	userID := c.Sender().ID
	title := strings.TrimSpace(c.Text())

	if title == "" {
		return c.Send("제목을 입력해주세요:")
	}

	b.state.SetTempData(userID, "title", title)
	b.state.SetWaiting(userID, WaitingTaskDescription)

	return c.Send("설명을 입력하세요 (스킵하려면 /skip):")
}

func (b *Bot) handleTaskDescriptionInput(c telebot.Context, state *UserState) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())

	description := ""
	if text != "/skip" {
		description = text
	}

	title := b.state.GetTempData(userID, "title").(string)

	// For now, just show confirmation (actual task creation would need feature selection)
	b.state.Clear(userID)

	return c.Send(fmt.Sprintf("✅ 태스크 정보:\n\n"+
		"제목: %s\n"+
		"설명: %s\n\n"+
		"(실제 생성은 CLI에서 진행해주세요: clari task push)",
		title, description))
}

func (b *Bot) taskStart(c telebot.Context, args []string) error {
	if len(args) < 1 {
		return c.Send("사용법: /task start <id>")
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("❌ 잘못된 태스크 ID입니다.")
	}

	task, err := b.svc.GetTask(id)
	if err != nil || task == nil {
		return c.Send("❌ 태스크를 찾을 수 없습니다.")
	}

	if task.Status != "pending" {
		return c.Send(fmt.Sprintf("❌ 시작할 수 없는 상태입니다: %s", task.Status))
	}

	if err := b.svc.UpdateTaskStatus(id, "doing"); err != nil {
		return c.Send("⚠️ 상태 변경에 실패했습니다.")
	}

	return c.Send(fmt.Sprintf("🔄 TASK-%d가 시작되었습니다.\n%s", id, task.Title))
}

func (b *Bot) taskDone(c telebot.Context, args []string) error {
	if len(args) < 1 {
		return c.Send("사용법: /task done <id>")
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("❌ 잘못된 태스크 ID입니다.")
	}

	task, err := b.svc.GetTask(id)
	if err != nil || task == nil {
		return c.Send("❌ 태스크를 찾을 수 없습니다.")
	}

	if task.Status != "doing" {
		return c.Send(fmt.Sprintf("❌ 완료할 수 없는 상태입니다: %s", task.Status))
	}

	if err := b.svc.UpdateTaskStatus(id, "done"); err != nil {
		return c.Send("⚠️ 상태 변경에 실패했습니다.")
	}

	return c.Send(fmt.Sprintf("✅ TASK-%d가 완료되었습니다.\n%s", id, task.Title))
}

func (b *Bot) taskFail(c telebot.Context, args []string) error {
	if len(args) < 1 {
		return c.Send("사용법: /task fail <id> [이유]")
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("❌ 잘못된 태스크 ID입니다.")
	}

	task, err := b.svc.GetTask(id)
	if err != nil || task == nil {
		return c.Send("❌ 태스크를 찾을 수 없습니다.")
	}

	if err := b.svc.UpdateTaskStatus(id, "failed"); err != nil {
		return c.Send("⚠️ 상태 변경에 실패했습니다.")
	}

	return c.Send(fmt.Sprintf("❌ TASK-%d가 실패 처리되었습니다.\n%s", id, task.Title))
}

func (b *Bot) handleTaskCallback(c telebot.Context, parts []string) error {
	if len(parts) < 3 {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 요청"})
	}

	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "잘못된 ID"})
	}

	switch parts[1] {
	case "start":
		err := b.taskStart(c, []string{parts[2]})
		if err == nil {
			c.Respond(&telebot.CallbackResponse{Text: "태스크 시작됨"})
		}
		return err
	case "done":
		err := b.taskDone(c, []string{parts[2]})
		if err == nil {
			c.Respond(&telebot.CallbackResponse{Text: "태스크 완료됨"})
		}
		return err
	case "fail":
		err := b.taskFail(c, []string{parts[2]})
		if err == nil {
			c.Respond(&telebot.CallbackResponse{Text: "태스크 실패 처리됨"})
		}
		return err
	case "get":
		return b.taskGet(c, []string{fmt.Sprintf("%d", id)})
	case "list":
		return b.taskList(c, nil)
	default:
		return c.Respond(&telebot.CallbackResponse{Text: "알 수 없는 명령"})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
