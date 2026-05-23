// Package request 定义 GDPR 相关的 HTTP 请求结构体。
// 用于参数绑定和校验。
package request

// CreateGdprRequest 创建 GDPR 请求
type CreateGdprRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`                      // 所属用户 ID
	RequestType string `json:"request_type" binding:"required,oneof=delete export"` // 请求类型：delete 或 export
	Reason      string `json:"reason" binding:"omitempty,max=512"`              // 请求原因
}

// UpdateGdprRequest 更新 GDPR 请求
type UpdateGdprRequest struct {
	RequestType string `json:"request_type" binding:"omitempty,oneof=delete export"` // 请求类型
	Reason      string `json:"reason" binding:"omitempty,max=512"`                   // 请求原因
}

// ReviewGdprRequest 审核 GDPR 请求
type ReviewGdprRequest struct {
	Status     string `json:"status" binding:"required,oneof=approved rejected"` // 审核结果：approved 或 rejected
	ReviewedBy uint   `json:"reviewed_by" binding:"required"`                   // 审核人 ID
}