#!/bin/bash
# 每日 17:30 触发：只发桌面通知提醒记日报，不打开文件（创建当天 daily 文件仍然做，
# 方便碎片模式/日报模式随时写入，但不会弹出编辑器打断当前工作）

DAILY_DIR="$HOME/Desktop/weekly_reports/daily"
DATE=$(date +%Y%m%d)
FILE="$DAILY_DIR/$DATE.md"

# 如果文件不存在则创建，带日期标题
if [ ! -f "$FILE" ]; then
    echo "# $DATE" > "$FILE"
    echo "" >> "$FILE"
fi

# 桌面通知（不再打开编辑器）
notify-send "周报提醒" "记录一下今天的工作内容 👇" --icon=accessories-text-editor 2>/dev/null || true
