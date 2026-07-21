package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func dailyDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, "Desktop", "weekly_reports", "daily")
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
