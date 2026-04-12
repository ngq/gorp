// Package models 商品目录数据模型
//
// 中文说明:
// - 定义商品、分类、品牌等核心实体;
// - 对应 nopCommerce 的 Catalog 模块;
// - 支持商品属性、SKU、图片、评论等功能。
package models

import (
	"time"

	shareModels "nop-go/shared/models"
)

// Product 商品实体
//
// 中文说明:
// - 对应 nopCommerce Product 表;
// - 支持简单商品和组合商品两种类型;
// - 包含商品基本信息、SEO 信息、发布状态等。
type Product struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	SKU              string    `gorm:"size:64;uniqueIndex;not null" json:"sku"`
	Name             string    `gorm:"size:256;not null" json:"name"`
	ShortDescription string    `gorm:"size:512" json:"short_description"`
	FullDescription  string    `gorm:"type:text" json:"full_description"`
	ManufacturerID   *uint64   `gorm:"index" json:"manufacturer_id"`
	ProductType      string    `gorm:"size:16;default:'simple'" json:"product_type"` // simple, grouped
	IsPublished      bool      `gorm:"default:false" json:"is_published"`
	ShowOnHomepage   bool      `gorm:"default:false" json:"show_on_homepage"`
	DisplayOrder     int       `gorm:"default:0" json:"display_order"`
	SEOSlug          string    `gorm:"size:256" json:"seo_slug"`
	MetaKeywords     string    `gorm:"size:256" json:"meta_keywords"`
	MetaDescription  string    `gorm:"size:256" json:"meta_description"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        *time.Time `gorm:"index" json:"deleted_at"`

	// 关联
	Manufacturer    *Manufacturer     `gorm:"foreignKey:ManufacturerID" json:"manufacturer,omitempty"`
	Categories      []Category        `gorm:"many2many:product_categories;" json:"categories,omitempty"`
	Pictures        []ProductPicture  `gorm:"foreignKey:ProductID" json:"pictures,omitempty"`
	AttributeMaps   []ProductAttributeMapping `gorm:"foreignKey:ProductID" json:"attribute_maps,omitempty"`
	Reviews         []ProductReview   `gorm:"foreignKey:ProductID" json:"reviews,omitempty"`
}

// TableName 表名
func (Product) TableName() string {
	return "products"
}

// Category 商品分类
//
// 中文说明:
// - 对应 nopCommerce Category 表;
// - 支持多级分类(树形结构);
// - 包含分类基本信息和 SEO 信息。
type Category struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:256;not null" json:"name"`
	ParentID     *uint64   `gorm:"index" json:"parent_id"`
	Level        int       `gorm:"not null;default:1" json:"level"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsPublished  bool      `gorm:"default:true" json:"is_published"`
	SEOSlug      string    `gorm:"size:256" json:"seo_slug"`
	Description  string    `gorm:"type:text" json:"description"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Parent   *Category   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Products []Product   `gorm:"many2many:product_categories;" json:"products,omitempty"`
}

// TableName 表名
func (Category) TableName() string {
	return "categories"
}

// Manufacturer 品牌/制造商
//
// 中文说明:
// - 对应 nopCommerce Manufacturer 表;
// - 用于商品品牌分类;
// - 支持品牌 Logo 和 SEO 信息。
type Manufacturer struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:256;not null;uniqueIndex" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	LogoURL      string    `gorm:"size:512" json:"logo_url"`
	IsPublished  bool      `gorm:"default:true" json:"is_published"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	SEOSlug      string    `gorm:"size:256" json:"seo_slug"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Products []Product `gorm:"foreignKey:ManufacturerID" json:"products,omitempty"`
}

// TableName 表名
func (Manufacturer) TableName() string {
	return "manufacturers"
}

// ProductCategory 商品分类关联
//
// 中文说明:
// - 商品和分类的多对多关系表;
// - 支持设置主分类和排序。
type ProductCategory struct {
	ProductID    uint64 `gorm:"not null" json:"product_id"`
	CategoryID   uint64 `gorm:"not null" json:"category_id"`
	IsPrimary    bool   `gorm:"default:false" json:"is_primary"`
	DisplayOrder int    `gorm:"default:0" json:"display_order"`

	// 关联
	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Category  *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName 表名
func (ProductCategory) TableName() string {
	return "product_categories"
}

// ProductAttribute 商品属性
//
// 中文说明:
// - 对应 nopCommerce ProductAttribute 表;
// - 定义商品的可选属性(颜色、尺寸等);
// - 一个属性可以有多个可选值。
type ProductAttribute struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Values []ProductAttributeValue `gorm:"foreignKey:ProductAttributeID" json:"values,omitempty"`
}

// TableName 表名
func (ProductAttribute) TableName() string {
	return "product_attributes"
}

// ProductAttributeValue 商品属性值
//
// 中文说明:
// - 属性的可选值(如颜色属性下的红色、蓝色等);
// - 支持价格调整和重量调整;
// - 可设置默认选中状态。
type ProductAttributeValue struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	ProductAttributeID uint64   `gorm:"not null;index" json:"product_attribute_id"`
	Name              string    `gorm:"size:128;not null" json:"name"`
	PriceAdjustment   float64   `gorm:"type:decimal(10,2);default:0" json:"price_adjustment"`
	WeightAdjustment  float64   `gorm:"type:decimal(10,3);default:0" json:"weight_adjustment"`
	IsPreSelected     bool      `gorm:"default:false" json:"is_pre_selected"`
	DisplayOrder      int       `gorm:"default:0" json:"display_order"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Attribute *ProductAttribute `gorm:"foreignKey:ProductAttributeID" json:"attribute,omitempty"`
}

// TableName 表名
func (ProductAttributeValue) TableName() string {
	return "product_attribute_values"
}

// ProductAttributeMapping 商品属性映射
//
// 中文说明:
// - 商品和属性的关系;
// - 定义某商品使用哪些属性;
// - 设置属性是否必填、控件类型等。
type ProductAttributeMapping struct {
	ID                   uint64    `gorm:"primaryKey" json:"id"`
	ProductID            uint64    `gorm:"not null;index" json:"product_id"`
	ProductAttributeID   uint64    `gorm:"not null;index" json:"product_attribute_id"`
	IsRequired           bool      `gorm:"default:false" json:"is_required"`
	AttributeControlType string    `gorm:"size:32;default:'dropdown'" json:"attribute_control_type"` // dropdown, radio, checkbox, textbox, etc.
	DisplayOrder         int       `gorm:"default:0" json:"display_order"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Product   *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Attribute *ProductAttribute `gorm:"foreignKey:ProductAttributeID" json:"attribute,omitempty"`
}

// TableName 表名
func (ProductAttributeMapping) TableName() string {
	return "product_attribute_mappings"
}

// ProductPicture 商品图片
//
// 中文说明:
// - 商品的图片记录;
// - 支持排序和主图设置;
// - 存储图片 URL 和 Alt 文本。
type ProductPicture struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	ProductID    uint64    `gorm:"not null;index" json:"product_id"`
	PictureURL   string    `gorm:"size:512;not null" json:"picture_url"`
	AltText      string    `gorm:"size:256" json:"alt_text"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsMain       bool      `gorm:"default:false" json:"is_main"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 表名
func (ProductPicture) TableName() string {
	return "product_pictures"
}

// ProductReview 商品评论
//
// 中文说明:
// - 用户对商品的评论;
// - 支持 1-5 星评分;
// - 包含审核状态和有用度统计。
type ProductReview struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	ProductID   uint64    `gorm:"not null;index" json:"product_id"`
	CustomerID  uint64    `gorm:"not null;index" json:"customer_id"`
	Rating      int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Title       string    `gorm:"size:256" json:"title"`
	ReviewText  string    `gorm:"type:text" json:"review_text"`
	IsApproved  bool      `gorm:"default:false" json:"is_approved"`
	HelpfulYes  int       `gorm:"default:0" json:"helpful_yes"`
	HelpfulNo   int       `gorm:"default:0" json:"helpful_no"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 表名
func (ProductReview) TableName() string {
	return "product_reviews"
}

// ProductSpecificationAttribute 商品规格属性
//
// 中文说明:
// - 对应 nopCommerce SpecificationAttribute;
// - 用于商品筛选和对比(如屏幕尺寸、内存大小等);
// - 与商品属性不同,规格属性用于筛选而非选择。
type ProductSpecificationAttribute struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	DisplayOrder int      `gorm:"default:0" json:"display_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Options []ProductSpecificationAttributeOption `gorm:"foreignKey:SpecificationAttributeID" json:"options,omitempty"`
}

// TableName 表名
func (ProductSpecificationAttribute) TableName() string {
	return "product_specification_attributes"
}

// ProductSpecificationAttributeOption 规格属性选项
type ProductSpecificationAttributeOption struct {
	ID                    uint64    `gorm:"primaryKey" json:"id"`
	SpecificationAttributeID uint64 `gorm:"not null;index" json:"specification_attribute_id"`
	Name                  string    `gorm:"size:128;not null" json:"name"`
	DisplayOrder          int       `gorm:"default:0" json:"display_order"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (ProductSpecificationAttributeOption) TableName() string {
	return "product_specification_attribute_options"
}

// ProductSpecificationAttributeMapping 商品规格映射
type ProductSpecificationAttributeMapping struct {
	ID                           uint64    `gorm:"primaryKey" json:"id"`
	ProductID                    uint64    `gorm:"not null;index" json:"product_id"`
	SpecificationAttributeOptionID uint64  `gorm:"not null;index" json:"specification_attribute_option_id"`
	IsVisibleOnProductPage       bool      `gorm:"default:true" json:"is_visible_on_product_page"`
	DisplayOrder                 int       `gorm:"default:0" json:"display_order"`
	CreatedAt                    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (ProductSpecificationAttributeMapping) TableName() string {
	return "product_specification_attribute_mappings"
}

// RelatedProduct 关联商品
//
// 中文说明:
// - 商品之间的关联推荐;
// - 如搭配购买、相关商品等。
type RelatedProduct struct {
	ProductID1       uint64 `gorm:"not null" json:"product_id_1"`
	ProductID2       uint64 `gorm:"not null" json:"product_id_2"`
	DisplayOrder     int    `gorm:"default:0" json:"display_order"`
	IsBidirectional  bool   `gorm:"default:false" json:"is_bidirectional"`

	// 关联
	Product1 *Product `gorm:"foreignKey:ProductID1" json:"product1,omitempty"`
	Product2 *Product `gorm:"foreignKey:ProductID2" json:"product2,omitempty"`
}

// TableName 表名
func (RelatedProduct) TableName() string {
	return "related_products"
}

// DTO 定义

// ProductListRequest 商品列表请求
type ProductListRequest struct {
	CategoryID     *uint64 `json:"category_id"`
	ManufacturerID *uint64 `json:"manufacturer_id"`
	Keyword        string  `json:"keyword"`
	IsPublished    *bool   `json:"is_published"`
	Page           int     `json:"page"`
	PageSize       int     `json:"page_size"`
}

// ProductResponse 商品响应
type ProductResponse struct {
	ID               uint64             `json:"id"`
	SKU              string             `json:"sku"`
	Name             string             `json:"name"`
	ShortDescription string             `json:"short_description"`
	FullDescription  string             `json:"full_description"`
	ProductType      string             `json:"product_type"`
	IsPublished      bool               `json:"is_published"`
	ShowOnHomepage   bool               `json:"show_on_homepage"`
	Manufacturer     *ManufacturerInfo  `json:"manufacturer,omitempty"`
	Categories       []CategoryInfo     `json:"categories,omitempty"`
	MainPicture      *ProductPictureInfo `json:"main_picture,omitempty"`
	Pictures         []ProductPictureInfo `json:"pictures,omitempty"`
	SEOSlug          string             `json:"seo_slug"`
	CreatedAt        string             `json:"created_at"`
}

// ManufacturerInfo 品牌简要信息
type ManufacturerInfo struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// CategoryInfo 分类简要信息
type CategoryInfo struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// ProductPictureInfo 图片简要信息
type ProductPictureInfo struct {
	ID         uint64 `json:"id"`
	PictureURL string `json:"picture_url"`
	AltText    string `json:"alt_text"`
	IsMain     bool   `json:"is_main"`
}

// ToProductResponse 转换为商品响应
func ToProductResponse(p *Product) ProductResponse {
	resp := ProductResponse{
		ID:               p.ID,
		SKU:              p.SKU,
		Name:             p.Name,
		ShortDescription: p.ShortDescription,
		FullDescription:  p.FullDescription,
		ProductType:      p.ProductType,
		IsPublished:      p.IsPublished,
		ShowOnHomepage:   p.ShowOnHomepage,
		SEOSlug:          p.SEOSlug,
		CreatedAt:        p.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 品牌信息
	if p.Manufacturer != nil {
		resp.Manufacturer = &ManufacturerInfo{
			ID:   p.Manufacturer.ID,
			Name: p.Manufacturer.Name,
		}
	}

	// 分类信息
	if len(p.Categories) > 0 {
		resp.Categories = make([]CategoryInfo, len(p.Categories))
		for i, c := range p.Categories {
			resp.Categories[i] = CategoryInfo{ID: c.ID, Name: c.Name}
		}
	}

	// 图片信息
	if len(p.Pictures) > 0 {
		resp.Pictures = make([]ProductPictureInfo, len(p.Pictures))
		for i, pic := range p.Pictures {
			resp.Pictures[i] = ProductPictureInfo{
				ID:         pic.ID,
				PictureURL: pic.PictureURL,
				AltText:    pic.AltText,
				IsMain:     pic.IsMain,
			}
			if pic.IsMain {
				resp.MainPicture = &resp.Pictures[i]
			}
		}
	}

	return resp
}

// CategoryTreeResponse 分类树响应
type CategoryTreeResponse struct {
	ID           uint64                `json:"id"`
	Name         string                `json:"name"`
	Level        int                   `json:"level"`
	DisplayOrder int                   `json:"display_order"`
	IsPublished  bool                  `json:"is_published"`
	Children     []CategoryTreeResponse `json:"children,omitempty"`
}

// ToCategoryTreeResponse 转换为分类树响应
func ToCategoryTreeResponse(c *Category) CategoryTreeResponse {
	resp := CategoryTreeResponse{
		ID:           c.ID,
		Name:         c.Name,
		Level:        c.Level,
		DisplayOrder: c.DisplayOrder,
		IsPublished:  c.IsPublished,
	}

	if len(c.Children) > 0 {
		resp.Children = make([]CategoryTreeResponse, len(c.Children))
		for i, child := range c.Children {
			resp.Children[i] = ToCategoryTreeResponse(&child)
		}
	}

	return resp
}

// ProductReviewResponse 商品评论响应
type ProductReviewResponse struct {
	ID         uint64 `json:"id"`
	ProductID  uint64 `json:"product_id"`
	CustomerID uint64 `json:"customer_id"`
	Rating     int    `json:"rating"`
	Title      string `json:"title"`
	ReviewText string `json:"review_text"`
	IsApproved bool   `json:"is_approved"`
	HelpfulYes int    `json:"helpful_yes"`
	HelpfulNo  int    `json:"helpful_no"`
	CreatedAt  string `json:"created_at"`
}

// ToProductReviewResponse 转换为评论响应
func ToProductReviewResponse(r *ProductReview) ProductReviewResponse {
	return ProductReviewResponse{
		ID:         r.ID,
		ProductID:  r.ProductID,
		CustomerID: r.CustomerID,
		Rating:     r.Rating,
		Title:      r.Title,
		ReviewText: r.ReviewText,
		IsApproved: r.IsApproved,
		HelpfulYes: r.HelpfulYes,
		HelpfulNo:  r.HelpfulNo,
		CreatedAt:  r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// PagingResponse 分页响应
type PagingResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPagingResponse 创建分页响应
func NewPagingResponse(items interface{}, total int64, page, pageSize int) *shareModels.PagingResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &shareModels.PagingResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}