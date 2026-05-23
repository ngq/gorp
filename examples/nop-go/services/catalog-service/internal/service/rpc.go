// Package service 定义 catalog-service 的 gRPC RPC 接口。
// gorp proto from-service 命令通过解析此文件中的接口定义来生成 .proto 文件。
// 只有其他服务需要通过 gRPC 跨服务调用的方法才定义在此接口中。
package service

import "context"

// CatalogRPC 商品服务 gRPC 接口 —— 定义其他服务（trade-service、content-service）需要跨服务调用的方法。
type CatalogRPC interface {
	// GetProduct 根据 ID 获取商品信息（trade 结算时获取商品价格、content 商品展示时获取商品详情）
	GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error)

	// ListProducts 获取商品列表（trade 批量验证库存、content 商品展示列表）
	ListProducts(ctx context.Context, req *ListProductsReq) (*ListProductsResp, error)

	// GetCategory 根据 ID 获取分类信息（content 商品分类展示）
	GetCategory(ctx context.Context, req *GetCategoryReq) (*GetCategoryResp, error)

	// ListCountries 获取国家列表（trade 结算时选择配送国家）
	ListCountries(ctx context.Context, req *ListCountriesReq) (*ListCountriesResp, error)
}

// ======================== gRPC 请求/响应类型 ========================

// GetProductReq 获取商品请求
type GetProductReq struct {
	ID uint32 `json:"id" remark:"商品ID"`
}

// GetProductResp 获取商品响应
type GetProductResp struct {
	ID             uint32  `json:"id" remark:"商品ID"`
	Name           string  `json:"name" remark:"商品名称"`
	ShortDesc      string  `json:"short_desc" remark:"商品简介"`
	FullDesc       string  `json:"full_desc" remark:"商品详细描述"`
	Sku            string  `json:"sku" remark:"SKU编码"`
	Price          float64 `json:"price" remark:"售价"`
	OldPrice       float64 `json:"old_price" remark:"原价"`
	Cost           float64 `json:"cost" remark:"成本价"`
	Stock          int32   `json:"stock" remark:"库存数量"`
	CategoryID     uint32  `json:"category_id" remark:"分类ID"`
	ManufacturerID uint32  `json:"manufacturer_id" remark:"制造商ID"`
	IsPublished    bool    `json:"is_published" remark:"是否上架"`
	CreatedAt      int64   `json:"created_at" remark:"创建时间"`
	UpdatedAt      int64   `json:"updated_at" remark:"更新时间"`
}

// ListProductsReq 获取商品列表请求
type ListProductsReq struct {
	Page           int32  `json:"page" remark:"页码"`
	Size           int32  `json:"size" remark:"每页条数"`
	CategoryID     uint32 `json:"category_id" remark:"分类ID（可选）"`
	ManufacturerID uint32 `json:"manufacturer_id" remark:"制造商ID（可选）"`
	Keyword        string `json:"keyword" remark:"搜索关键词（可选）"`
}

// ListProductsResp 获取商品列表响应
type ListProductsResp struct {
	Items []GetProductResp `json:"items" remark:"商品列表"`
	Total int64            `json:"total" remark:"总数"`
}

// GetCategoryReq 获取分类请求
type GetCategoryReq struct {
	ID uint32 `json:"id" remark:"分类ID"`
}

// GetCategoryResp 获取分类响应
type GetCategoryResp struct {
	ID          uint32 `json:"id" remark:"分类ID"`
	Name        string `json:"name" remark:"分类名称"`
	Description string `json:"description" remark:"分类描述"`
	ParentID    uint32 `json:"parent_id" remark:"父分类ID"`
	SortOrder   int32  `json:"sort_order" remark:"排序权重"`
	IsPublished bool   `json:"is_published" remark:"是否启用"`
	CreatedAt   int64  `json:"created_at" remark:"创建时间"`
	UpdatedAt   int64  `json:"updated_at" remark:"更新时间"`
}

// ListCountriesReq 获取国家列表请求
type ListCountriesReq struct {
	Page int32 `json:"page" remark:"页码"`
	Size int32 `json:"size" remark:"每页条数"`
}

// ListCountriesResp 获取国家列表响应
type ListCountriesResp struct {
	Items []CountryItem `json:"items" remark:"国家列表"`
	Total int64         `json:"total" remark:"总数"`
}

// CountryItem 国家条目
type CountryItem struct {
	ID               uint32 `json:"id" remark:"国家ID"`
	Name             string `json:"name" remark:"国家名称"`
	IsoCode2         string `json:"iso_code2" remark:"ISO 2字母代码"`
	IsoCode3         string `json:"iso_code3" remark:"ISO 3字母代码"`
	AddressFormat    string `json:"address_format" remark:"地址格式"`
	PostcodeRequired bool   `json:"postcode_required" remark:"是否必填邮编"`
	CreatedAt        string `json:"created_at" remark:"创建时间"`
	UpdatedAt        string `json:"updated_at" remark:"更新时间"`
}