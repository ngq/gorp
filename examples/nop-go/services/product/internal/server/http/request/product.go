// Package request 定义 HTTP 请求结构体。
// 用于 handler 层绑定和校验客户端提交的请求数据。
package request

// ---------------------------------------------------------------------------
// 商品相关请求
// ---------------------------------------------------------------------------

// CreateProduct 创建商品请求。
type CreateProduct struct {
	Name           string  `json:"name" binding:"required"`            // 商品名称（必填）
	ShortDesc      string  `json:"short_desc"`                         // 简短描述
	FullDesc       string  `json:"full_desc"`                          // 完整描述（富文本）
	SKU            string  `json:"sku"`                                // 库存单位编码
	Price          float64 `json:"price" binding:"required"`           // 商品价格（必填）
	OldPrice       float64 `json:"old_price"`                          // 原价
	Cost           float64 `json:"cost"`                               // 成本价
	Stock          int     `json:"stock"`                              // 库存数量
	CategoryID     uint    `json:"category_id"`                        // 所属分类ID
	ManufacturerID uint    `json:"manufacturer_id"`                    // 制造商ID
	IsPublished    bool    `json:"is_published"`                       // 是否上架
}

// UpdateProduct 更新商品请求。
// 所有字段均为可选，仅传递需要更新的字段。
type UpdateProduct struct {
	Name           *string  `json:"name"`                               // 商品名称
	ShortDesc      *string  `json:"short_desc"`                         // 简短描述
	FullDesc       *string  `json:"full_desc"`                          // 完整描述
	SKU            *string  `json:"sku"`                                // SKU
	Price          *float64 `json:"price"`                              // 商品价格
	OldPrice       *float64 `json:"old_price"`                          // 原价
	Cost           *float64 `json:"cost"`                               // 成本价
	Stock          *int     `json:"stock"`                              // 库存数量
	CategoryID     *uint    `json:"category_id"`                        // 所属分类ID
	ManufacturerID *uint    `json:"manufacturer_id"`                    // 制造商ID
	IsPublished    *bool    `json:"is_published"`                       // 是否上架
}

// ListProductRequest 商品列表查询请求（通过 query 参数绑定）。
type ListProductRequest struct {
	Page           int    `form:"page" binding:"min=1"`               // 页码，默认1
	Size           int    `form:"size" binding:"min=1,max=100"`       // 每页数量，默认10
	CategoryID     uint   `form:"category_id"`                        // 按分类筛选
	ManufacturerID uint   `form:"manufacturer_id"`                    // 按制造商筛选
	Keyword        string `form:"keyword"`                            // 搜索关键词
}

// ---------------------------------------------------------------------------
// 分类相关请求
// ---------------------------------------------------------------------------

// CreateCategory 创建分类请求。
type CreateCategory struct {
	Name        string `json:"name" binding:"required"`             // 分类名称（必填）
	Description string `json:"description"`                         // 分类描述
	ParentID    uint   `json:"parent_id"`                           // 父分类ID，0 表示顶级分类
	SortOrder   int    `json:"sort_order"`                          // 排序权重
	IsPublished bool   `json:"is_published"`                        // 是否启用
}

// ListCategoryRequest 分类列表查询请求。
type ListCategoryRequest struct {
	Page int `form:"page" binding:"min=1"`                // 页码
	Size int `form:"size" binding:"min=1,max=100"`        // 每页数量
}

// ---------------------------------------------------------------------------
// 制造商相关请求
// ---------------------------------------------------------------------------

// CreateManufacturer 创建制造商请求。
type CreateManufacturer struct {
	Name        string `json:"name" binding:"required"`             // 制造商名称（必填）
	Description string `json:"description"`                         // 制造商描述
	IsPublished bool   `json:"is_published"`                        // 是否启用
}

// ListManufacturerRequest 制造商列表查询请求。
type ListManufacturerRequest struct {
	Page int `form:"page" binding:"min=1"`                // 页码
	Size int `form:"size" binding:"min=1,max=100"`        // 每页数量
}

// ---------------------------------------------------------------------------
// 商品评论相关请求
// ---------------------------------------------------------------------------

// CreateProductReview 创建商品评论请求。
type CreateProductReview struct {
	Title   string `json:"title" binding:"required"`             // 评论标题（必填）
	Content string `json:"content" binding:"required"`           // 评论内容（必填）
	Rating  int    `json:"rating" binding:"required,min=1,max=5"` // 评分 1-5（必填）
}

// ListProductReviewRequest 商品评论列表查询请求。
type ListProductReviewRequest struct {
	Page int `form:"page" binding:"min=1"`                 // 页码
	Size int `form:"size" binding:"min=1,max=100"`         // 每页数量
}

// ---------------------------------------------------------------------------
// 最近浏览 & 对比 & 搜索相关请求
// ---------------------------------------------------------------------------

// RecentlyViewedRequest 最近浏览商品查询请求。
type RecentlyViewedRequest struct {
	CustomerID uint `form:"customer_id" binding:"required"`       // 客户ID（必填）
	Limit      int  `form:"limit"`                                // 返回数量，默认10
}

// CompareProductsRequest 商品对比请求。
type CompareProductsRequest struct {
	ProductIDs []uint `json:"product_ids" binding:"required,min=1"` // 商品ID列表（必填，至少1个）
}

// SearchRequest 商品搜索请求。
type SearchRequest struct {
	Keyword string `form:"keyword" binding:"required"`           // 搜索关键词（必填）
	Page    int    `form:"page" binding:"min=1"`                 // 页码
	Size    int    `form:"size" binding:"min=1,max=100"`         // 每页数量
}

// AutocompleteRequest 搜索自动完成请求。
type AutocompleteRequest struct {
	Term string `form:"term" binding:"required"`              // 搜索词（必填）
}
