// Package data 数据访问层。
// 包含商品、分类、制造商、评论等实体的持久化对象（PO）和仓储实现。
// PO 结构体同时包含 gorm 和 db(sqlx) tag，支持双 ORM 映射。
package data

import (
	"context"
	"time"

	"nop-go/services/product/internal/biz"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// 商品持久化对象
// ---------------------------------------------------------------------------

// ProductPO 商品持久化对象，映射 products 表。
// 包含商品基本信息、价格、库存、SKU 等核心字段。
type ProductPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name        string    `gorm:"size:256;not null;column:name" db:"name" json:"name"`               // 商品名称
	ShortDesc   string    `gorm:"size:512;column:short_desc" db:"short_desc" json:"short_desc"`       // 简短描述
	FullDesc    string    `gorm:"type:text;column:full_desc" db:"full_desc" json:"full_desc"`         // 完整描述（富文本）
	SKU         string    `gorm:"size:64;column:sku" db:"sku" json:"sku"`                             // 库存单位编码
	Price       float64   `gorm:"type:decimal(18,2);not null;column:price" db:"price" json:"price"`   // 商品价格
	OldPrice    float64   `gorm:"type:decimal(18,2);default:0;column:old_price" db:"old_price" json:"old_price"` // 原价（用于划线价展示）
	Cost        float64   `gorm:"type:decimal(18,2);default:0;column:cost" db:"cost" json:"cost"`     // 成本价
	Stock       int       `gorm:"default:0;column:stock" db:"stock" json:"stock"`                     // 库存数量
	CategoryID  uint      `gorm:"column:category_id;index" db:"category_id" json:"category_id"`       // 所属分类ID
	ManufacturerID uint   `gorm:"column:manufacturer_id;index" db:"manufacturer_id" json:"manufacturer_id"` // 制造商ID
	IsPublished bool      `gorm:"default:false;column:is_published" db:"is_published" json:"is_published"` // 是否上架
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 返回表名。
func (ProductPO) TableName() string {
	return "products"
}

// ToEntity 将 PO 转换为领域实体。
func (po *ProductPO) ToEntity() *biz.Product {
	return &biz.Product{
		ID:             po.ID,
		Name:           po.Name,
		ShortDesc:      po.ShortDesc,
		FullDesc:       po.FullDesc,
		SKU:            po.SKU,
		Price:          po.Price,
		OldPrice:       po.OldPrice,
		Cost:           po.Cost,
		Stock:          po.Stock,
		CategoryID:     po.CategoryID,
		ManufacturerID: po.ManufacturerID,
		IsPublished:    po.IsPublished,
		CreatedAt:      po.CreatedAt,
		UpdatedAt:      po.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// 分类持久化对象
// ---------------------------------------------------------------------------

// CategoryPO 分类持久化对象，映射 categories 表。
// 支持树形结构（ParentID），用于商品分类管理。
type CategoryPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name        string    `gorm:"size:128;not null;column:name" db:"name" json:"name"`             // 分类名称
	Description string    `gorm:"size:512;column:description" db:"description" json:"description"` // 分类描述
	ParentID    uint      `gorm:"default:0;column:parent_id;index" db:"parent_id" json:"parent_id"` // 父分类ID，0 表示顶级分类
	SortOrder   int       `gorm:"default:0;column:sort_order" db:"sort_order" json:"sort_order"`   // 排序权重
	IsPublished bool      `gorm:"default:true;column:is_published" db:"is_published" json:"is_published"` // 是否启用
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 返回表名。
func (CategoryPO) TableName() string {
	return "categories"
}

// ToEntity 将 PO 转换为领域实体。
func (po *CategoryPO) ToEntity() *biz.Category {
	return &biz.Category{
		ID:          po.ID,
		Name:        po.Name,
		Description: po.Description,
		ParentID:    po.ParentID,
		SortOrder:   po.SortOrder,
		IsPublished: po.IsPublished,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// 制造商持久化对象
// ---------------------------------------------------------------------------

// ManufacturerPO 制造商持久化对象，映射 manufacturers 表。
type ManufacturerPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name        string    `gorm:"size:128;not null;column:name" db:"name" json:"name"`             // 制造商名称
	Description string    `gorm:"size:512;column:description" db:"description" json:"description"` // 制造商描述
	IsPublished bool      `gorm:"default:true;column:is_published" db:"is_published" json:"is_published"` // 是否启用
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 返回表名。
func (ManufacturerPO) TableName() string {
	return "manufacturers"
}

// ToEntity 将 PO 转换为领域实体。
func (po *ManufacturerPO) ToEntity() *biz.Manufacturer {
	return &biz.Manufacturer{
		ID:          po.ID,
		Name:        po.Name,
		Description: po.Description,
		IsPublished: po.IsPublished,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// 商品评论持久化对象
// ---------------------------------------------------------------------------

// ProductReviewPO 商品评论持久化对象，映射 product_reviews 表。
type ProductReviewPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	ProductID   uint      `gorm:"not null;column:product_id;index" db:"product_id" json:"product_id"`     // 商品ID
	CustomerID  uint      `gorm:"not null;column:customer_id" db:"customer_id" json:"customer_id"`         // 评论者客户ID
	CustomerName string   `gorm:"size:64;column:customer_name" db:"customer_name" json:"customer_name"`   // 评论者名称（冗余，避免跨服务查询）
	Title       string    `gorm:"size:128;column:title" db:"title" json:"title"`                           // 评论标题
	Content     string    `gorm:"type:text;column:content" db:"content" json:"content"`                    // 评论内容
	Rating      int       `gorm:"default:5;column:rating" db:"rating" json:"rating"`                       // 评分 1-5
	IsApproved  bool      `gorm:"default:false;column:is_approved" db:"is_approved" json:"is_approved"`   // 是否审核通过
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 返回表名。
func (ProductReviewPO) TableName() string {
	return "product_reviews"
}

// ToEntity 将 PO 转换为领域实体。
func (po *ProductReviewPO) ToEntity() *biz.ProductReview {
	return &biz.ProductReview{
		ID:           po.ID,
		ProductID:    po.ProductID,
		CustomerID:   po.CustomerID,
		CustomerName: po.CustomerName,
		Title:        po.Title,
		Content:      po.Content,
		Rating:       po.Rating,
		IsApproved:   po.IsApproved,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// 最近浏览记录持久化对象
// ---------------------------------------------------------------------------

// RecentlyViewedPO 最近浏览商品记录，映射 recently_viewed 表。
// 用于记录用户浏览商品的历史，支持"最近浏览"功能。
type RecentlyViewedPO struct {
	ID         uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	CustomerID uint      `gorm:"not null;column:customer_id;index" db:"customer_id" json:"customer_id"` // 客户ID
	ProductID  uint      `gorm:"not null;column:product_id;index" db:"product_id" json:"product_id"`   // 商品ID
	ViewedAt   time.Time `gorm:"column:viewed_at" db:"viewed_at" json:"viewed_at"`                     // 浏览时间
}

// TableName 返回表名。
func (RecentlyViewedPO) TableName() string {
	return "recently_viewed"
}

// ---------------------------------------------------------------------------
// 仓储实现
// ---------------------------------------------------------------------------

// productRepo 商品仓储实现，基于 gorm。
type productRepo struct {
	db *gorm.DB
}

// NewProductRepo 创建商品仓储实例。
func NewProductRepo(db *gorm.DB) biz.ProductRepository {
	return &productRepo{db: db}
}

// Create 创建商品。
func (r *productRepo) Create(ctx context.Context, product *biz.Product) error {
	po := &ProductPO{
		Name:           product.Name,
		ShortDesc:      product.ShortDesc,
		FullDesc:       product.FullDesc,
		SKU:            product.SKU,
		Price:          product.Price,
		OldPrice:       product.OldPrice,
		Cost:           product.Cost,
		Stock:          product.Stock,
		CategoryID:     product.CategoryID,
		ManufacturerID: product.ManufacturerID,
		IsPublished:    product.IsPublished,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取商品。
func (r *productRepo) GetByID(ctx context.Context, id uint) (*biz.Product, error) {
	var po ProductPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新商品信息。
func (r *productRepo) Update(ctx context.Context, product *biz.Product) error {
	po := &ProductPO{
		Name:           product.Name,
		ShortDesc:      product.ShortDesc,
		FullDesc:       product.FullDesc,
		SKU:            product.SKU,
		Price:          product.Price,
		OldPrice:       product.OldPrice,
		Cost:           product.Cost,
		Stock:          product.Stock,
		CategoryID:     product.CategoryID,
		ManufacturerID: product.ManufacturerID,
		IsPublished:    product.IsPublished,
	}
	return r.db.WithContext(ctx).Model(&ProductPO{}).Where("id = ?", product.ID).Updates(po).Error
}

// List 获取商品列表，支持按分类、制造商、关键词筛选。
func (r *productRepo) List(ctx context.Context, page, size int, categoryID, manufacturerID uint, keyword string) ([]*biz.Product, int64, error) {
	var pos []ProductPO
	var total int64

	query := r.db.WithContext(ctx).Model(&ProductPO{})

	// 按分类筛选
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	// 按制造商筛选
	if manufacturerID > 0 {
		query = query.Where("manufacturer_id = ?", manufacturerID)
	}
	// 按关键词模糊搜索（商品名称或SKU）
	if keyword != "" {
		query = query.Where("name LIKE ? OR sku LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	products := make([]*biz.Product, len(pos))
	for i, po := range pos {
		products[i] = po.ToEntity()
	}
	return products, total, nil
}

// Delete 删除商品。
func (r *productRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ProductPO{}, id).Error
}

// ---------------------------------------------------------------------------
// 分类仓储实现
// ---------------------------------------------------------------------------

// categoryRepo 分类仓储实现。
type categoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepo 创建分类仓储实例。
func NewCategoryRepo(db *gorm.DB) biz.CategoryRepository {
	return &categoryRepo{db: db}
}

// Create 创建分类。
func (r *categoryRepo) Create(ctx context.Context, category *biz.Category) error {
	po := &CategoryPO{
		Name:        category.Name,
		Description: category.Description,
		ParentID:    category.ParentID,
		SortOrder:   category.SortOrder,
		IsPublished: category.IsPublished,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取分类。
func (r *categoryRepo) GetByID(ctx context.Context, id uint) (*biz.Category, error) {
	var po CategoryPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取分类列表。
func (r *categoryRepo) List(ctx context.Context, page, size int) ([]*biz.Category, int64, error) {
	var pos []CategoryPO
	var total int64

	r.db.WithContext(ctx).Model(&CategoryPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	categories := make([]*biz.Category, len(pos))
	for i, po := range pos {
		categories[i] = po.ToEntity()
	}
	return categories, total, nil
}

// ---------------------------------------------------------------------------
// 制造商仓储实现
// ---------------------------------------------------------------------------

// manufacturerRepo 制造商仓储实现。
type manufacturerRepo struct {
	db *gorm.DB
}

// NewManufacturerRepo 创建制造商仓储实例。
func NewManufacturerRepo(db *gorm.DB) biz.ManufacturerRepository {
	return &manufacturerRepo{db: db}
}

// Create 创建制造商。
func (r *manufacturerRepo) Create(ctx context.Context, manufacturer *biz.Manufacturer) error {
	po := &ManufacturerPO{
		Name:        manufacturer.Name,
		Description: manufacturer.Description,
		IsPublished: manufacturer.IsPublished,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取制造商。
func (r *manufacturerRepo) GetByID(ctx context.Context, id uint) (*biz.Manufacturer, error) {
	var po ManufacturerPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取制造商列表。
func (r *manufacturerRepo) List(ctx context.Context, page, size int) ([]*biz.Manufacturer, int64, error) {
	var pos []ManufacturerPO
	var total int64

	r.db.WithContext(ctx).Model(&ManufacturerPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id ASC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	manufacturers := make([]*biz.Manufacturer, len(pos))
	for i, po := range pos {
		manufacturers[i] = po.ToEntity()
	}
	return manufacturers, total, nil
}

// ---------------------------------------------------------------------------
// 商品评论仓储实现
// ---------------------------------------------------------------------------

// productReviewRepo 商品评论仓储实现。
type productReviewRepo struct {
	db *gorm.DB
}

// NewProductReviewRepo 创建商品评论仓储实例。
func NewProductReviewRepo(db *gorm.DB) biz.ProductReviewRepository {
	return &productReviewRepo{db: db}
}

// Create 创建商品评论。
func (r *productReviewRepo) Create(ctx context.Context, review *biz.ProductReview) error {
	po := &ProductReviewPO{
		ProductID:    review.ProductID,
		CustomerID:   review.CustomerID,
		CustomerName: review.CustomerName,
		Title:        review.Title,
		Content:      review.Content,
		Rating:       review.Rating,
		IsApproved:   review.IsApproved,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// ListByProductID 获取指定商品的评论列表。
func (r *productReviewRepo) ListByProductID(ctx context.Context, productID uint, page, size int) ([]*biz.ProductReview, int64, error) {
	var pos []ProductReviewPO
	var total int64

	query := r.db.WithContext(ctx).Model(&ProductReviewPO{}).Where("product_id = ?", productID)
	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	reviews := make([]*biz.ProductReview, len(pos))
	for i, po := range pos {
		reviews[i] = po.ToEntity()
	}
	return reviews, total, nil
}

// ---------------------------------------------------------------------------
// 最近浏览仓储实现
// ---------------------------------------------------------------------------

// recentlyViewedRepo 最近浏览仓储实现。
type recentlyViewedRepo struct {
	db *gorm.DB
}

// NewRecentlyViewedRepo 创建最近浏览仓储实例。
func NewRecentlyViewedRepo(db *gorm.DB) biz.RecentlyViewedRepository {
	return &recentlyViewedRepo{db: db}
}

// Record 记录一次浏览行为。若已存在则更新浏览时间。
func (r *recentlyViewedRepo) Record(ctx context.Context, customerID, productID uint) error {
	var existing RecentlyViewedPO
	err := r.db.WithContext(ctx).
		Where("customer_id = ? AND product_id = ?", customerID, productID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 新增浏览记录
		return r.db.WithContext(ctx).Create(&RecentlyViewedPO{
			CustomerID: customerID,
			ProductID:  productID,
			ViewedAt:   time.Now(),
		}).Error
	}
	if err != nil {
		return err
	}
	// 已存在则更新浏览时间
	return r.db.WithContext(ctx).Model(&existing).Update("viewed_at", time.Now()).Error
}

// ListByCustomerID 获取指定客户的最近浏览商品ID列表。
func (r *recentlyViewedRepo) ListByCustomerID(ctx context.Context, customerID uint, limit int) ([]uint, error) {
	var productIDs []uint
	err := r.db.WithContext(ctx).
		Model(&RecentlyViewedPO{}).
		Where("customer_id = ?", customerID).
		Order("viewed_at DESC").
		Limit(limit).
		Pluck("product_id", &productIDs).Error
	return productIDs, err
}
