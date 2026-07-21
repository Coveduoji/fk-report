package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

var menuItems = []struct {
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
	focus         focusArea
	cursor        int
	active        paneKind

	daily    dailyModel
	fragment fragmentModel
	report   reportModel
}

func newRootModel() rootModel {
	return rootModel{
		focus:    focusMenu,
		active:   paneDaily,
		daily:    newDailyModel(),
		fragment: newFragmentModel(),
		report:   newReportModel(),
	}
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentW := m.width - sidebarWidth - 6
		contentH := m.height - 4
		m.daily = m.daily.SetSize(contentW, contentH)
		m.fragment = m.fragment.SetSize(contentW, contentH)
		return m, nil

	case tea.KeyMsg:
		if m.focus == focusMenu {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(menuItems)-1 {
					m.cursor++
				}
			case "enter":
				m.active = menuItems[m.cursor].kind
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

	default:
		// 非按键消息（spinner tick、异步加载/保存结果等）无条件转发给全部子面板，
		// 这样即使焦点/激活面板已经切走，后台任务（比如周报生成）依旧能推进。
		var cmd tea.Cmd
		m.daily, cmd = m.daily.Update(msg)
		cmds = append(cmds, cmd)
		m.fragment, cmd = m.fragment.Update(msg)
		cmds = append(cmds, cmd)
		m.report, cmd = m.report.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m rootModel) View() string {
	if m.width == 0 {
		return "加载中..."
	}

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
	b.WriteString(titleStyle.Render("wr 周报"))
	b.WriteString("\n\n")
	for i, item := range menuItems {
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
		return "↑/↓ 选择   enter 进入   q 退出"
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
