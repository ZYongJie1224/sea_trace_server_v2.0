package utils

import (
	"time"
)

// FormatCurrentTime 格式化当前时间为 YYYY-MM-DD HH:MM:SS
func FormatCurrentTime() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

// ParseUTCTime 解析 YYYY-MM-DD HH:MM:SS 格式的UTC时间
func ParseUTCTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}
