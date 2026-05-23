// Package biz 包含交易服务的业务逻辑层
// tax.go 定义税务领域实体、仓库接口与税务用例
// 注意：原 tax 服务中的 Provider 实体已重命名为 TaxProvider，Category 已重命名为 TaxCategory
package biz

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// 税务领域实体
// ============================================================================

// TaxProvider 税务服务商实体（原 Provider，重命名以避免与 ShippingProvider 冲突）
type TaxProvider struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"` // 税务服务商编码
	Description string         `json:"description"`
	IsActive    bool           `json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TaxCategory 税种分类实体（原 Category，重命名以避免与业务分类混淆）
type TaxCategory struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"` // 税种编码，如 vat、sales_tax、gst
	Description string         `json:"description"`
	Rate        float64        `json:"rate"` // 默认税率
	IsActive    bool           `json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TaxRate 税率实体，按区域和税种组合定价
type TaxRate struct {
	ID           string         `json:"id"`
	TaxCategoryID string        `json:"taxCategoryId"`
	Region       string         `json:"region"` // 区域编码，如 CN、US-CA
	Rate         float64        `json:"rate"`
	IsActive     bool           `json:"isActive"`
	EffectiveFrom time.Time     `json:"effectiveFrom"`
	EffectiveTo   *time.Time    `json:"effectiveTo"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TaxTransaction 税务交易记录实体
type TaxTransaction struct {
	ID           string         `json:"id"`
	OrderID      string         `json:"orderId"`
	TaxCategoryID string        `json:"taxCategoryId"`
	TaxRateID    string         `json:"taxRateId"`
	TaxableAmount float64       `json:"taxableAmount"`
	TaxAmount    float64        `json:"taxAmount"`
	Currency     string         `json:"currency"`
	CreatedAt    time.Time      `json:"createdAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// ============================================================================
// 仓库接口定义
// ============================================================================

// TaxProviderRepo 税务服务商数据仓库接口
type TaxProviderRepo interface {
	// Create 创建税务服务商
	Create(ctx context.Context, provider *TaxProvider) error
	// GetByID 根据ID获取税务服务商
	GetByID(ctx context.Context, id string) (*TaxProvider, error)
	// Update 更新税务服务商
	Update(ctx context.Context, provider *TaxProvider) error
	// Delete 删除税务服务商
	Delete(ctx context.Context, id string) error
	// List 获取税务服务商列表
	List(ctx context.Context) ([]*TaxProvider, error)
	// GetByCode 根据编码获取税务服务商
	GetByCode(ctx context.Context, code string) (*TaxProvider, error)
}

// TaxCategoryRepo 税种分类数据仓库接口
type TaxCategoryRepo interface {
	// Create 创建税种分类
	Create(ctx context.Context, category *TaxCategory) error
	// GetByID 根据ID获取税种分类
	GetByID(ctx context.Context, id string) (*TaxCategory, error)
	// Update 更新税种分类
	Update(ctx context.Context, category *TaxCategory) error
	// Delete 删除税种分类
	Delete(ctx context.Context, id string) error
	// List 获取税种分类列表
	List(ctx context.Context) ([]*TaxCategory, error)
	// GetByCode 根据编码获取税种分类
	GetByCode(ctx context.Context, code string) (*TaxCategory, error)
}

// TaxRateRepo 税率数据仓库接口
type TaxRateRepo interface {
	// Create 创建税率
	Create(ctx context.Context, rate *TaxRate) error
	// GetByID 根据ID获取税率
	GetByID(ctx context.Context, id string) (*TaxRate, error)
	// Update 更新税率
	Update(ctx context.Context, rate *TaxRate) error
	// Delete 删除税率
	Delete(ctx context.Context, id string) error
	// ListByCategory 按税种分类查询税率列表
	ListByCategory(ctx context.Context, categoryID string) ([]*TaxRate, error)
	// FindEffectiveRate 查询指定区域和税种的有效税率
	FindEffectiveRate(ctx context.Context, categoryID, region string) (*TaxRate, error)
}

// TaxTransactionRepo 税务交易记录数据仓库接口
type TaxTransactionRepo interface {
	// Create 创建税务交易记录
	Create(ctx context.Context, txn *TaxTransaction) error
	// GetByID 根据ID获取税务交易记录
	GetByID(ctx context.Context, id string) (*TaxTransaction, error)
	// ListByOrderID 按订单ID查询税务交易记录
	ListByOrderID(ctx context.Context, orderID string) ([]*TaxTransaction, error)
	// ListByCategory 按税种分类查询税务交易记录
	ListByCategory(ctx context.Context, categoryID string, page, pageSize int) ([]*TaxTransaction, int64, error)
}

// ============================================================================
// 用例（UseCase）
// ============================================================================

// TaxUseCase 税务业务用例，封装税务服务商、税种、税率、税务交易的管理逻辑
// 重构说明：UseCase 方法返回领域实体而非 response DTO，DTO 转换移至 service 层
type TaxUseCase struct {
	providerRepo    TaxProviderRepo
	categoryRepo    TaxCategoryRepo
	rateRepo        TaxRateRepo
	transactionRepo TaxTransactionRepo
}

// NewTaxUseCase 创建税务用例实例
func NewTaxUseCase(
	providerRepo TaxProviderRepo,
	categoryRepo TaxCategoryRepo,
	rateRepo TaxRateRepo,
	transactionRepo TaxTransactionRepo,
) *TaxUseCase {
	return &TaxUseCase{
		providerRepo:    providerRepo,
		categoryRepo:    categoryRepo,
		rateRepo:        rateRepo,
		transactionRepo: transactionRepo,
	}
}

// --- 税务服务商操作 ---

// CreateTaxProvider 创建税务服务商
func (uc *TaxUseCase) CreateTaxProvider(ctx context.Context, provider *TaxProvider) error {
	return uc.providerRepo.Create(ctx, provider)
}

// GetTaxProvider 获取税务服务商详情
func (uc *TaxUseCase) GetTaxProvider(ctx context.Context, id string) (*TaxProvider, error) {
	provider, err := uc.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取税务服务商失败: %w", err)
	}
	return provider, nil
}

// UpdateTaxProvider 更新税务服务商
func (uc *TaxUseCase) UpdateTaxProvider(ctx context.Context, provider *TaxProvider) error {
	return uc.providerRepo.Update(ctx, provider)
}

// DeleteTaxProvider 删除税务服务商
func (uc *TaxUseCase) DeleteTaxProvider(ctx context.Context, id string) error {
	return uc.providerRepo.Delete(ctx, id)
}

// ListTaxProviders 获取税务服务商列表
func (uc *TaxUseCase) ListTaxProviders(ctx context.Context) ([]*TaxProvider, error) {
	return uc.providerRepo.List(ctx)
}

// --- 税种分类操作 ---

// CreateTaxCategory 创建税种分类
func (uc *TaxUseCase) CreateTaxCategory(ctx context.Context, category *TaxCategory) error {
	return uc.categoryRepo.Create(ctx, category)
}

// GetTaxCategory 获取税种分类详情
func (uc *TaxUseCase) GetTaxCategory(ctx context.Context, id string) (*TaxCategory, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取税种分类失败: %w", err)
	}
	return category, nil
}

// UpdateTaxCategory 更新税种分类
func (uc *TaxUseCase) UpdateTaxCategory(ctx context.Context, category *TaxCategory) error {
	return uc.categoryRepo.Update(ctx, category)
}

// DeleteTaxCategory 删除税种分类
func (uc *TaxUseCase) DeleteTaxCategory(ctx context.Context, id string) error {
	return uc.categoryRepo.Delete(ctx, id)
}

// ListTaxCategories 获取税种分类列表
func (uc *TaxUseCase) ListTaxCategories(ctx context.Context) ([]*TaxCategory, error) {
	return uc.categoryRepo.List(ctx)
}

// --- 税率操作 ---

// CreateTaxRate 创建税率
func (uc *TaxUseCase) CreateTaxRate(ctx context.Context, rate *TaxRate) error {
	return uc.rateRepo.Create(ctx, rate)
}

// GetTaxRate 获取税率详情
func (uc *TaxUseCase) GetTaxRate(ctx context.Context, id string) (*TaxRate, error) {
	return uc.rateRepo.GetByID(ctx, id)
}

// UpdateTaxRate 更新税率
func (uc *TaxUseCase) UpdateTaxRate(ctx context.Context, rate *TaxRate) error {
	return uc.rateRepo.Update(ctx, rate)
}

// DeleteTaxRate 删除税率
func (uc *TaxUseCase) DeleteTaxRate(ctx context.Context, id string) error {
	return uc.rateRepo.Delete(ctx, id)
}

// ListTaxRates 按税种分类查询税率列表
func (uc *TaxUseCase) ListTaxRates(ctx context.Context, categoryID string) ([]*TaxRate, error) {
	return uc.rateRepo.ListByCategory(ctx, categoryID)
}

// FindEffectiveTaxRate 查询指定区域和税种的有效税率
func (uc *TaxUseCase) FindEffectiveTaxRate(ctx context.Context, categoryID, region string) (*TaxRate, error) {
	return uc.rateRepo.FindEffectiveRate(ctx, categoryID, region)
}

// --- 税务交易记录操作 ---

// CreateTaxTransaction 创建税务交易记录
func (uc *TaxUseCase) CreateTaxTransaction(ctx context.Context, txn *TaxTransaction) error {
	return uc.transactionRepo.Create(ctx, txn)
}

// GetTaxTransaction 获取税务交易记录详情
func (uc *TaxUseCase) GetTaxTransaction(ctx context.Context, id string) (*TaxTransaction, error) {
	return uc.transactionRepo.GetByID(ctx, id)
}

// ListTaxTransactionsByOrder 按订单ID查询税务交易记录
func (uc *TaxUseCase) ListTaxTransactionsByOrder(ctx context.Context, orderID string) ([]*TaxTransaction, error) {
	return uc.transactionRepo.ListByOrderID(ctx, orderID)
}

// ListTaxTransactionsByCategory 按税种分类分页查询税务交易记录
func (uc *TaxUseCase) ListTaxTransactionsByCategory(ctx context.Context, categoryID string, page, pageSize int) ([]*TaxTransaction, int64, error) {
	return uc.transactionRepo.ListByCategory(ctx, categoryID, page, pageSize)
}
