#!/bin/bash
# 每日 10:00 触发：创建当天 daily 文件（若不存在）

DAILY_DIR="$HOME/Desktop/weekly_reports/daily"
DATE=$(date +%Y%m%d)
FILE="$DAILY_DIR/$DATE.md"

mkdir -p "$DAILY_DIR"

if [ ! -f "$FILE" ]; then
    echo "# $DATE" > "$FILE"
    echo "" >> "$FILE"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 创建 $FILE"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $FILE 已存在，跳过"
fi
