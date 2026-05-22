// Package response 提供 shipping 服务的 HTTP 响应结构体定义
package response

// ==================== 配送提供者 ====================

// ProviderResponse 配送提供者响应
// 对应 nopCommerce Admin/Shipping/Providers 返回数据
type ProviderResponse struct {
	ID            uint   `json:"id"`              // 提供者 ID
	Name          string `json:"name"`            // 提供者名称
	SystemKeyword string `json:"system_keyword"`  // 系统关键字标识
	DisplayOrder  int    `json:"display_order"`   // 显示排序
	IsActive      bool   `json:"is_active"`       // 是否启用
	LogoURL       string `json:"logo_url"`        // Logo 地址
	TrackingURL   string `json:"tracking_url"`    // 物流追踪 URL 模板
	CreatedAt     int64  `json:"created_at"`      // 创建时间（Unix 时间戳）
	UpdatedAt     int64  `json:"updated_at"`      // 更新时间（Unix 时间戳）
}

// ListProvidersResponse 配送提供者列表响应
type ListProvidersResponse struct {
	Total int64              `json:"total"` // 总数
	Items []*ProviderResponse `json:"items"` // 提供者列表
}

// ==================== 配送方式 ====================

// MethodResponse 配送方式响应
// 对应 nopCommerce Admin/Shipping/Methods 返回数据
type MethodResponse struct {
	ID               uint    `json:"id"`                // 配送方式 ID
	Name             string  `json:"name"`              // 配送方式名称
	SystemKeyword    string  `json:"system_keyword"`    // 系统关键字标识
	ProviderID       uint    `json:"provider_id"`       // 关联的配送提供者 ID
	ProviderName     string  `json:"provider_name"`     // 配送提供者名称（冗余展示）
	DisplayOrder     int     `json:"display_order"`     // 显示排序
	IsActive         bool    `json:"is_active"`         // 是否启用
	Rate             float64 `json:"rate"`              // 基础运费
	MinOrderAmount   float64 `json:"min_order_amount"`  // 免运费最低订单金额
	MaxOrderAmount   float64 `json:"max_order_amount"`  // 运费适用最高订单金额
	EstimatedDays    int     `json:"estimated_days"`    // 预计配送天数
	Description      string  `json:"description"`       // 配送方式描述
	CreatedAt        int64   `json:"created_at"`        // 创建时间（Unix 时间戳）
	UpdatedAt        int64   `json:"updated_at"`        // 更新时间（Unix 时间戳）
}

// ListMethodsResponse 配送方式列表响应
type ListMethodsResponse struct {
	Total int64            `json:"total"` // 总数
	Items []*MethodResponse `json:"items"` // 配送方式列表
}

// ==================== 配送日期 ====================

// DeliveryDateResponse 配送日期响应
// 对应 nopCommerce Admin/Shipping/DeliveryDates 返回数据
type DeliveryDateResponse struct {
	ID               uint   `json:"id"`                // 配送日期 ID
	ShippingMethodID uint   `json:"shipping_method_id"` // 关联的配送方式 ID
	DeliveryDate     string `json:"delivery_date"`     // 可选配送日期
	IsAvailable      bool   `json:"is_available"`      // 该日期是否可选
	Description      string `json:"description"`       // 日期说明
	CreatedAt        int64  `json:"created_at"`        // 创建时间（Unix 时间戳）
	UpdatedAt        int64  `json:"updated_at"`        // 更新时间（Unix 时间戳）
}

// ListDeliveryDatesResponse 配送日期列表响应
type ListDeliveryDatesResponse struct {
	Total int64                  `json:"total"` // 总数
	Items []*DeliveryDateResponse `json:"items"` // 配送日期列表
}

// ==================== 仓库 ====================

// WarehouseResponse 仓库响应
// 对应 nopCommerce Admin/Shipping/Warehouses 返回数据
type WarehouseResponse struct {
	ID          uint    `json:"id"`           // 仓库 ID
	Name        string  `json:"name"`         // 仓库名称
	Code        string  `json:"code"`         // 仓库编码
	Address     string  `json:"address"`      // 仓库地址
	City        string  `json:"city"`         // 城市
	CountryID   uint    `json:"country_id"`   // 国家 ID
	StateID     uint    `json:"state_id"`     // 省/州 ID
	ZipCode     string  `json:"zip_code"`     // 邮编
	PhoneNumber string  `json:"phone_number"` // 联系电话
	IsActive    bool    `json:"is_active"`    // 是否启用
	Latitude    float64 `json:"latitude"`     // 纬度
	Longitude   float64 `json:"longitude"`    // 经度
	CreatedAt   int64   `json:"created_at"`   // 创建时间（Unix 时间戳）
	UpdatedAt   int64   `json:"updated_at"`   // 更新时间（Unix 时间戳）
}

// ListWarehousesResponse 仓库列表响应
type ListWarehousesResponse struct {
	Total int64               `json:"total"` // 总数
	Items []*WarehouseResponse `json:"items"` // 仓库列表
}

// ==================== 运费估算 ====================

// ShippingEstimateResponse 运费估算响应
// 对应 nopCommerce Admin/Shipping/Estimate 返回数据
type ShippingEstimateResponse struct {
	MethodID       uint    `json:"method_id"`       // 配送方式 ID
	MethodName     string  `json:"method_name"`     // 配送方式名称
	ProviderName   string  `json:"provider_name"`   // 配送提供者名称
	Rate           float64 `json:"rate"`            // 估算运费
	EstimatedDays  int     `json:"estimated_days"`  // 预计配送天数
	IsFreeShipping bool    `json:"is_free_shipping"` // 是否免运费
}

// ListShippingEstimatesResponse 运费估算列表响应
// 根据条件返回所有可用配送方式的运费估算
type ListShippingEstimatesResponse struct {
	Items []*ShippingEstimateResponse `json:"items"` // 运费估算列表
}
