package data

import "time"

// unixToTime 将 Unix 时间戳转换为 time.Time。
// 如果时间戳为 0，则返回零值，避免返回 1970-01-01。
func unixToTime(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}