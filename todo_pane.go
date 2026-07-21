package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type todoUIMode int

const (
	todoModeList todoUIMode = iota
	todoModeAdding
)

type todosLoadedMsg struct {
	items []todoItem
	err   error
}

type todosSavedMsg struct {
	err error
}

// persistTodosCmd 把当前待办列表整体覆盖写回 todo.md。传入前拷贝一份快照，
// 避免异步写入期间上层又修改了同一个 slice 的底层数组。
func persistTodosCmd(items []todoItem) tea.Cmd {
	snapshot := append([]todoItem(nil), items...)
	return func() tea.Msg {
		return todosSavedMsg{err: writeTodos(snapshot)}
	}
}

type todoModel struct {
	items  []todoItem
	cursor int
	mode   todoUIMode
	input  textinput.Model
	status string
}

func newTodoModel() todoModel {
	ti := textinput.New()
	ti.Placeholder = "新待办内容..."
	return todoModel{input: ti}
}

func (m todoModel) SetSize(w, h int) todoModel {
	if w > 4 {
		m.input.Width = w - 4
	}
	return m
}

func (m todoModel) Load() tea.Cmd {
	return func() tea.Msg {
		items, err := readTodos()
		return todosLoadedMsg{items: items, err: err}
	}
}

func (m todoModel) clampCursor() todoModel {
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	return m
}

func (m todoModel) Update(msg tea.Msg) (todoModel, tea.Cmd) {
	switch msg := msg.(type) {
	case todosLoadedMsg:
		if msg.err != nil {
			m.status = "读取失败: " + msg.err.Error()
			return m, nil
		}
		m.items = msg.items
		m = m.clampCursor()
		m.status = ""
		return m, nil

	case todosSavedMsg:
		if msg.err != nil {
			m.status = "保存失败: " + msg.err.Error()
		}
		return m, nil

	case tea.KeyMsg:
		if m.mode == todoModeAdding {
			switch msg.String() {
			case "enter":
				text := strings.TrimSpace(m.input.Value())
				m.input.SetValue("")
				m.mode = todoModeList
				if text == "" {
					return m, nil
				}
				m.items = append(m.items, todoItem{text: text})
				m.cursor = len(m.items) - 1
				return m, persistTodosCmd(m.items)
			case "esc":
				m.input.SetValue("")
				m.mode = todoModeList
				return m, nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if len(m.items) > 0 {
				m.items[m.cursor].done = !m.items[m.cursor].done
				return m, persistTodosCmd(m.items)
			}
		case "-":
			if len(m.items) > 0 {
				m.items = append(m.items[:m.cursor], m.items[m.cursor+1:]...)
				m = m.clampCursor()
				return m, persistTodosCmd(m.items)
			}
		case "=":
			m.mode = todoModeAdding
			return m, m.input.Focus()
		}
	}
	return m, nil
}

func (m todoModel) View() string {
	var b strings.Builder
	if len(m.items) == 0 {
		b.WriteString(itemStyle.Render("暂无待办，按 = 新增") + "\n")
	}
	for i, it := range m.items {
		cursor := "  "
		if i == m.cursor && m.mode == todoModeList {
			cursor = "> "
		}
		mark := "[ ]"
		style := itemStyle
		if it.done {
			mark = "[x]"
			style = doneItemStyle
		}
		b.WriteString(cursor + mark + " " + style.Render(it.text) + "\n")
	}

	if m.mode == todoModeAdding {
		b.WriteString("\n" + m.input.View())
	}
	if m.status != "" {
		b.WriteString("\n" + errorStyle.Render(m.status))
	}
	return b.String()
}

func (m todoModel) helpText() string {
	if m.mode == todoModeAdding {
		return "输入内容后 enter 确认   esc 取消"
	}
	return "↑/↓ 选择   空格 完成/取消   - 删除   = 新增   esc 返回欢迎页"
}
