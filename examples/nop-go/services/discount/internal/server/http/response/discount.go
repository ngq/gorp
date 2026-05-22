// Package response 提供 discount 服务的 HTTP 响应结构体定义
package response

// ==================== 折扣 CRUD ====================

// DiscountResponse 折扣响应
// 对应 nopCommerce Admin/Discount 返回数据
type DiscountResponse struct {
	ID                uint    `json:"id"`                     // 折扣 ID
	Name              string  `json:"name"`                   // 折扣名称
	DiscountType      string  `json:"discount_type"`          // 折扣类型（percentage/fixed/free_shipping）
	DiscountAmount    float64 `json:"discount_amount"`        // 折扣金额/百分比
	StartDate         string  `json:"start_date"`             // 折扣开始日期
	EndDate           string  `json:"end_date"`               // 折扣结束日期
	RequiresCouponCode bool   `json:"requires_coupon_code"`   // 是否需要优惠券码
	CouponCode        string  `json:"coupon_code"`            // 优惠券码
	IsCumulative      bool    `json:"is_cumulative"`          // 是否可叠加使用
	DisplayOrder      int     `json:"display_order"`          // 显示排序
	IsActive          bool    `json:"is_active"`              // 是否启用
	LimitationTimes   int     `json:"limitation_times"`       // 使用次数限制
	CreatedAt         int64   `json:"created_at"`             // 创建时间（Unix 时间戳）
	UpdatedAt         int64   `json:"updated_at"`             // 更新时间（Unix 时间戳）
}

// ListDiscountsResponse 折扣列表响应
type ListDiscountsResponse struct {
	Total int64              `json:"total"` // 总数
	Items []*DiscountResponse `json:"items"` // 折扣列表
}

// ==================== 折扣关联商品 ====================

// DiscountProductResponse 折扣关联商品响应
// 对应 nopCommerce Admin/Discount/ProductList 返回数据
type DiscountProductResponse struct {
	ID          uint   `json:"id"`           // 关联记录 ID
	DiscountID  uint   `json:"discount_id"`  // 折扣 ID
	ProductID   uint   `json:"product_id"`   // 商品 ID
	ProductName string `json:"product_name"` // 商品名称（冗余展示）
	CreatedAt   int64  `json:"created_at"`   // 创建时间（Unix 时间戳）
}

// ListDiscountProductsResponse 折扣关联商品列表响应
type ListDiscountProductsResponse struct {
	Total int64                    `json:"total"` // 总数
	Items []*DiscountProductResponse `json:"items"` // 关联商品列表
}

// ==================== 折扣关联分类 ====================

// DiscountCategoryResponse 折扣关联分类响应
// 对应 nopCommerce Admin/Discount/CategoryList 返回数据
type DiscountCategoryResponse struct {
	ID           uint   `json:"id"`            // 关联记录 ID
	DiscountID   uint   `json:"discount_id"`   // 折扣 ID
	CategoryID   uint   `json:"category_id"`   // 分类 ID
	CategoryName string `json:"category_name"` // 分类名称（冗余展示）
	CreatedAt    int64  `json:"created_at"`    // 创建时间（Unix 时间戳）
}

// ListDiscountCategoriesResponse 折扣关联分类列表响应
type ListDiscountCategoriesResponse struct {
	Total int64                       `json:"total"` // 总数
	Items []*DiscountCategoryResponse `json:"items"` // 关联分类列表
}

// ==================== 折扣关联制造商 ====================

// DiscountManufacturerResponse 折扣关联制造商响应
// 对应 nopCommerce Admin/Discount/ManufacturerList 返回数据
type DiscountManufacturerResponse struct {
	ID               uint   `json:"id"`                // 关联记录 ID
	DiscountID       uint   `json:"discount_id"`       // 折扣 ID
	ManufacturerID   uint   `json:"manufacturer_id"`   // 制造商 ID
	ManufacturerName string `json:"manufacturer_name"` // 制造商名称（冗余展示）
	CreatedAt        int64  `json:"created_at"`        // 创建时间（Unix 时间戳）
}

// ListDiscountManufacturersResponse 折扣关联制造商列表响应
type ListDiscountManufacturersResponse struct {
	Total int64                           `json:"total"` // 总数
	Items []*DiscountManufacturerResponse `json:"items"` // 关联制造商列表
}

// ==================== 折扣使用历史 ====================

// DiscountUsageHistoryResponse 折扣使用历史响应
// 对应 nopCommerce Admin/Discount/UsageHistoryList 返回数据
type DiscountUsageHistoryResponse struct {
	ID            uint   `json:"id"`             // 使用记录 ID
	DiscountID    uint   `json:"discount_id"`    // 折扣 ID
	OrderID       uint   `json:"order_id"`       // 订单 ID
	CustomerID    uint   `json:"customer_id"`    // 客户 ID
	CustomerName  string `json:"customer_name"`  // 客户名称（冗余展示）
	CouponCode    string `json:"coupon_code"`    // 使用的优惠券码
	UsedOn        string `json:"used_on"`        // 使用日期
	CreatedAt     int64  `json:"created_at"`     // 创建时间（Unix 时间戳）
}

// ListDiscountUsageHistoryResponse 折扣使用历史列表响应
type ListDiscountUsageHistoryResponse struct {
	Total int64                           `json:"total"` // 总数
	Items []*DiscountUsageHistoryResponse `json:"items"` // 使用历史列表
}