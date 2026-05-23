// Package response 定义 HTTP 响应结构体（商品、分类、制造商、评论相关）。
package response

// ---------------------------------------------------------------------------
// 商品相关响应
// ---------------------------------------------------------------------------

// Product 商品详情响应。
type Product struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	ShortDesc      string  `json:"short_desc"`
	FullDesc       string  `json:"full_desc"`
	SKU            string  `json:"sku"`
	Price          float64 `json:"price"`
	OldPrice       float64 `json:"old_price"`
	Cost           float64 `json:"cost"`
	Stock          int     `json:"stock"`
	CategoryID     uint    `json:"category_id"`
	ManufacturerID uint    `json:"manufacturer_id"`
	IsPublished    bool    `json:"is_published"`
	CreatedAt      int64   `json:"created_at"`
	UpdatedAt      int64   `json:"updated_at"`
}

// ProductListItem 商品列表项响应（精简字段，用于列表展示）。
type ProductListItem struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	ShortDesc      string  `json:"short_desc"`
	SKU            string  `json:"sku"`
	Price          float64 `json:"price"`
	OldPrice       float64 `json:"old_price"`
	Stock          int     `json:"stock"`
	CategoryID     uint    `json:"category_id"`
	ManufacturerID uint    `json:"manufacturer_id"`
	IsPublished    bool    `json:"is_published"`
	CreatedAt      int64   `json:"created_at"`
}

// ProductList 商品列表响应（含分页信息）。
type ProductList struct {
	Items []ProductListItem `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

// ---------------------------------------------------------------------------
// 分类相关响应
// ---------------------------------------------------------------------------

// Category 分类详情响应。
type Category struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    uint   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// CategoryList 分类列表响应。
type CategoryList struct {
	Items []Category `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// ---------------------------------------------------------------------------
// 制造商相关响应
// ---------------------------------------------------------------------------

// Manufacturer 制造商详情响应。
type Manufacturer struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// ManufacturerList 制造商列表响应。
type ManufacturerList struct {
	Items []Manufacturer `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
}

// ---------------------------------------------------------------------------
// 商品评论相关响应
// ---------------------------------------------------------------------------

// ProductReview 商品评论响应。
type ProductReview struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	CustomerID   uint   `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Rating       int    `json:"rating"`
	IsApproved   bool   `json:"is_approved"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// ProductReviewList 商品评论列表响应。
type ProductReviewList struct {
	Items []ProductReview `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// ---------------------------------------------------------------------------
// 最近浏览 & 对比 & 搜索相关响应
// ---------------------------------------------------------------------------

// RecentlyViewedList 最近浏览商品列表响应。
type RecentlyViewedList struct {
	Items []ProductListItem `json:"items"`
}

// CompareResult 商品对比结果响应。
type CompareResult struct {
	Items []Product `json:"items"`
}

// SearchResult 商品搜索结果响应。
type SearchResult struct {
	Items []ProductListItem `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

// AutocompleteResult 搜索自动完成响应。
type AutocompleteResult struct {
	Suggestions []string `json:"suggestions"`
}