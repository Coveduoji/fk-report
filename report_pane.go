package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type reportDoneMsg struct {
	err error
}

type reportModel struct {
	sp      spinner.Model
	running bool
	result  string
	errText string
}

func newReportModel() reportModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return reportModel{sp: sp}
}

// Begin 启动一次周报生成：调用 claude -p "生成周报"，复用现有 weekly-report skill 的
// 整合润色逻辑，这里只负责触发和展示状态。
func (m reportModel) Begin() (reportModel, tea.Cmd) {
	m.running = true
	m.result = ""
	m.errText = ""
	return m, tea.Batch(m.sp.Tick, runReportCmd)
}

func runReportCmd() tea.Msg {
	home, _ := os.UserHomeDir()
	// 非交互的 -p 模式下没有人能点击权限确认，必须显式带上 acceptEdits，
	// 否则 claude 会静默拒绝写入 reports/ 目录，却仍然以 exit code 0 返回。
	cmd := exec.Command("claude", "-p", "生成周报", "--permission-mode", "acceptEdits")
	cmd.Dir = home
	out, err := cmd.CombinedOutput()
	if err == nil && strings.Contains(string(out), "权限") {
		err = errors.New(strings.TrimSpace(string(out)))
	}
	return reportDoneMsg{err: err}
}

func (m reportModel) Update(msg tea.Msg) (reportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case reportDoneMsg:
		m.running = false
		if msg.err != nil {
			m.errText = "生成失败: " + msg.err.Error()
		} else {
			m.result = "周报已生成，保存在 reports/ 目录"
		}
		return m, nil

	case spinner.TickMsg:
		if !m.running {
			return m, nil
		}
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "r" && !m.running {
			return m.Begin()
		}
	}
	return m, nil
}

func (m reportModel) View() string {
	if m.running {
		return m.sp.View() + " 正在调用 claude 生成本周周报，请稍候...\n\n" +
			"（复用 weekly-report skill 的整合润色逻辑，读取本周全部日报文件）"
	}
	if m.errText != "" {
		return errorStyle.Render(m.errText) + "\n\n按 r 重试"
	}
	if m.result != "" {
		return statusStyle.Render(m.result) + "\n\n按 r 重新生成"
	}
	return "进入本页面即会自动开始生成..."
}
