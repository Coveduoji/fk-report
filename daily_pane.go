package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type dailyLoadedMsg struct {
	content string
	err     error
}

type dailySavedMsg struct {
	err error
}

type dailyModel struct {
	ta     textarea.Model
	status string
}

func newDailyModel() dailyModel {
	ta := textarea.New()
	ta.Placeholder = "今天做了什么..."
	ta.ShowLineNumbers = false
	return dailyModel{ta: ta}
}

func (m dailyModel) SetSize(w, h int) dailyModel {
	if w > 4 {
		m.ta.SetWidth(w - 2)
	}
	if h > 4 {
		m.ta.SetHeight(h - 2)
	}
	return m
}

func (m dailyModel) Load() tea.Cmd {
	return func() tea.Msg {
		content, err := readDailyFile()
		return dailyLoadedMsg{content: content, err: err}
	}
}

func (m dailyModel) Update(msg tea.Msg) (dailyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dailyLoadedMsg:
		if msg.err != nil {
			m.status = "读取失败: " + msg.err.Error()
			return m, nil
		}
		m.ta.SetValue(msg.content)
		m.status = ""
		return m, m.ta.Focus()

	case dailySavedMsg:
		if msg.err != nil {
			m.status = "保存失败: " + msg.err.Error()
		} else {
			m.status = "已保存"
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+s" {
			content := m.ta.Value()
			return m, func() tea.Msg {
				return dailySavedMsg{err: writeDailyFile(content)}
			}
		}
	}

	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd
}

func (m dailyModel) View() string {
	view := m.ta.View()
	if m.status != "" {
		view += "\n" + statusStyle.Render(m.status)
	}
	return view
}
