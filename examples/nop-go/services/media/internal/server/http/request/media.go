package request

// UploadMedia 异步上传图片请求。
type UploadMedia struct {
	FileName string `json:"file_name" binding:"required"` // 文件名
	MimeType string `json:"mime_type" binding:"required"` // MIME类型，如 image/png
	FileSize int64  `json:"file_size" binding:"required"` // 文件大小（字节）
	FileURL  string `json:"file_url" binding:"required"`  // 文件存储URL
	AltText  string `json:"alt_text"`                     // 图片替代文本
}
