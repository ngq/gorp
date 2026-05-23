// Package biz 业务逻辑层。
// 定义目录服务（directory）的领域实体、仓储接口和用例。
// 包含国家、省/州、货币等地理/货币信息管理。
package biz

import (
	"context"
	"time"
)

// ==================== 国家 ====================

// Country 国家领域实体。
type Country struct {
	ID               uint      // 国家ID
	Name             string    // 国家名称
	IsoCode2         string    // ISO 2字母代码
	IsoCode3         string    // ISO 3字母代码
	AddressFormat    string    // 地址格式
	PostcodeRequired bool      // 是否需要邮编
	CreatedAt        time.Time // 创建时间
	UpdatedAt        time.Time // 更新时间
}

// CountryRepository 国家仓储接口。
type CountryRepository interface {
	Create(ctx context.Context, country *Country) error
	GetByID(ctx context.Context, id uint) (*Country, error)
	List(ctx context.Context, page, size int) ([]*Country, int64, error)
	Update(ctx context.Context, country *Country) error
	Delete(ctx context.Context, id uint) error
}

// CountryUseCase 国家用例。
type CountryUseCase struct {
	repo CountryRepository
}

// NewCountryUseCase 创建国家用例。
func NewCountryUseCase(repo CountryRepository) *CountryUseCase {
	return &CountryUseCase{repo: repo}
}

// Create 创建国家。
func (uc *CountryUseCase) Create(ctx context.Context, name, isoCode2, isoCode3, addressFormat string, postcodeRequired bool) (*Country, error) {
	country := &Country{
		Name:             name,
		IsoCode2:         isoCode2,
		IsoCode3:         isoCode3,
		AddressFormat:    addressFormat,
		PostcodeRequired: postcodeRequired,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := uc.repo.Create(ctx, country); err != nil {
		return nil, err
	}
	return country, nil
}

// GetByID 根据ID获取国家。
func (uc *CountryUseCase) GetByID(ctx context.Context, id uint) (*Country, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取国家列表。
func (uc *CountryUseCase) List(ctx context.Context, page, size int) ([]*Country, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新国家。
func (uc *CountryUseCase) Update(ctx context.Context, id uint, name, isoCode2, isoCode3, addressFormat string, postcodeRequired bool) (*Country, error) {
	country, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	country.Name = name
	country.IsoCode2 = isoCode2
	country.IsoCode3 = isoCode3
	country.AddressFormat = addressFormat
	country.PostcodeRequired = postcodeRequired
	country.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, country); err != nil {
		return nil, err
	}
	return country, nil
}

// Delete 删除国家。
func (uc *CountryUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// ==================== 省/州 ====================

// State 省/州领域实体。
type State struct {
	ID        uint      // 省/州ID
	CountryID uint      // 所属国家ID
	Name      string    // 省/州名称
	IsoCode   string    // ISO 代码
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// StateRepository 省/州仓储接口。
type StateRepository interface {
	Create(ctx context.Context, state *State) error
	GetByID(ctx context.Context, id uint) (*State, error)
	ListByCountryID(ctx context.Context, countryID uint, page, size int) ([]*State, int64, error)
	Update(ctx context.Context, state *State) error
	Delete(ctx context.Context, id uint) error
}

// StateUseCase 省/州用例。
type StateUseCase struct {
	repo StateRepository
}

// NewStateUseCase 创建省/州用例。
func NewStateUseCase(repo StateRepository) *StateUseCase {
	return &StateUseCase{repo: repo}
}

// Create 创建省/州。
func (uc *StateUseCase) Create(ctx context.Context, countryID uint, name, isoCode string) (*State, error) {
	state := &State{
		CountryID: countryID,
		Name:      name,
		IsoCode:   isoCode,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, state); err != nil {
		return nil, err
	}
	return state, nil
}

// GetByID 根据ID获取省/州。
func (uc *StateUseCase) GetByID(ctx context.Context, id uint) (*State, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListByCountryID 获取指定国家下的省/州列表。
func (uc *StateUseCase) ListByCountryID(ctx context.Context, countryID uint, page, size int) ([]*State, int64, error) {
	return uc.repo.ListByCountryID(ctx, countryID, page, size)
}

// Update 更新省/州。
func (uc *StateUseCase) Update(ctx context.Context, id uint, name, isoCode string) (*State, error) {
	state, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	state.Name = name
	state.IsoCode = isoCode
	state.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, state); err != nil {
		return nil, err
	}
	return state, nil
}

// Delete 删除省/州。
func (uc *StateUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// ==================== 货币 ====================

// Currency 货币领域实体。
type Currency struct {
	ID        uint      // 货币ID
	Name      string    // 货币名称
	Code      string    // 货币代码（如 CNY、USD）
	Symbol    string    // 货币符号（如 ¥、$）
	Rate      float64   // 汇率（相对于基础货币）
	IsActive  bool      // 是否启用
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// CurrencyRepository 货币仓储接口。
type CurrencyRepository interface {
	Create(ctx context.Context, currency *Currency) error
	GetByID(ctx context.Context, id uint) (*Currency, error)
	List(ctx context.Context, page, size int) ([]*Currency, int64, error)
	Update(ctx context.Context, currency *Currency) error
	Delete(ctx context.Context, id uint) error
	BatchUpdateRates(ctx context.Context, rates []CurrencyRateItem) error
}

// CurrencyRateItem 汇率更新项。
type CurrencyRateItem struct {
	CurrencyID uint    // 货币ID
	Rate       float64 // 汇率
}

// CurrencyUseCase 货币用例。
type CurrencyUseCase struct {
	repo CurrencyRepository
}

// NewCurrencyUseCase 创建货币用例。
func NewCurrencyUseCase(repo CurrencyRepository) *CurrencyUseCase {
	return &CurrencyUseCase{repo: repo}
}

// Create 创建货币。
func (uc *CurrencyUseCase) Create(ctx context.Context, name, code, symbol string, rate float64, isActive bool) (*Currency, error) {
	currency := &Currency{
		Name:      name,
		Code:      code,
		Symbol:    symbol,
		Rate:      rate,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, currency); err != nil {
		return nil, err
	}
	return currency, nil
}

// GetByID 根据ID获取货币。
func (uc *CurrencyUseCase) GetByID(ctx context.Context, id uint) (*Currency, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取货币列表。
func (uc *CurrencyUseCase) List(ctx context.Context, page, size int) ([]*Currency, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新货币。
func (uc *CurrencyUseCase) Update(ctx context.Context, id uint, name, code, symbol string, rate float64, isActive bool) (*Currency, error) {
	currency, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	currency.Name = name
	currency.Code = code
	currency.Symbol = symbol
	currency.Rate = rate
	currency.IsActive = isActive
	currency.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, currency); err != nil {
		return nil, err
	}
	return currency, nil
}

// Delete 删除货币。
func (uc *CurrencyUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// ApplyRates 批量应用汇率更新。
func (uc *CurrencyUseCase) ApplyRates(ctx context.Context, rates []CurrencyRateItem) error {
	return uc.repo.BatchUpdateRates(ctx, rates)
}
