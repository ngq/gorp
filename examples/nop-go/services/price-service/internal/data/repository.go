// Package data 价格服务数据访问层
package data

import (
	"context"

	"nop-go/services/price-service/internal/models"

	"gorm.io/gorm"
)

type TaxRateRepository interface {
	Create(ctx context.Context, rate *models.TaxRate) error
	GetByID(ctx context.Context, id uint64) (*models.TaxRate, error)
	GetByLocation(ctx context.Context, countryCode, stateCode, zipCode string) (*models.TaxRate, error)
	Update(ctx context.Context, rate *models.TaxRate) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.TaxRate, error)
}

type DiscountRepository interface {
	Create(ctx context.Context, discount *models.Discount) error
	GetByID(ctx context.Context, id uint64) (*models.Discount, error)
	GetByCouponCode(ctx context.Context, code string) (*models.Discount, error)
	Update(ctx context.Context, discount *models.Discount) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.Discount, error)
	IncrementUsage(ctx context.Context, id uint64) error
}

type ProductPriceRepository interface {
	Create(ctx context.Context, price *models.ProductPrice) error
	GetByProductID(ctx context.Context, productID uint64) (*models.ProductPrice, error)
	Update(ctx context.Context, price *models.ProductPrice) error
}

type taxRateRepo struct{ db *gorm.DB }

func NewTaxRateRepository(db *gorm.DB) TaxRateRepository {
	return &taxRateRepo{db: db}
}

func (r *taxRateRepo) Create(ctx context.Context, t *models.TaxRate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *taxRateRepo) GetByID(ctx context.Context, id uint64) (*models.TaxRate, error) {
	var t models.TaxRate
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *taxRateRepo) GetByLocation(ctx context.Context, countryCode, stateCode, zipCode string) (*models.TaxRate, error) {
	var t models.TaxRate
	query := r.db.WithContext(ctx).Where("country_code = ?", countryCode)
	if stateCode != "" {
		query = query.Where("state_code = ? OR state_code = ''", stateCode)
	}
	if zipCode != "" {
		query = query.Where("zip_code = ? OR zip_code = ''", zipCode)
	}
	err := query.Order("priority DESC").First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *taxRateRepo) Update(ctx context.Context, t *models.TaxRate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *taxRateRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.TaxRate{}, id).Error
}

func (r *taxRateRepo) List(ctx context.Context) ([]*models.TaxRate, error) {
	var list []*models.TaxRate
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

type discountRepo struct{ db *gorm.DB }

func NewDiscountRepository(db *gorm.DB) DiscountRepository {
	return &discountRepo{db: db}
}

func (r *discountRepo) Create(ctx context.Context, d *models.Discount) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *discountRepo) GetByID(ctx context.Context, id uint64) (*models.Discount, error) {
	var d models.Discount
	err := r.db.WithContext(ctx).First(&d, id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *discountRepo) GetByCouponCode(ctx context.Context, code string) (*models.Discount, error) {
	var d models.Discount
	err := r.db.WithContext(ctx).Where("coupon_code = ? AND is_active = ?", code, true).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *discountRepo) Update(ctx context.Context, d *models.Discount) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *discountRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Discount{}, id).Error
}

func (r *discountRepo) List(ctx context.Context) ([]*models.Discount, error) {
	var list []*models.Discount
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&list).Error
	return list, err
}

func (r *discountRepo) IncrementUsage(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&models.Discount{}).
		Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

type productPriceRepo struct{ db *gorm.DB }

func NewProductPriceRepository(db *gorm.DB) ProductPriceRepository {
	return &productPriceRepo{db: db}
}

func (r *productPriceRepo) Create(ctx context.Context, p *models.ProductPrice) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *productPriceRepo) GetByProductID(ctx context.Context, productID uint64) (*models.ProductPrice, error) {
	var p models.ProductPrice
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productPriceRepo) Update(ctx context.Context, p *models.ProductPrice) error {
	return r.db.WithContext(ctx).Save(p).Error
}