// Package data 商品目录数据访问层
//
// 中文说明:
// - 定义商品、分类、品牌等仓储接口;
// - 使用 GORM 实现数据持久化;
// - 支持复杂的商品查询和筛选。
package data

import (
	"context"
	"errors"

	"nop-go/services/catalog-service/internal/models"

	"gorm.io/gorm"
)

// ProductRepository 商品仓储接口
//
// 中文说明:
// - 定义商品数据访问的抽象接口;
// - 支持基本的 CRUD 操作和复杂查询;
// - 便于业务层解耦和测试。
type ProductRepository interface {
	// Create 创建商品
	Create(ctx context.Context, product *models.Product) error
	// GetByID 根据ID获取商品
	GetByID(ctx context.Context, id uint64) (*models.Product, error)
	// GetBySKU 根据SKU获取商品
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	// Update 更新商品
	Update(ctx context.Context, product *models.Product) error
	// Delete 删除商品 (软删除)
	Delete(ctx context.Context, id uint64) error
	// List 商品列表 (支持筛选)
	List(ctx context.Context, req *models.ProductListRequest) ([]*models.Product, int64, error)
	// GetByCategoryID 根据分类获取商品
	GetByCategoryID(ctx context.Context, categoryID uint64, page, pageSize int) ([]*models.Product, int64, error)
	// GetByManufacturerID 根据品牌获取商品
	GetByManufacturerID(ctx context.Context, manufacturerID uint64, page, pageSize int) ([]*models.Product, int64, error)
	// GetHomepageProducts 获取首页展示商品
	GetHomepageProducts(ctx context.Context, limit int) ([]*models.Product, error)
	// AddCategory 添加商品分类
	AddCategory(ctx context.Context, productID, categoryID uint64, isPrimary bool) error
	// RemoveCategory 移除商品分类
	RemoveCategory(ctx context.Context, productID, categoryID uint64) error
}

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	// Create 创建分类
	Create(ctx context.Context, category *models.Category) error
	// GetByID 根据ID获取分类
	GetByID(ctx context.Context, id uint64) (*models.Category, error)
	// GetBySlug 根据Slug获取分类
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	// Update 更新分类
	Update(ctx context.Context, category *models.Category) error
	// Delete 删除分类
	Delete(ctx context.Context, id uint64) error
	// List 分类列表
	List(ctx context.Context) ([]*models.Category, error)
	// GetTree 获取分类树
	GetTree(ctx context.Context) ([]*models.Category, error)
	// GetChildren 获取子分类
	GetChildren(ctx context.Context, parentID uint64) ([]*models.Category, error)
}

// ManufacturerRepository 品牌/制造商仓储接口
type ManufacturerRepository interface {
	// Create 创建品牌
	Create(ctx context.Context, manufacturer *models.Manufacturer) error
	// GetByID 根据ID获取品牌
	GetByID(ctx context.Context, id uint64) (*models.Manufacturer, error)
	// GetBySlug 根据Slug获取品牌
	GetBySlug(ctx context.Context, slug string) (*models.Manufacturer, error)
	// Update 更新品牌
	Update(ctx context.Context, manufacturer *models.Manufacturer) error
	// Delete 删除品牌
	Delete(ctx context.Context, id uint64) error
	// List 品牌列表
	List(ctx context.Context) ([]*models.Manufacturer, error)
}

// ProductPictureRepository 商品图片仓储接口
type ProductPictureRepository interface {
	// Create 创建图片
	Create(ctx context.Context, picture *models.ProductPicture) error
	// GetByID 根据ID获取图片
	GetByID(ctx context.Context, id uint64) (*models.ProductPicture, error)
	// GetByProductID 获取商品所有图片
	GetByProductID(ctx context.Context, productID uint64) ([]*models.ProductPicture, error)
	// Update 更新图片
	Update(ctx context.Context, picture *models.ProductPicture) error
	// Delete 删除图片
	Delete(ctx context.Context, id uint64) error
	// SetMain 设置主图
	SetMain(ctx context.Context, productID, pictureID uint64) error
}

// ProductReviewRepository 商品评论仓储接口
type ProductReviewRepository interface {
	// Create 创建评论
	Create(ctx context.Context, review *models.ProductReview) error
	// GetByID 根据ID获取评论
	GetByID(ctx context.Context, id uint64) (*models.ProductReview, error)
	// GetByProductID 获取商品评论
	GetByProductID(ctx context.Context, productID uint64, page, pageSize int) ([]*models.ProductReview, int64, error)
	// Update 更新评论
	Update(ctx context.Context, review *models.ProductReview) error
	// Delete 删除评论
	Delete(ctx context.Context, id uint64) error
	// Approve 批准评论
	Approve(ctx context.Context, id uint64) error
	// MarkHelpful 标记有用
	MarkHelpful(ctx context.Context, id uint64, helpful bool) error
}

// productRepo 商品仓储实现
type productRepo struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepo) GetByID(ctx context.Context, id uint64) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Categories").
		Preload("Pictures").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Preload("Manufacturer").
		Preload("Categories").
		Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) Update(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}

func (r *productRepo) List(ctx context.Context, req *models.ProductListRequest) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Product{})

	// 筛选条件
	if req.CategoryID != nil {
		db = db.Joins("JOIN product_categories ON product_categories.product_id = products.id").
			Where("product_categories.category_id = ?", *req.CategoryID)
	}
	if req.ManufacturerID != nil {
		db = db.Where("manufacturer_id = ?", *req.ManufacturerID)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR sku LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.IsPublished != nil {
		db = db.Where("is_published = ?", *req.IsPublished)
	}

	db.Count(&total)

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Manufacturer").
		Preload("Categories").
		Preload("Pictures").
		Offset(offset).Limit(pageSize).
		Order("display_order ASC, created_at DESC").
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepo) GetByCategoryID(ctx context.Context, categoryID uint64, page, pageSize int) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Product{}).
		Joins("JOIN product_categories ON product_categories.product_id = products.id").
		Where("product_categories.category_id = ?", categoryID)

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Preload("Manufacturer").
		Preload("Pictures").
		Offset(offset).Limit(pageSize).
		Order("product_categories.display_order ASC, products.display_order ASC").
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepo) GetByManufacturerID(ctx context.Context, manufacturerID uint64, page, pageSize int) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("manufacturer_id = ?", manufacturerID)

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Preload("Categories").
		Preload("Pictures").
		Offset(offset).Limit(pageSize).
		Order("display_order ASC, created_at DESC").
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepo) GetHomepageProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.WithContext(ctx).
		Where("show_on_homepage = ? AND is_published = ?", true, true).
		Preload("Manufacturer").
		Preload("Pictures").
		Order("display_order ASC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

func (r *productRepo) AddCategory(ctx context.Context, productID, categoryID uint64, isPrimary bool) error {
	pc := models.ProductCategory{
		ProductID:  productID,
		CategoryID: categoryID,
		IsPrimary:  isPrimary,
	}
	return r.db.WithContext(ctx).Create(&pc).Error
}

func (r *productRepo) RemoveCategory(ctx context.Context, productID, categoryID uint64) error {
	return r.db.WithContext(ctx).
		Where("product_id = ? AND category_id = ?", productID, categoryID).
		Delete(&models.ProductCategory{}).Error
}

// categoryRepo 分类仓储实现
type categoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepo) GetByID(ctx context.Context, id uint64) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepo) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Where("seo_slug = ?", slug).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepo) Update(ctx context.Context, category *models.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error
}

func (r *categoryRepo) List(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("is_published = ?", true).
		Order("level ASC, display_order ASC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) GetTree(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("is_published = ?", true).
		Order("level ASC, display_order ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	tree := buildCategoryTree(categories, nil)
	return tree, nil
}

func (r *categoryRepo) GetChildren(ctx context.Context, parentID uint64) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND is_published = ?", parentID, true).
		Order("display_order ASC").
		Find(&categories).Error
	return categories, err
}

// buildCategoryTree 构建分类树
func buildCategoryTree(categories []*models.Category, parentID *uint64) []*models.Category {
	var result []*models.Category
	for _, c := range categories {
		if (parentID == nil && c.ParentID == nil) || (parentID != nil && c.ParentID != nil && *c.ParentID == *parentID) {
			children := buildCategoryTree(categories, &c.ID)
			// 转换为值类型赋值给 Children 字段
			c.Children = make([]models.Category, len(children))
			for i, child := range children {
				c.Children[i] = *child
			}
			result = append(result, c)
		}
	}
	return result
}

// manufacturerRepo 品牌仓储实现
type manufacturerRepo struct {
	db *gorm.DB
}

// NewManufacturerRepository 创建品牌仓储
func NewManufacturerRepository(db *gorm.DB) ManufacturerRepository {
	return &manufacturerRepo{db: db}
}

func (r *manufacturerRepo) Create(ctx context.Context, manufacturer *models.Manufacturer) error {
	return r.db.WithContext(ctx).Create(manufacturer).Error
}

func (r *manufacturerRepo) GetByID(ctx context.Context, id uint64) (*models.Manufacturer, error) {
	var manufacturer models.Manufacturer
	err := r.db.WithContext(ctx).First(&manufacturer, id).Error
	if err != nil {
		return nil, err
	}
	return &manufacturer, nil
}

func (r *manufacturerRepo) GetBySlug(ctx context.Context, slug string) (*models.Manufacturer, error) {
	var manufacturer models.Manufacturer
	err := r.db.WithContext(ctx).
		Where("seo_slug = ?", slug).
		First(&manufacturer).Error
	if err != nil {
		return nil, err
	}
	return &manufacturer, nil
}

func (r *manufacturerRepo) Update(ctx context.Context, manufacturer *models.Manufacturer) error {
	return r.db.WithContext(ctx).Save(manufacturer).Error
}

func (r *manufacturerRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Manufacturer{}, id).Error
}

func (r *manufacturerRepo) List(ctx context.Context) ([]*models.Manufacturer, error) {
	var manufacturers []*models.Manufacturer
	err := r.db.WithContext(ctx).
		Where("is_published = ?", true).
		Order("display_order ASC").
		Find(&manufacturers).Error
	return manufacturers, err
}

// productPictureRepo 商品图片仓储实现
type productPictureRepo struct {
	db *gorm.DB
}

// NewProductPictureRepository 创建商品图片仓储
func NewProductPictureRepository(db *gorm.DB) ProductPictureRepository {
	return &productPictureRepo{db: db}
}

func (r *productPictureRepo) Create(ctx context.Context, picture *models.ProductPicture) error {
	return r.db.WithContext(ctx).Create(picture).Error
}

func (r *productPictureRepo) GetByID(ctx context.Context, id uint64) (*models.ProductPicture, error) {
	var picture models.ProductPicture
	err := r.db.WithContext(ctx).First(&picture, id).Error
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func (r *productPictureRepo) GetByProductID(ctx context.Context, productID uint64) ([]*models.ProductPicture, error) {
	var pictures []*models.ProductPicture
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("display_order ASC").
		Find(&pictures).Error
	return pictures, err
}

func (r *productPictureRepo) Update(ctx context.Context, picture *models.ProductPicture) error {
	return r.db.WithContext(ctx).Save(picture).Error
}

func (r *productPictureRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.ProductPicture{}, id).Error
}

func (r *productPictureRepo) SetMain(ctx context.Context, productID, pictureID uint64) error {
	// 先清除所有主图
	if err := r.db.WithContext(ctx).Model(&models.ProductPicture{}).
		Where("product_id = ?", productID).
		Update("is_main", false).Error; err != nil {
		return err
	}
	// 设置指定图片为主图
	return r.db.WithContext(ctx).Model(&models.ProductPicture{}).
		Where("id = ?", pictureID).
		Update("is_main", true).Error
}

// productReviewRepo 商品评论仓储实现
type productReviewRepo struct {
	db *gorm.DB
}

// NewProductReviewRepository 创建商品评论仓储
func NewProductReviewRepository(db *gorm.DB) ProductReviewRepository {
	return &productReviewRepo{db: db}
}

func (r *productReviewRepo) Create(ctx context.Context, review *models.ProductReview) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *productReviewRepo) GetByID(ctx context.Context, id uint64) (*models.ProductReview, error) {
	var review models.ProductReview
	err := r.db.WithContext(ctx).First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *productReviewRepo) GetByProductID(ctx context.Context, productID uint64, page, pageSize int) ([]*models.ProductReview, int64, error) {
	var reviews []*models.ProductReview
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ProductReview{}).
		Where("product_id = ? AND is_approved = ?", productID, true)

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (r *productReviewRepo) Update(ctx context.Context, review *models.ProductReview) error {
	return r.db.WithContext(ctx).Save(review).Error
}

func (r *productReviewRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.ProductReview{}, id).Error
}

func (r *productReviewRepo) Approve(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&models.ProductReview{}).
		Where("id = ?", id).
		Update("is_approved", true).Error
}

func (r *productReviewRepo) MarkHelpful(ctx context.Context, id uint64, helpful bool) error {
	if helpful {
		return r.db.WithContext(ctx).Model(&models.ProductReview{}).
			Where("id = ?", id).
			UpdateColumn("helpful_yes", gorm.Expr("helpful_yes + 1")).Error
	}
	return r.db.WithContext(ctx).Model(&models.ProductReview{}).
		Where("id = ?", id).
		UpdateColumn("helpful_no", gorm.Expr("helpful_no + 1")).Error
}

// IsErrProductNotFound 判断是否为商品不存在错误
func IsErrProductNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}