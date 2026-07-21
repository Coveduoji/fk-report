package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screenKind int

const (
	screenWelcome screenKind = iota
	screenReport
	screenTodo
)

type focusArea int

const (
	focusMenu focusArea = iota
	focusContent
)

type paneKind int

const (
	paneDaily paneKind = iota
	paneFragment
	paneReport
)

var reportMenuItems = []struct {
	kind  paneKind
	title string
}{
	{paneDaily, "日报模式"},
	{paneFragment, "碎片模式"},
	{paneReport, "生成周报"},
}

const sidebarWidth = 16

type rootModel struct {
	width, height int
	screen        screenKind

	// screenReport 内部状态
	focus  focusArea
	cursor int
	active paneKind

	welcome  welcomeModel
	daily    dailyModel
	fragment fragmentModel
	report   reportModel
	todo     todoModel
}

func newRootModel() rootModel {
	return rootModel{
		screen:   screenWelcome,
		focus:    focusMenu,
		active:   paneDaily,
		welcome:  newWelcomeModel(),
		daily:    newDailyModel(),
		fragment: newFragmentModel(),
		report:   newReportModel(),
		todo:     newTodoModel(),
	}
}

func (m rootModel) Init() tea.Cmd {
	return m.todo.Load()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentW := m.width - sidebarWidth - 6
		contentH := m.height - 4
		m.daily = m.daily.SetSize(contentW, contentH)
		m.fragment = m.fragment.SetSize(contentW, contentH)
		m.todo = m.todo.SetSize(m.width-6, contentH)
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case screenWelcome:
			return m.updateWelcome(msg)
		case screenReport:
			return m.updateReport(msg)
		case screenTodo:
			return m.updateTodo(msg)
		}
		return m, nil

	default:
		// 非按键消息（spinner tick、异步加载/保存结果等）无条件广播给全部子面板，
		// 这样切走屏幕/面板之后，后台任务（周报生成、待办读写）依旧能推进。
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.daily, cmd = m.daily.Update(msg)
		cmds = append(cmds, cmd)
		m.fragment, cmd = m.fragment.Update(msg)
		cmds = append(cmds, cmd)
		m.report, cmd = m.report.Update(msg)
		cmds = append(cmds, cmd)
		m.todo, cmd = m.todo.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
}

func (m rootModel) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	var navigate bool
	var target screenKind
	m.welcome, cmd, navigate, target = m.welcome.Update(msg, m.todo.items)
	if navigate {
		m.screen = target
		m.focus = focusMenu
	}
	return m, cmd
}

func (m rootModel) updateReport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.focus == focusMenu {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			m.screen = screenWelcome
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(reportMenuItems)-1 {
				m.cursor++
			}
		case "enter":
			m.active = reportMenuItems[m.cursor].kind
			m.focus = focusContent
			switch m.active {
			case paneDaily:
				cmds = append(cmds, m.daily.Load())
			case paneFragment:
				cmds = append(cmds, m.fragment.Load())
			case paneReport:
				var cmd tea.Cmd
				m.report, cmd = m.report.Begin()
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
	}

	// focus == focusContent
	if msg.String() == "esc" {
		m.focus = focusMenu
		return m, nil
	}
	switch m.active {
	case paneDaily:
		var cmd tea.Cmd
		m.daily, cmd = m.daily.Update(msg)
		cmds = append(cmds, cmd)
	case paneFragment:
		var cmd tea.Cmd
		m.fragment, cmd = m.fragment.Update(msg)
		cmds = append(cmds, cmd)
	case paneReport:
		var cmd tea.Cmd
		m.report, cmd = m.report.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m rootModel) updateTodo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.todo.mode == todoModeList {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			m.screen = screenWelcome
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.todo, cmd = m.todo.Update(msg)
	return m, cmd
}

func (m rootModel) View() string {
	if m.width == 0 {
		return "加载中..."
	}

	switch m.screen {
	case screenWelcome:
		body := m.welcome.View(m.todo.items, m.width, m.height-2)
		return lipgloss.JoinVertical(lipgloss.Left, body, helpStyle.Render(m.welcome.helpText()))
	case screenTodo:
		return m.viewTodo()
	case screenReport:
		return m.viewReport()
	}
	return ""
}

func (m rootModel) viewTodo() string {
	box := contentStyle.
		Width(m.width - 4).
		Height(m.height - 4).
		BorderForeground(colorAccent).
		Render(m.todo.View())
	return lipgloss.JoinVertical(lipgloss.Left, box, helpStyle.Render(m.todo.helpText()))
}

func (m rootModel) viewReport() string {
	sidebarBorder := sidebarStyle.Width(sidebarWidth).Height(m.height - 4)
	contentBorder := contentStyle.Width(m.width - sidebarWidth - 6).Height(m.height - 4)
	if m.focus == focusMenu {
		sidebarBorder = sidebarBorder.BorderForeground(colorAccent)
	} else {
		contentBorder = contentBorder.BorderForeground(colorAccent)
	}

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarBorder.Render(m.renderSidebar()),
		contentBorder.Render(m.renderContent()),
	)
	return lipgloss.JoinVertical(lipgloss.Left, body, helpStyle.Render(m.helpText()))
}

func (m rootModel) renderSidebar() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("fk-report"))
	b.WriteString("\n\n")
	for i, item := range reportMenuItems {
		cursor := "  "
		style := itemStyle
		if i == m.cursor && m.focus == focusMenu {
			cursor = "> "
			style = selectedItemStyle
		} else if item.kind == m.active {
			style = selectedItemStyle
		}
		b.WriteString(cursor + style.Render(item.title) + "\n")
	}
	return b.String()
}

func (m rootModel) renderContent() string {
	switch m.active {
	case paneDaily:
		return m.daily.View()
	case paneFragment:
		return m.fragment.View()
	case paneReport:
		return m.report.View()
	}
	return ""
}

func (m rootModel) helpText() string {
	if m.focus == focusMenu {
		return "↑/↓ 选择   enter 进入   esc 返回欢迎页   q 退出"
	}
	switch m.active {
	case paneDaily:
		return "编辑内容   ctrl+s 保存   esc 返回菜单"
	case paneFragment:
		return "输入文字后 enter 记录一条   esc 返回菜单"
	case paneReport:
		return "r 重新生成   esc 返回菜单（生成仍在后台继续）"
	}
	return "esc 返回菜单"
}
