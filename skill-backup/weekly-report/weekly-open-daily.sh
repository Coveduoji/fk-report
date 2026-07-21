#!/bin/bash
# 每日 17:30 触发：桌面通知 + 用默认编辑器打开/创建当天 daily 文件

DAILY_DIR="$HOME/Desktop/weekly_reports/daily"
DATE=$(date +%Y%m%d)
FILE="$DAILY_DIR/$DATE.md"

# 如果文件不存在则创建，带日期标题
if [ ! -f "$FILE" ]; then
    echo "# $DATE" > "$FILE"
    echo "" >> "$FILE"
fi

# 桌面通知
notify-send "周报提醒" "记录一下今天的工作内容 👇" --icon=accessories-text-editor 2>/dev/null || true

# 打开文件（优先 gedit，其次 xdg-open）
if command -v gedit &>/dev/null; then
    gedit "$FILE" &
else
    xdg-open "$FILE" &
fi
