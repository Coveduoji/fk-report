package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type welcomeButton int

const (
	buttonTodo welcomeButton = iota
	buttonReport
)

const previewWindowSize = 5

const bigTitle = "F K - R E P O R T"

type welcomeModel struct {
	button  welcomeButton
	showAll bool // false = 只看未完成，true = 全部待办
	cursor  int  // 光标在"过滤后列表"里的位置
	offset  int  // 预览窗口滚动起始位置
}

func newWelcomeModel() welcomeModel {
	return welcomeModel{}
}

// filtered 返回按当前模式过滤后的待办列表，以及每一项对应到 items 里的原始下标
// （用于空格切换完成态时写回正确的原始元素）。
func (m welcomeModel) filtered(items []todoItem) ([]todoItem, []int) {
	var list []todoItem
	var idx []int
	for i, it := range items {
		if m.showAll || !it.done {
			list = append(list, it)
			idx = append(idx, i)
		}
	}
	return list, idx
}

func (m welcomeModel) clamp(n int) welcomeModel {
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.offset > m.cursor {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+previewWindowSize {
		m.offset = m.cursor - previewWindowSize + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	return m
}

// Update 处理欢迎页按键。items 和调用方（rootModel.todo.items）共享同一个底层数组，
// 空格切换完成态时直接原地修改 items[idx[cursor]].done，调用方不需要额外同步。
// 返回值：更新后的模型、待执行的命令、是否需要切换屏幕、切换到哪个屏幕。
func (m welcomeModel) Update(msg tea.KeyMsg, items []todoItem) (welcomeModel, tea.Cmd, bool, screenKind) {
	list, idx := m.filtered(items)

	switch msg.String() {
	case "left", "h":
		m.button = buttonTodo
	case "right", "l":
		m.button = buttonReport
	case "up", "k":
		m.cursor--
		m = m.clamp(len(list))
	case "down", "j":
		m.cursor++
		m = m.clamp(len(list))
	case "a":
		m.showAll = !m.showAll
		list, _ = m.filtered(items)
		m = m.clamp(len(list))
	case " ":
		if len(list) > 0 {
			items[idx[m.cursor]].done = !items[idx[m.cursor]].done
			return m, persistTodosCmd(items), false, screenWelcome
		}
	case "enter":
		if m.button == buttonReport {
			return m, nil, true, screenReport
		}
		return m, nil, true, screenTodo
	}
	return m, nil, false, screenWelcome
}

func (m welcomeModel) View(items []todoItem, width, height int) string {
	title := bannerStyle.Render(bigTitle)

	todoBtn := buttonStyle
	reportBtn := buttonStyle
	if m.button == buttonTodo {
		todoBtn = buttonActiveStyle
	} else {
		reportBtn = buttonActiveStyle
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, todoBtn.Render("TODO"), "  ", reportBtn.Render("REPORT"))

	list, _ := m.filtered(items)
	modeLabel := "未完成待办"
	if m.showAll {
		modeLabel = "全部待办"
	}

	var b strings.Builder
	b.WriteString(previewHeaderStyle.Render(fmt.Sprintf("%s（按 a 切换，共 %d 条）", modeLabel, len(list))))
	b.WriteString("\n")

	if len(list) == 0 {
		b.WriteString(itemStyle.Render("暂无待办"))
	} else {
		end := m.offset + previewWindowSize
		if end > len(list) {
			end = len(list)
		}
		for i := m.offset; i < end; i++ {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			mark := "[ ]"
			style := itemStyle
			if list[i].done {
				mark = "[x]"
				style = doneItemStyle
			}
			b.WriteString(cursor + mark + " " + style.Render(list[i].text) + "\n")
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		buttons,
		"",
		b.String(),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, body)
}

func (m welcomeModel) helpText() string {
	return "←/→ 切换按钮   ↑/↓ 滚动预览   空格 完成/取消   a 全部/未完成   enter 进入   q 退出"
}
