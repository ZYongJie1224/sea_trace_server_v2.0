package utils

import (
	"strconv"
)

// StringToInt64 将字符串转换为int64
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
