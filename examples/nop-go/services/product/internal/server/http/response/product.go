// Package response 定义 HTTP 响应结构体。
// 用于 handler 层构造返回给客户端的 JSON 数据结构。
package response

// ---------------------------------------------------------------------------
// 商品相关响应
// ---------------------------------------------------------------------------

// Product 商品详情响应。
type Product struct {
	ID             uint    `json:"id"`               // 商品ID
	Name           string  `json:"name"`             // 商品名称
	ShortDesc      string  `json:"short_desc"`       // 简短描述
	FullDesc       string  `json:"full_desc"`        // 完整描述
	SKU            string  `json:"sku"`              // 库存单位编码
	Price          float64 `json:"price"`            // 商品价格
	OldPrice       float64 `json:"old_price"`        // 原价
	Cost           float64 `json:"cost"`             // 成本价
	Stock          int     `json:"stock"`            // 库存数量
	CategoryID     uint    `json:"category_id"`      // 所属分类ID
	ManufacturerID uint    `json:"manufacturer_id"`  // 制造商ID
	IsPublished    bool    `json:"is_published"`     // 是否上架
	CreatedAt      int64   `json:"created_at"`       // 创建时间（Unix时间戳）
	UpdatedAt      int64   `json:"updated_at"`       // 更新时间（Unix时间戳）
}

// ProductListItem 商品列表项响应（精简字段，用于列表展示）。
type ProductListItem struct {
	ID             uint    `json:"id"`               // 商品ID
	Name           string  `json:"name"`             // 商品名称
	ShortDesc      string  `json:"short_desc"`       // 简短描述
	SKU            string  `json:"sku"`              // SKU
	Price          float64 `json:"price"`            // 商品价格
	OldPrice       float64 `json:"old_price"`        // 原价
	Stock          int     `json:"stock"`            // 库存数量
	CategoryID     uint    `json:"category_id"`      // 所属分类ID
	ManufacturerID uint    `json:"manufacturer_id"`  // 制造商ID
	IsPublished    bool    `json:"is_published"`     // 是否上架
	CreatedAt      int64   `json:"created_at"`       // 创建时间
}

// ProductList 商品列表响应（含分页信息）。
type ProductList struct {
	Items []ProductListItem `json:"items"` // 商品列表
	Total int64             `json:"total"` // 总记录数
	Page  int               `json:"page"`  // 当前页码
	Size  int               `json:"size"`  // 每页数量
}

// ---------------------------------------------------------------------------
// 分类相关响应
// ---------------------------------------------------------------------------

// Category 分类详情响应。
type Category struct {
	ID          uint   `json:"id"`           // 分类ID
	Name        string `json:"name"`         // 分类名称
	Description string `json:"description"`  // 分类描述
	ParentID    uint   `json:"parent_id"`    // 父分类ID
	SortOrder   int    `json:"sort_order"`   // 排序权重
	IsPublished bool   `json:"is_published"` // 是否启用
	CreatedAt   int64  `json:"created_at"`   // 创建时间
	UpdatedAt   int64  `json:"updated_at"`   // 更新时间
}

// CategoryList 分类列表响应。
type CategoryList struct {
	Items []Category `json:"items"` // 分类列表
	Total int64      `json:"total"` // 总记录数
	Page  int        `json:"page"`  // 当前页码
	Size  int        `json:"size"`  // 每页数量
}

// ---------------------------------------------------------------------------
// 制造商相关响应
// ---------------------------------------------------------------------------

// Manufacturer 制造商详情响应。
type Manufacturer struct {
	ID          uint   `json:"id"`           // 制造商ID
	Name        string `json:"name"`         // 制造商名称
	Description string `json:"description"`  // 制造商描述
	IsPublished bool   `json:"is_published"` // 是否启用
	CreatedAt   int64  `json:"created_at"`   // 创建时间
	UpdatedAt   int64  `json:"updated_at"`   // 更新时间
}

// ManufacturerList 制造商列表响应。
type ManufacturerList struct {
	Items []Manufacturer `json:"items"` // 制造商列表
	Total int64          `json:"total"` // 总记录数
	Page  int            `json:"page"`  // 当前页码
	Size  int            `json:"size"`  // 每页数量
}

// ---------------------------------------------------------------------------
// 商品评论相关响应
// ---------------------------------------------------------------------------

// ProductReview 商品评论响应。
type ProductReview struct {
	ID           uint   `json:"id"`            // 评论ID
	ProductID    uint   `json:"product_id"`    // 商品ID
	CustomerID   uint   `json:"customer_id"`   // 评论者客户ID
	CustomerName string `json:"customer_name"` // 评论者名称
	Title        string `json:"title"`         // 评论标题
	Content      string `json:"content"`       // 评论内容
	Rating       int    `json:"rating"`        // 评分 1-5
	IsApproved   bool   `json:"is_approved"`   // 是否审核通过
	CreatedAt    int64  `json:"created_at"`    // 创建时间
	UpdatedAt    int64  `json:"updated_at"`    // 更新时间
}

// ProductReviewList 商品评论列表响应。
type ProductReviewList struct {
	Items []ProductReview `json:"items"` // 评论列表
	Total int64           `json:"total"` // 总记录数
	Page  int             `json:"page"`  // 当前页码
	Size  int             `json:"size"`  // 每页数量
}

// ---------------------------------------------------------------------------
// 最近浏览 & 对比 & 搜索相关响应
// ---------------------------------------------------------------------------

// RecentlyViewedList 最近浏览商品列表响应。
type RecentlyViewedList struct {
	Items []ProductListItem `json:"items"` // 商品列表
}

// CompareResult 商品对比结果响应。
type CompareResult struct {
	Items []Product `json:"items"` // 商品详情列表
}

// SearchResult 商品搜索结果响应。
type SearchResult struct {
	Items []ProductListItem `json:"items"` // 搜索结果列表
	Total int64             `json:"total"` // 总匹配数
	Page  int               `json:"page"`  // 当前页码
	Size  int               `json:"size"`  // 每页数量
}

// AutocompleteResult 搜索自动完成响应。
type AutocompleteResult struct {
	Suggestions []string `json:"suggestions"` // 建议词列表
}
