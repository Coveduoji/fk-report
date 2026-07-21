#!/bin/bash
# 周五 17:50 触发：调用 opencode CLI 自动生成本周周报

LOG_DIR="$HOME/Desktop/weekly_reports"

# 计算本周一和本周五日期
MON=$(date -d "monday" +%Y-%m-%d 2>/dev/null || date -v-mon +%Y-%m-%d)
FRI=$(date -d "friday" +%Y-%m-%d 2>/dev/null || date -v+fri +%Y-%m-%d)

notify-send "周报生成中" "正在自动生成本周周报（opencode），请稍候..." --icon=document-new 2>/dev/null || true

cd "$HOME" && opencode run "生成周报" >> "$LOG_DIR/cron.log" 2>&1

notify-send "周报已生成" "本周周报已写入 reports/ 目录" --icon=dialog-information 2>/dev/null || true
