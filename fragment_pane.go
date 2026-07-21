package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type fragmentLoadedMsg struct {
	lines []string
	err   error
}

type fragmentAppendedMsg struct {
	line string
	err  error
}

type fragmentModel struct {
	input  textinput.Model
	vp     viewport.Model
	lines  []string
	status string
}

func newFragmentModel() fragmentModel {
	ti := textinput.New()
	ti.Placeholder = "输入一条记录，回车追加..."
	vp := viewport.New(10, 10)
	return fragmentModel{input: ti, vp: vp}
}

func (m fragmentModel) SetSize(w, h int) fragmentModel {
	if w > 4 {
		m.input.Width = w - 4
		m.vp.Width = w - 2
	}
	if h > 6 {
		m.vp.Height = h - 4
	}
	return m
}

func (m fragmentModel) Load() tea.Cmd {
	return func() tea.Msg {
		content, err := readDailyFile()
		if err != nil {
			return fragmentLoadedMsg{err: err}
		}
		return fragmentLoadedMsg{lines: fragmentLines(content)}
	}
}

func (m fragmentModel) refreshViewport() fragmentModel {
	m.vp.SetContent(strings.Join(m.lines, "\n"))
	m.vp.GotoBottom()
	return m
}

func (m fragmentModel) Update(msg tea.Msg) (fragmentModel, tea.Cmd) {
	switch msg := msg.(type) {
	case fragmentLoadedMsg:
		if msg.err != nil {
			m.status = "读取失败: " + msg.err.Error()
			return m, nil
		}
		m.lines = msg.lines
		m = m.refreshViewport()
		m.status = ""
		return m, m.input.Focus()

	case fragmentAppendedMsg:
		if msg.err != nil {
			m.status = "记录失败: " + msg.err.Error()
			return m, nil
		}
		m.lines = append(m.lines, msg.line)
		m = m.refreshViewport()
		m.status = "已记录"
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "enter" {
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			m.input.SetValue("")
			return m, func() tea.Msg {
				line, err := appendFragment(text)
				return fragmentAppendedMsg{line: line, err: err}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m fragmentModel) View() string {
	view := m.vp.View() + "\n" + m.input.View()
	if m.status != "" {
		view += "\n" + statusStyle.Render(m.status)
	}
	return view
}
