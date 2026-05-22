package response

import "time"

// Media 媒体响应结构体。
type Media struct {
	ID        uint      `json:"id"`         // 媒体ID
	FileName  string    `json:"file_name"`  // 文件名
	MimeType  string    `json:"mime_type"`  // MIME类型
	FileSize  int64     `json:"file_size"`  // 文件大小（字节）
	FileURL   string    `json:"file_url"`   // 文件存储URL
	AltText   string    `json:"alt_text"`   // 图片替代文本
	CreatedAt time.Time `json:"created_at"` // 创建时间
}
