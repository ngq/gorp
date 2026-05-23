// Package response 定义 HTTP 响应结构体（媒体相关）。
package response

import "time"

// Media 媒体响应结构体。
type Media struct {
	ID        uint      `json:"id"`
	FileName  string    `json:"file_name"`
	MimeType  string    `json:"mime_type"`
	FileSize  int64     `json:"file_size"`
	FileURL   string    `json:"file_url"`
	AltText   string    `json:"alt_text"`
	CreatedAt time.Time `json:"created_at"`
}