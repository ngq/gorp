// Package service 定义 admin-service 的 gRPC RPC 接口。
// gorp proto from-service 命令通过解析此文件中的接口定义来生成 .proto 文件。
// 只有其他服务需要通过 gRPC 跨服务调用的方法才定义在此接口中。
package service

import "context"

// AdminRPC 管理服务 gRPC 接口 —— 定义其他服务（trade-service）需要跨服务调用的方法。
// trade-service 结算时需要调用 GetDiscount 和 GetDiscountByCode，
// admin 权限检查需要调用 CheckPermission（未来 gateway 也可使用）。
// 注意：税率查询（GetTaxRate）属于 trade-service 自身，不在 admin-service 中。
type AdminRPC interface {
	// GetDiscount 根据 ID 获取优惠规则（trade 结算时计算折扣）
	GetDiscount(ctx context.Context, req *GetDiscountReq) (*GetDiscountResp, error)

	// GetDiscountByCode 根据优惠码获取优惠规则（trade 结算时验证优惠码）
	GetDiscountByCode(ctx context.Context, req *GetDiscountByCodeReq) (*GetDiscountResp, error)

	// CheckPermission 检查权限（admin 权限检查、gateway 接口鉴权）
	CheckPermission(ctx context.Context, req *CheckPermissionReq) (*CheckPermissionResp, error)
}

// ======================== gRPC 请求/响应类型 ========================

// GetDiscountReq 获取优惠规则请求
type GetDiscountReq struct {
	ID uint32 `json:"id" remark:"优惠ID"`
}

// GetDiscountByCodeReq 根据优惠码获取请求
type GetDiscountByCodeReq struct {
	Code string `json:"code" remark:"优惠码"`
}

// GetDiscountResp 获取优惠规则响应
type GetDiscountResp struct {
	ID           uint32  `json:"id" remark:"优惠ID"`
	Name         string  `json:"name" remark:"优惠名称"`
	Code         string  `json:"code" remark:"优惠码"`
	Type         int32   `json:"type" remark:"优惠类型:0-百分比,1-固定金额"`
	Value        float64 `json:"value" remark:"优惠值"`
	MinAmount    float64 `json:"min_amount" remark:"最低消费金额"`
	MaxDiscount  float64 `json:"max_discount" remark:"最大优惠金额"`
	StartTime    string  `json:"start_time" remark:"生效开始时间"`
	EndTime      string  `json:"end_time" remark:"生效结束时间"`
	TotalQuota   int32   `json:"total_quota" remark:"总使用限额"`
	UsedQuota    int32   `json:"used_quota" remark:"已使用次数"`
	PerUserLimit int32   `json:"per_user_limit" remark:"每人使用限额"`
	Status       int32   `json:"status" remark:"状态:0-禁用,1-启用"`
	Description  string  `json:"description" remark:"描述"`
}

// CheckPermissionReq 检查权限请求
type CheckPermissionReq struct {
	RoleID   uint32 `json:"role_id" remark:"角色ID"`
	Resource string `json:"resource" remark:"资源标识"`
	Action   string `json:"action" remark:"操作:read/write/delete"`
}

// CheckPermissionResp 检查权限响应
type CheckPermissionResp struct {
	Allowed bool   `json:"allowed" remark:"是否允许"`
	Effect  string `json:"effect" remark:"效果:allow/deny"`
}