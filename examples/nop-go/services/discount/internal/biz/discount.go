// Package biz 提供 discount 服务的业务逻辑层
//
// 折扣服务包含以下子领域：
// 1. 折扣（Discount）— 折扣主体 CRUD
// 2. 折扣关联商品（DiscountProduct）— 折扣适用的商品
// 3. 折扣关联分类（DiscountCategory）— 折扣适用的分类
// 4. 折扣关联制造商（DiscountManufacturer）— 折扣适用的制造商
// 5. 折扣使用历史（DiscountUsageHistory）— 折扣使用记录
package biz

import (
	"context"
	"time"

	"nop-go/services/discount/internal/server/http/request"
	"nop-go/services/discount/internal/server/http/response"
)

// ==================== 领域实体定义 ====================

// Discount 折扣领域实体
type Discount struct {
	ID                uint      // 折扣 ID
	Name              string    // 折扣名称
	DiscountType      string    // 折扣类型（percentage/fixed/free_shipping）
	DiscountAmount    float64   // 折扣金额/百分比
	StartDate         time.Time // 折扣开始日期
	EndDate           time.Time // 折扣结束日期
	RequiresCouponCode bool     // 是否需要优惠券码
	CouponCode        string    // 优惠券码
	IsCumulative      bool      // 是否可叠加使用
	DisplayOrder      int       // 显示排序
	IsActive          bool      // 是否启用
	LimitationTimes   int       // 使用次数限制
	CreatedAt         time.Time // 创建时间
	UpdatedAt         time.Time // 更新时间
}

// DiscountProduct 折扣关联商品领域实体
type DiscountProduct struct {
	ID          uint      // 关联记录 ID
	DiscountID  uint      // 折扣 ID
	ProductID   uint      // 商品 ID
	ProductName string    // 商品名称（冗余展示）
	CreatedAt   time.Time // 创建时间
}

// DiscountCategory 折扣关联分类领域实体
type DiscountCategory struct {
	ID           uint      // 关联记录 ID
	DiscountID   uint      // 折扣 ID
	CategoryID   uint      // 分类 ID
	CategoryName string    // 分类名称（冗余展示）
	CreatedAt    time.Time // 创建时间
}

// DiscountManufacturer 折扣关联制造商领域实体
type DiscountManufacturer struct {
	ID               uint      // 关联记录 ID
	DiscountID       uint      // 折扣 ID
	ManufacturerID   uint      // 制造商 ID
	ManufacturerName string    // 制造商名称（冗余展示）
	CreatedAt        time.Time // 创建时间
}

// DiscountUsageHistory 折扣使用历史领域实体
type DiscountUsageHistory struct {
	ID           uint      // 使用记录 ID
	DiscountID   uint      // 折扣 ID
	OrderID      uint      // 订单 ID
	CustomerID   uint      // 客户 ID
	CustomerName string    // 客户名称（冗余展示）
	CouponCode   string    // 使用的优惠券码
	UsedOn       time.Time // 使用日期
	CreatedAt    time.Time // 创建时间
}

// ==================== 仓储接口定义 ====================

// DiscountRepository 折扣仓储接口
type DiscountRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*Discount, int64, error)
	Create(ctx context.Context, discount *Discount) (*Discount, error)
	Update(ctx context.Context, discount *Discount) (*Discount, error)
	Delete(ctx context.Context, id uint) error
}

// DiscountProductRepository 折扣关联商品仓储接口
type DiscountProductRepository interface {
	ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*DiscountProduct, int64, error)
}

// DiscountCategoryRepository 折扣关联分类仓储接口
type DiscountCategoryRepository interface {
	ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*DiscountCategory, int64, error)
}

// DiscountManufacturerRepository 折扣关联制造商仓储接口
type DiscountManufacturerRepository interface {
	ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*DiscountManufacturer, int64, error)
}

// DiscountUsageHistoryRepository 折扣使用历史仓储接口
type DiscountUsageHistoryRepository interface {
	ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*DiscountUsageHistory, int64, error)
}

// ==================== 用例实现 ====================

// DiscountUsecase 折扣服务业务用例
type DiscountUsecase struct {
	discountRepo      DiscountRepository
	productRepo       DiscountProductRepository
	categoryRepo      DiscountCategoryRepository
	manufacturerRepo  DiscountManufacturerRepository
	usageHistoryRepo  DiscountUsageHistoryRepository
}

// NewDiscountUsecase 创建折扣服务业务用例
func NewDiscountUsecase(
	discountRepo DiscountRepository,
	productRepo DiscountProductRepository,
	categoryRepo DiscountCategoryRepository,
	manufacturerRepo DiscountManufacturerRepository,
	usageHistoryRepo DiscountUsageHistoryRepository,
) *DiscountUsecase {
	return &DiscountUsecase{
		discountRepo:      discountRepo,
		productRepo:       productRepo,
		categoryRepo:      categoryRepo,
		manufacturerRepo:  manufacturerRepo,
		usageHistoryRepo:  usageHistoryRepo,
	}
}

// ==================== 折扣 CRUD 用例 ====================

// ListDiscounts 获取折扣列表
func (uc *DiscountUsecase) ListDiscounts(ctx context.Context, page, pageSize int) ([]*response.DiscountResponse, int64, error) {
	discounts, total, err := uc.discountRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.DiscountResponse, len(discounts))
	for i, d := range discounts {
		items[i] = toDiscountResponse(d)
	}
	return items, total, nil
}

// CreateDiscount 创建折扣
func (uc *DiscountUsecase) CreateDiscount(ctx context.Context, req request.CreateDiscountRequest) (*response.DiscountResponse, error) {
	discount := &Discount{
		Name:              req.Name,
		DiscountType:      req.DiscountType,
		DiscountAmount:    req.DiscountAmount,
		RequiresCouponCode: req.RequiresCouponCode,
		CouponCode:        req.CouponCode,
		IsCumulative:      req.IsCumulative,
		DisplayOrder:      req.DisplayOrder,
		IsActive:          req.IsActive,
		LimitationTimes:   req.LimitationTimes,
	}

	// 解析日期字段
	if req.StartDate != "" {
		discount.StartDate, _ = time.Parse("2006-01-02", req.StartDate)
	}
	if req.EndDate != "" {
		discount.EndDate, _ = time.Parse("2006-01-02", req.EndDate)
	}

	created, err := uc.discountRepo.Create(ctx, discount)
	if err != nil {
		return nil, err
	}
	return toDiscountResponse(created), nil
}

// UpdateDiscount 更新折扣
func (uc *DiscountUsecase) UpdateDiscount(ctx context.Context, req request.UpdateDiscountRequest) (*response.DiscountResponse, error) {
	discount := &Discount{
		ID:                req.ID,
		Name:              req.Name,
		DiscountType:      req.DiscountType,
		DiscountAmount:    req.DiscountAmount,
		RequiresCouponCode: req.RequiresCouponCode,
		CouponCode:        req.CouponCode,
		IsCumulative:      req.IsCumulative,
		DisplayOrder:      req.DisplayOrder,
		IsActive:          req.IsActive,
		LimitationTimes:   req.LimitationTimes,
	}

	// 解析日期字段
	if req.StartDate != "" {
		discount.StartDate, _ = time.Parse("2006-01-02", req.StartDate)
	}
	if req.EndDate != "" {
		discount.EndDate, _ = time.Parse("2006-01-02", req.EndDate)
	}

	updated, err := uc.discountRepo.Update(ctx, discount)
	if err != nil {
		return nil, err
	}
	return toDiscountResponse(updated), nil
}

// DeleteDiscount 删除折扣
func (uc *DiscountUsecase) DeleteDiscount(ctx context.Context, id uint) error {
	return uc.discountRepo.Delete(ctx, id)
}

// ==================== 折扣关联子资源用例 ====================

// ListDiscountProducts 获取折扣关联商品列表
func (uc *DiscountUsecase) ListDiscountProducts(ctx context.Context, discountID uint, page, pageSize int) ([]*response.DiscountProductResponse, int64, error) {
	products, total, err := uc.productRepo.ListByDiscountID(ctx, discountID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.DiscountProductResponse, len(products))
	for i, p := range products {
		items[i] = toDiscountProductResponse(p)
	}
	return items, total, nil
}

// ListDiscountCategories 获取折扣关联分类列表
func (uc *DiscountUsecase) ListDiscountCategories(ctx context.Context, discountID uint, page, pageSize int) ([]*response.DiscountCategoryResponse, int64, error) {
	categories, total, err := uc.categoryRepo.ListByDiscountID(ctx, discountID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.DiscountCategoryResponse, len(categories))
	for i, c := range categories {
		items[i] = toDiscountCategoryResponse(c)
	}
	return items, total, nil
}

// ListDiscountManufacturers 获取折扣关联制造商列表
func (uc *DiscountUsecase) ListDiscountManufacturers(ctx context.Context, discountID uint, page, pageSize int) ([]*response.DiscountManufacturerResponse, int64, error) {
	manufacturers, total, err := uc.manufacturerRepo.ListByDiscountID(ctx, discountID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.DiscountManufacturerResponse, len(manufacturers))
	for i, m := range manufacturers {
		items[i] = toDiscountManufacturerResponse(m)
	}
	return items, total, nil
}

// ListDiscountUsageHistory 获取折扣使用历史列表
func (uc *DiscountUsecase) ListDiscountUsageHistory(ctx context.Context, discountID uint, page, pageSize int) ([]*response.DiscountUsageHistoryResponse, int64, error) {
	history, total, err := uc.usageHistoryRepo.ListByDiscountID(ctx, discountID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.DiscountUsageHistoryResponse, len(history))
	for i, h := range history {
		items[i] = toDiscountUsageHistoryResponse(h)
	}
	return items, total, nil
}

// ==================== 内部转换函数 ====================

// toDiscountResponse 领域实体转换为折扣响应 DTO
func toDiscountResponse(d *Discount) *response.DiscountResponse {
	resp := &response.DiscountResponse{
		ID:                d.ID,
		Name:              d.Name,
		DiscountType:      d.DiscountType,
		DiscountAmount:    d.DiscountAmount,
		RequiresCouponCode: d.RequiresCouponCode,
		CouponCode:        d.CouponCode,
		IsCumulative:      d.IsCumulative,
		DisplayOrder:      d.DisplayOrder,
		IsActive:          d.IsActive,
		LimitationTimes:   d.LimitationTimes,
		CreatedAt:         d.CreatedAt.Unix(),
		UpdatedAt:         d.UpdatedAt.Unix(),
	}
	// 日期格式化为 YYYY-MM-DD
	if !d.StartDate.IsZero() {
		resp.StartDate = d.StartDate.Format("2006-01-02")
	}
	if !d.EndDate.IsZero() {
		resp.EndDate = d.EndDate.Format("2006-01-02")
	}
	return resp
}

// toDiscountProductResponse 领域实体转换为折扣关联商品响应 DTO
func toDiscountProductResponse(p *DiscountProduct) *response.DiscountProductResponse {
	return &response.DiscountProductResponse{
		ID:          p.ID,
		DiscountID:  p.DiscountID,
		ProductID:   p.ProductID,
		ProductName: p.ProductName,
		CreatedAt:   p.CreatedAt.Unix(),
	}
}

// toDiscountCategoryResponse 领域实体转换为折扣关联分类响应 DTO
func toDiscountCategoryResponse(c *DiscountCategory) *response.DiscountCategoryResponse {
	return &response.DiscountCategoryResponse{
		ID:           c.ID,
		DiscountID:   c.DiscountID,
		CategoryID:   c.CategoryID,
		CategoryName: c.CategoryName,
		CreatedAt:    c.CreatedAt.Unix(),
	}
}

// toDiscountManufacturerResponse 领域实体转换为折扣关联制造商响应 DTO
func toDiscountManufacturerResponse(m *DiscountManufacturer) *response.DiscountManufacturerResponse {
	return &response.DiscountManufacturerResponse{
		ID:               m.ID,
		DiscountID:       m.DiscountID,
		ManufacturerID:   m.ManufacturerID,
		ManufacturerName: m.ManufacturerName,
		CreatedAt:        m.CreatedAt.Unix(),
	}
}

// toDiscountUsageHistoryResponse 领域实体转换为折扣使用历史响应 DTO
func toDiscountUsageHistoryResponse(h *DiscountUsageHistory) *response.DiscountUsageHistoryResponse {
	resp := &response.DiscountUsageHistoryResponse{
		ID:           h.ID,
		DiscountID:   h.DiscountID,
		OrderID:      h.OrderID,
		CustomerID:   h.CustomerID,
		CustomerName: h.CustomerName,
		CouponCode:   h.CouponCode,
		CreatedAt:    h.CreatedAt.Unix(),
	}
	if !h.UsedOn.IsZero() {
		resp.UsedOn = h.UsedOn.Format("2006-01-02")
	}
	return resp
}