package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func weeklyReportsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, "Desktop", "weekly_reports")
}

func dailyDir() string {
	return filepath.Join(weeklyReportsRoot(), "daily")
}

func todoFile() string {
	return filepath.Join(weeklyReportsRoot(), "todo.md")
}

func todayFile() string {
	return filepath.Join(dailyDir(), time.Now().Format("20060102")+".md")
}

func nowClock() string {
	return time.Now().Format("15:04")
}

// ensureDailyFile 确保当天日报文件存在，返回其路径。
func ensureDailyFile() (string, error) {
	dir := dailyDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	file := todayFile()
	if _, err := os.Stat(file); os.IsNotExist(err) {
		content := "# " + time.Now().Format("20060102") + "\n\n"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			return "", err
		}
	}
	return file, nil
}

func readDailyFile() (string, error) {
	file, err := ensureDailyFile()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeDailyFile(content string) error {
	file, err := ensureDailyFile()
	if err != nil {
		return err
	}
	return os.WriteFile(file, []byte(content), 0644)
}

// appendFragment 追加一条带时间戳的碎片记录，返回写入的完整行（不含换行符）。
func appendFragment(text string) (string, error) {
	file, err := ensureDailyFile()
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	line := "- " + nowClock() + " " + text
	if _, err := f.WriteString(line + "\n"); err != nil {
		return "", err
	}
	return line, nil
}

// fragmentLines 从当天日报内容里提取形如 "- HH:MM ..." 的碎片记录行。
func fragmentLines(content string) []string {
	var lines []string
	for _, l := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "- ") {
			lines = append(lines, l)
		}
	}
	return lines
}

type todoItem struct {
	text string
	done bool
}

// readTodos 读取 todo.md，格式为 "- [ ] 内容" / "- [x] 内容"。文件不存在时返回空列表。
func readTodos() ([]todoItem, error) {
	b, err := os.ReadFile(todoFile())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var items []todoItem
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "- [x] "):
			items = append(items, todoItem{text: strings.TrimPrefix(l, "- [x] "), done: true})
		case strings.HasPrefix(l, "- [ ] "):
			items = append(items, todoItem{text: strings.TrimPrefix(l, "- [ ] "), done: false})
		}
	}
	return items, nil
}

// writeTodos 把整个待办列表覆盖写回 todo.md。
func writeTodos(items []todoItem) error {
	dir := weeklyReportsRoot()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var b strings.Builder
	for _, it := range items {
		mark := " "
		if it.done {
			mark = "x"
		}
		b.WriteString("- [" + mark + "] " + it.text + "\n")
	}
	return os.WriteFile(todoFile(), []byte(b.String()), 0644)
}
