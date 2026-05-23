// Package service 定义 user-service 的 gRPC RPC 接口。
// gorp proto from-service 命令通过解析此文件中的接口定义来生成 .proto 文件。
// 只有其他服务需要通过 gRPC 跨服务调用的方法才定义在此接口中，
// HTTP handler 对应的 CRUD 方法不需要出现在 RPC 接口中。
package service

import "context"

// UserRPC 用户服务 gRPC 接口 —— 定义其他服务（trade-service、admin-service）需要跨服务调用的方法。
type UserRPC interface {
	// GetUser 根据 ID 获取用户基本信息（trade 结算时验证客户、admin 权限检查时获取用户角色）
	GetUser(ctx context.Context, req *GetUserReq) (*GetUserResp, error)

	// ListUserAddresses 获取用户地址列表（trade 结算时获取收货地址）
	ListUserAddresses(ctx context.Context, req *ListUserAddressesReq) (*ListUserAddressesResp, error)
}

// ======================== gRPC 请求/响应类型 ========================

// GetUserReq 获取用户请求
type GetUserReq struct {
	ID uint32 `json:"id" remark:"用户ID"`
}

// GetUserResp 获取用户响应
type GetUserResp struct {
	ID        uint32 `json:"id" remark:"用户ID"`
	Username  string `json:"username" remark:"用户名"`
	Email     string `json:"email" remark:"邮箱"`
	Phone     string `json:"phone" remark:"手机号"`
	Nickname  string `json:"nickname" remark:"昵称"`
	Avatar    string `json:"avatar" remark:"头像URL"`
	Status    int32  `json:"status" remark:"状态"`
	CreatedAt string `json:"created_at" remark:"创建时间"`
	UpdatedAt string `json:"updated_at" remark:"更新时间"`
}

// ListUserAddressesReq 获取用户地址列表请求
type ListUserAddressesReq struct {
	UserID uint32 `json:"user_id" remark:"用户ID"`
}

// ListUserAddressesResp 获取用户地址列表响应
type ListUserAddressesResp struct {
	Items []AddressItem `json:"items" remark:"地址列表"`
}

// AddressItem 地址条目
type AddressItem struct {
	ID            uint32 `json:"id" remark:"地址ID"`
	UserID        uint32 `json:"user_id" remark:"用户ID"`
	RecipientName string `json:"recipient_name" remark:"收件人姓名"`
	Phone         string `json:"phone" remark:"收件人电话"`
	Province      string `json:"province" remark:"省份"`
	City          string `json:"city" remark:"城市"`
	District      string `json:"district" remark:"区县"`
	Detail        string `json:"detail" remark:"详细地址"`
	IsDefault     bool   `json:"is_default" remark:"是否默认地址"`
}