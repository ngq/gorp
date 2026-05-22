// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/directory/internal/biz"

	"gorm.io/gorm"
)

// ==================== 国家 ====================

// CountryPO 国家持久化对象。
type CountryPO struct {
	ID               uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name             string    `gorm:"size:128;column:name" db:"name" json:"name"`
	IsoCode2         string    `gorm:"size:2;column:iso_code2" db:"iso_code2" json:"iso_code2"`
	IsoCode3         string    `gorm:"size:3;column:iso_code3" db:"iso_code3" json:"iso_code3"`
	AddressFormat    string    `gorm:"size:256;column:address_format" db:"address_format" json:"address_format"`
	PostcodeRequired bool      `gorm:"column:postcode_required" db:"postcode_required" json:"postcode_required"`
	CreatedAt        time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (CountryPO) TableName() string { return "countries" }

// ToEntity 转换为领域实体。
func (po *CountryPO) ToEntity() *biz.Country {
	return &biz.Country{
		ID:               po.ID,
		Name:             po.Name,
		IsoCode2:         po.IsoCode2,
		IsoCode3:         po.IsoCode3,
		AddressFormat:    po.AddressFormat,
		PostcodeRequired: po.PostcodeRequired,
		CreatedAt:        po.CreatedAt,
		UpdatedAt:        po.UpdatedAt,
	}
}

// countryRepo 国家仓储实现。
type countryRepo struct{ db *gorm.DB }

// NewCountryRepo 创建国家仓储。
func NewCountryRepo(db *gorm.DB) biz.CountryRepository { return &countryRepo{db: db} }

// Create 创建国家。
func (r *countryRepo) Create(ctx context.Context, country *biz.Country) error {
	po := &CountryPO{
		Name: country.Name, IsoCode2: country.IsoCode2, IsoCode3: country.IsoCode3,
		AddressFormat: country.AddressFormat, PostcodeRequired: country.PostcodeRequired,
		CreatedAt: country.CreatedAt, UpdatedAt: country.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取国家。
func (r *countryRepo) GetByID(ctx context.Context, id uint) (*biz.Country, error) {
	var po CountryPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取国家列表。
func (r *countryRepo) List(ctx context.Context, page, size int) ([]*biz.Country, int64, error) {
	var pos []CountryPO
	var total int64
	r.db.WithContext(ctx).Model(&CountryPO{}).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Country, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

// Update 更新国家。
func (r *countryRepo) Update(ctx context.Context, country *biz.Country) error {
	return r.db.WithContext(ctx).Model(&CountryPO{}).Where("id = ?", country.ID).Updates(map[string]interface{}{
		"name": country.Name, "iso_code2": country.IsoCode2, "iso_code3": country.IsoCode3,
		"address_format": country.AddressFormat, "postcode_required": country.PostcodeRequired,
		"updated_at": country.UpdatedAt,
	}).Error
}

// Delete 删除国家。
func (r *countryRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&CountryPO{}, id).Error
}

// ==================== 省/州 ====================

// StatePO 省/州持久化对象。
type StatePO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	CountryID uint      `gorm:"index;column:country_id" db:"country_id" json:"country_id"`
	Name      string    `gorm:"size:128;column:name" db:"name" json:"name"`
	IsoCode   string    `gorm:"size:10;column:iso_code" db:"iso_code" json:"iso_code"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (StatePO) TableName() string { return "states" }

// ToEntity 转换为领域实体。
func (po *StatePO) ToEntity() *biz.State {
	return &biz.State{
		ID: po.ID, CountryID: po.CountryID, Name: po.Name, IsoCode: po.IsoCode,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// stateRepo 省/州仓储实现。
type stateRepo struct{ db *gorm.DB }

// NewStateRepo 创建省/州仓储。
func NewStateRepo(db *gorm.DB) biz.StateRepository { return &stateRepo{db: db} }

// Create 创建省/州。
func (r *stateRepo) Create(ctx context.Context, state *biz.State) error {
	po := &StatePO{
		CountryID: state.CountryID, Name: state.Name, IsoCode: state.IsoCode,
		CreatedAt: state.CreatedAt, UpdatedAt: state.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取省/州。
func (r *stateRepo) GetByID(ctx context.Context, id uint) (*biz.State, error) {
	var po StatePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// ListByCountryID 获取指定国家下的省/州列表。
func (r *stateRepo) ListByCountryID(ctx context.Context, countryID uint, page, size int) ([]*biz.State, int64, error) {
	var pos []StatePO
	var total int64
	r.db.WithContext(ctx).Model(&StatePO{}).Where("country_id = ?", countryID).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Where("country_id = ?", countryID).Order("id DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.State, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

// Update 更新省/州。
func (r *stateRepo) Update(ctx context.Context, state *biz.State) error {
	return r.db.WithContext(ctx).Model(&StatePO{}).Where("id = ?", state.ID).Updates(map[string]interface{}{
		"name": state.Name, "iso_code": state.IsoCode, "updated_at": state.UpdatedAt,
	}).Error
}

// Delete 删除省/州。
func (r *stateRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&StatePO{}, id).Error
}

// ==================== 货币 ====================

// CurrencyPO 货币持久化对象。
type CurrencyPO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name      string    `gorm:"size:64;column:name" db:"name" json:"name"`
	Code      string    `gorm:"size:3;column:code" db:"code" json:"code"`
	Symbol    string    `gorm:"size:10;column:symbol" db:"symbol" json:"symbol"`
	Rate      float64   `gorm:"column:rate" db:"rate" json:"rate"`
	IsActive  bool      `gorm:"column:is_active" db:"is_active" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (CurrencyPO) TableName() string { return "currencies" }

// ToEntity 转换为领域实体。
func (po *CurrencyPO) ToEntity() *biz.Currency {
	return &biz.Currency{
		ID: po.ID, Name: po.Name, Code: po.Code, Symbol: po.Symbol,
		Rate: po.Rate, IsActive: po.IsActive, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// currencyRepo 货币仓储实现。
type currencyRepo struct{ db *gorm.DB }

// NewCurrencyRepo 创建货币仓储。
func NewCurrencyRepo(db *gorm.DB) biz.CurrencyRepository { return &currencyRepo{db: db} }

// Create 创建货币。
func (r *currencyRepo) Create(ctx context.Context, currency *biz.Currency) error {
	po := &CurrencyPO{
		Name: currency.Name, Code: currency.Code, Symbol: currency.Symbol,
		Rate: currency.Rate, IsActive: currency.IsActive,
		CreatedAt: currency.CreatedAt, UpdatedAt: currency.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取货币。
func (r *currencyRepo) GetByID(ctx context.Context, id uint) (*biz.Currency, error) {
	var po CurrencyPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取货币列表。
func (r *currencyRepo) List(ctx context.Context, page, size int) ([]*biz.Currency, int64, error) {
	var pos []CurrencyPO
	var total int64
	r.db.WithContext(ctx).Model(&CurrencyPO{}).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Currency, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

// Update 更新货币。
func (r *currencyRepo) Update(ctx context.Context, currency *biz.Currency) error {
	return r.db.WithContext(ctx).Model(&CurrencyPO{}).Where("id = ?", currency.ID).Updates(map[string]interface{}{
		"name": currency.Name, "code": currency.Code, "symbol": currency.Symbol,
		"rate": currency.Rate, "is_active": currency.IsActive, "updated_at": currency.UpdatedAt,
	}).Error
}

// Delete 删除货币。
func (r *currencyRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&CurrencyPO{}, id).Error
}

// BatchUpdateRates 批量更新汇率。
//
// 使用事务确保所有汇率更新原子性完成。
func (r *currencyRepo) BatchUpdateRates(ctx context.Context, rates []biz.CurrencyRateItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range rates {
			if err := tx.Model(&CurrencyPO{}).Where("id = ?", item.CurrencyID).
				Updates(map[string]interface{}{"rate": item.Rate, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
