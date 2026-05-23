// Package response 定义用户相关的 HTTP 响应结构体。
// 用于统一响应格式。
package response

import "time"

// UserResponse 用户响应
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddressResponse 地址响应
type AddressResponse struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	Detail        string `json:"detail"`
	IsDefault     bool   `json:"is_default"`
}

// ExternalAssociationResponse 外部关联响应
type ExternalAssociationResponse struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Platform     string `json:"platform"`
	ExternalID   string `json:"external_id"`
	ExternalData string `json:"external_data"`
}

// DownloadableProductResponse 可下载产品响应
type DownloadableProductResponse struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	ProductID   string     `json:"product_id"`
	ProductName string     `json:"product_name"`
	DownloadURL string     `json:"download_url"`
	ExpireAt    *time.Time `json:"expire_at"`
}

// UserListResponse 用户列表响应（带分页信息）
type UserListResponse struct {
	Data  []*UserResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}