// Package request 提供 shipping 服务的 HTTP 请求结构体定义
package request

// ==================== 配送提供者 ====================

// UpdateProviderRequest 更新配送提供者请求
// 对应路由: PUT /api/v1/shipping/providers/:id
type UpdateProviderRequest struct {
	ID             uint   `json:"id"`               // 提供者 ID（路径参数传入）
	Name           string `json:"name"`             // 提供者名称（如顺丰、圆通）
	SystemKeyword  string `json:"system_keyword"`   // 系统关键字标识
	DisplayOrder   int    `json:"display_order"`    // 显示排序
	IsActive       bool   `json:"is_active"`        // 是否启用
	LogoURL        string `json:"logo_url"`         // 提供者 Logo 地址
	TrackingURL    string `json:"tracking_url"`     // 物流追踪 URL 模板
}

// ==================== 配送方式 ====================

// CreateMethodRequest 创建配送方式请求
// 对应路由: POST /api/v1/shipping/methods
type CreateMethodRequest struct {
	Name              string  `json:"name"`                // 配送方式名称（如标准配送、加急配送）
	SystemKeyword     string  `json:"system_keyword"`      // 系统关键字标识
	ProviderID        uint    `json:"provider_id"`         // 关联的配送提供者 ID
	DisplayOrder      int     `json:"display_order"`       // 显示排序
	IsActive          bool    `json:"is_active"`           // 是否启用
	Rate              float64 `json:"rate"`                // 基础运费
	MinOrderAmount    float64 `json:"min_order_amount"`    // 免运费最低订单金额（0 表示不免运费）
	MaxOrderAmount    float64 `json:"max_order_amount"`    // 运费适用最高订单金额
	EstimatedDays     int     `json:"estimated_days"`      // 预计配送天数
	Description       string  `json:"description"`         // 配送方式描述
}

// UpdateMethodRequest 更新配送方式请求
// 对应路由: PUT /api/v1/shipping/methods/:id
type UpdateMethodRequest struct {
	ID                uint    `json:"id"`                  // 配送方式 ID（路径参数传入）
	Name              string  `json:"name"`                // 配送方式名称
	SystemKeyword     string  `json:"system_keyword"`      // 系统关键字标识
	ProviderID        uint    `json:"provider_id"`         // 关联的配送提供者 ID
	DisplayOrder      int     `json:"display_order"`       // 显示排序
	IsActive          bool    `json:"is_active"`           // 是否启用
	Rate              float64 `json:"rate"`                // 基础运费
	MinOrderAmount    float64 `json:"min_order_amount"`    // 免运费最低订单金额
	MaxOrderAmount    float64 `json:"max_order_amount"`    // 运费适用最高订单金额
	EstimatedDays     int     `json:"estimated_days"`      // 预计配送天数
	Description       string  `json:"description"`         // 配送方式描述
}

// ==================== 配送日期 ====================

// CreateDeliveryDateRequest 创建配送日期请求
// 对应路由: POST /api/v1/shipping/delivery-dates
type CreateDeliveryDateRequest struct {
	ShippingMethodID uint   `json:"shipping_method_id"` // 关联的配送方式 ID
	DeliveryDate     string `json:"delivery_date"`      // 可选配送日期（YYYY-MM-DD 格式）
	IsAvailable      bool   `json:"is_available"`        // 该日期是否可选
	Description      string `json:"description"`         // 日期说明（如节假日提示）
}

// UpdateDeliveryDateRequest 更新配送日期请求
// 对应路由: PUT /api/v1/shipping/delivery-dates/:id
type UpdateDeliveryDateRequest struct {
	ID               uint   `json:"id"`                  // 配送日期 ID（路径参数传入）
	ShippingMethodID uint   `json:"shipping_method_id"`  // 关联的配送方式 ID
	DeliveryDate     string `json:"delivery_date"`       // 可选配送日期
	IsAvailable      bool   `json:"is_available"`        // 该日期是否可选
	Description      string `json:"description"`         // 日期说明
}

// ==================== 仓库 ====================

// CreateWarehouseRequest 创建仓库请求
// 对应路由: POST /api/v1/shipping/warehouses
type CreateWarehouseRequest struct {
	Name        string `json:"name"`          // 仓库名称
	Code        string `json:"code"`          // 仓库编码
	Address     string `json:"address"`       // 仓库地址
	City        string `json:"city"`          // 城市
	CountryID   uint   `json:"country_id"`    // 国家 ID
	StateID     uint   `json:"state_id"`      // 省/州 ID
	ZipCode     string `json:"zip_code"`      // 邮编
	PhoneNumber string `json:"phone_number"`  // 联系电话
	IsActive    bool   `json:"is_active"`     // 是否启用
	Latitude    float64 `json:"latitude"`     // 纬度
	Longitude   float64 `json:"longitude"`    // 经度
}

// UpdateWarehouseRequest 更新仓库请求
// 对应路由: PUT /api/v1/shipping/warehouses/:id
type UpdateWarehouseRequest struct {
	ID          uint    `json:"id"`            // 仓库 ID（路径参数传入）
	Name        string  `json:"name"`          // 仓库名称
	Code        string  `json:"code"`          // 仓库编码
	Address     string  `json:"address"`       // 仓库地址
	City        string  `json:"city"`          // 城市
	CountryID   uint    `json:"country_id"`    // 国家 ID
	StateID     uint    `json:"state_id"`      // 省/州 ID
	ZipCode     string  `json:"zip_code"`      // 邮编
	PhoneNumber string  `json:"phone_number"`  // 联系电话
	IsActive    bool    `json:"is_active"`     // 是否启用
	Latitude    float64 `json:"latitude"`      // 纬度
	Longitude   float64 `json:"longitude"`     // 经度
}

// ==================== 运费估算 ====================

// EstimateShippingRequest 运费估算请求
// 对应路由: GET /api/v1/shipping/estimate
type EstimateShippingRequest struct {
	WarehouseID string `form:"warehouse_id" json:"warehouse_id"` // 发货仓库 ID
	CountryID   string `form:"country_id" json:"country_id"`     // 收货国家 ID
	StateID     string `form:"state_id" json:"state_id"`         // 收货省/州 ID
	ZipCode     string `form:"zip_code" json:"zip_code"`         // 收货邮编
	SubTotal    string `form:"sub_total" json:"sub_total"`       // 订单小计金额
	Weight      string `form:"weight" json:"weight"`             // 商品总重量（kg）
}
