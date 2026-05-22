package service

import (
	"context"

	"nop-go/services/directory/internal/biz"
	"nop-go/services/directory/internal/data"

	"gorm.io/gorm"
)

// Services 目录服务集合。
type Services struct {
	Directory *DirectoryService
}

// NewServices 创建目录服务集合。
func NewServices(db *gorm.DB) *Services {
	countryRepo := data.NewCountryRepo(db)
	stateRepo := data.NewStateRepo(db)
	currencyRepo := data.NewCurrencyRepo(db)

	countryUC := biz.NewCountryUseCase(countryRepo)
	stateUC := biz.NewStateUseCase(stateRepo)
	currencyUC := biz.NewCurrencyUseCase(currencyRepo)

	return &Services{
		Directory: &DirectoryService{
			countryUC:  countryUC,
			stateUC:    stateUC,
			currencyUC: currencyUC,
		},
	}
}

// DirectoryService 目录服务，聚合国家/省/州/货币。
type DirectoryService struct {
	countryUC  *biz.CountryUseCase
	stateUC    *biz.StateUseCase
	currencyUC *biz.CurrencyUseCase
}

// ==================== 国家请求/响应 ====================

// CreateCountryRequest 创建国家请求。
type CreateCountryRequest struct {
	Name             string `json:"name" binding:"required"`              // 国家名称
	IsoCode2         string `json:"iso_code2" binding:"required"`         // ISO 2字母代码
	IsoCode3         string `json:"iso_code3"`                            // ISO 3字母代码
	AddressFormat    string `json:"address_format"`                       // 地址格式
	PostcodeRequired bool   `json:"postcode_required"`                    // 是否需要邮编
}

// UpdateCountryRequest 更新国家请求。
type UpdateCountryRequest struct {
	Name             string `json:"name" binding:"required"`              // 国家名称
	IsoCode2         string `json:"iso_code2" binding:"required"`         // ISO 2字母代码
	IsoCode3         string `json:"iso_code3"`                            // ISO 3字母代码
	AddressFormat    string `json:"address_format"`                       // 地址格式
	PostcodeRequired bool   `json:"postcode_required"`                    // 是否需要邮编
}

// CountryResponse 国家响应。
type CountryResponse struct {
	ID               uint   `json:"id"`                // 国家ID
	Name             string `json:"name"`              // 国家名称
	IsoCode2         string `json:"iso_code2"`         // ISO 2字母代码
	IsoCode3         string `json:"iso_code3"`         // ISO 3字母代码
	AddressFormat    string `json:"address_format"`    // 地址格式
	PostcodeRequired bool   `json:"postcode_required"` // 是否需要邮编
	CreatedAt        string `json:"created_at"`        // 创建时间
	UpdatedAt        string `json:"updated_at"`        // 更新时间
}

// ==================== 省/州请求/响应 ====================

// CreateStateRequest 创建省/州请求。
type CreateStateRequest struct {
	CountryID uint   `json:"country_id" binding:"required"` // 所属国家ID
	Name      string `json:"name" binding:"required"`       // 省/州名称
	IsoCode   string `json:"iso_code"`                      // ISO 代码
}

// UpdateStateRequest 更新省/州请求。
type UpdateStateRequest struct {
	Name    string `json:"name" binding:"required"`    // 省/州名称
	IsoCode string `json:"iso_code"`                   // ISO 代码
}

// StateResponse 省/州响应。
type StateResponse struct {
	ID        uint   `json:"id"`         // 省/州ID
	CountryID uint   `json:"country_id"` // 所属国家ID
	Name      string `json:"name"`       // 省/州名称
	IsoCode   string `json:"iso_code"`   // ISO 代码
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// ==================== 货币请求/响应 ====================

// CreateCurrencyRequest 创建货币请求。
type CreateCurrencyRequest struct {
	Name     string  `json:"name" binding:"required"`   // 货币名称
	Code     string  `json:"code" binding:"required"`   // 货币代码
	Symbol   string  `json:"symbol"`                    // 货币符号
	Rate     float64 `json:"rate"`                      // 汇率
	IsActive bool    `json:"is_active"`                 // 是否启用
}

// UpdateCurrencyRequest 更新货币请求。
type UpdateCurrencyRequest struct {
	Name     string  `json:"name" binding:"required"`   // 货币名称
	Code     string  `json:"code" binding:"required"`   // 货币代码
	Symbol   string  `json:"symbol"`                    // 货币符号
	Rate     float64 `json:"rate"`                      // 汇率
	IsActive bool    `json:"is_active"`                 // 是否启用
}

// CurrencyResponse 货币响应。
type CurrencyResponse struct {
	ID        uint    `json:"id"`         // 货币ID
	Name      string  `json:"name"`       // 货币名称
	Code      string  `json:"code"`       // 货币代码
	Symbol    string  `json:"symbol"`     // 货币符号
	Rate      float64 `json:"rate"`       // 汇率
	IsActive  bool    `json:"is_active"`  // 是否启用
	CreatedAt string  `json:"created_at"` // 创建时间
	UpdatedAt string  `json:"updated_at"` // 更新时间
}

// CurrencyRateItem 汇率更新项。
type CurrencyRateItem struct {
	CurrencyID uint    `json:"currency_id"` // 货币ID
	Rate       float64 `json:"rate"`        // 汇率
}

// ==================== 国家方法 ====================

// ListCountries 获取国家列表。
func (s *DirectoryService) ListCountries(ctx context.Context, page, size int) ([]CountryResponse, int64, error) {
	countries, total, err := s.countryUC.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]CountryResponse, len(countries))
	for i, c := range countries {
		items[i] = CountryResponse{
			ID: c.ID, Name: c.Name, IsoCode2: c.IsoCode2, IsoCode3: c.IsoCode3,
			AddressFormat: c.AddressFormat, PostcodeRequired: c.PostcodeRequired,
			CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// CreateCountry 创建国家。
func (s *DirectoryService) CreateCountry(ctx context.Context, req CreateCountryRequest) (*CountryResponse, error) {
	c, err := s.countryUC.Create(ctx, req.Name, req.IsoCode2, req.IsoCode3, req.AddressFormat, req.PostcodeRequired)
	if err != nil {
		return nil, err
	}
	return &CountryResponse{
		ID: c.ID, Name: c.Name, IsoCode2: c.IsoCode2, IsoCode3: c.IsoCode3,
		AddressFormat: c.AddressFormat, PostcodeRequired: c.PostcodeRequired,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateCountry 更新国家。
func (s *DirectoryService) UpdateCountry(ctx context.Context, id uint, req UpdateCountryRequest) (*CountryResponse, error) {
	c, err := s.countryUC.Update(ctx, id, req.Name, req.IsoCode2, req.IsoCode3, req.AddressFormat, req.PostcodeRequired)
	if err != nil {
		return nil, err
	}
	return &CountryResponse{
		ID: c.ID, Name: c.Name, IsoCode2: c.IsoCode2, IsoCode3: c.IsoCode3,
		AddressFormat: c.AddressFormat, PostcodeRequired: c.PostcodeRequired,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteCountry 删除国家。
func (s *DirectoryService) DeleteCountry(ctx context.Context, id uint) error {
	return s.countryUC.Delete(ctx, id)
}

// ==================== 省/州方法 ====================

// ListStates 获取指定国家下的省/州列表。
func (s *DirectoryService) ListStates(ctx context.Context, countryID uint, page, size int) ([]StateResponse, int64, error) {
	states, total, err := s.stateUC.ListByCountryID(ctx, countryID, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]StateResponse, len(states))
	for i, st := range states {
		items[i] = StateResponse{
			ID: st.ID, CountryID: st.CountryID, Name: st.Name, IsoCode: st.IsoCode,
			CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// CreateState 创建省/州。
func (s *DirectoryService) CreateState(ctx context.Context, req CreateStateRequest) (*StateResponse, error) {
	st, err := s.stateUC.Create(ctx, req.CountryID, req.Name, req.IsoCode)
	if err != nil {
		return nil, err
	}
	return &StateResponse{
		ID: st.ID, CountryID: st.CountryID, Name: st.Name, IsoCode: st.IsoCode,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateState 更新省/州。
func (s *DirectoryService) UpdateState(ctx context.Context, id uint, req UpdateStateRequest) (*StateResponse, error) {
	st, err := s.stateUC.Update(ctx, id, req.Name, req.IsoCode)
	if err != nil {
		return nil, err
	}
	return &StateResponse{
		ID: st.ID, CountryID: st.CountryID, Name: st.Name, IsoCode: st.IsoCode,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteState 删除省/州。
func (s *DirectoryService) DeleteState(ctx context.Context, id uint) error {
	return s.stateUC.Delete(ctx, id)
}

// ==================== 货币方法 ====================

// ListCurrencies 获取货币列表。
func (s *DirectoryService) ListCurrencies(ctx context.Context, page, size int) ([]CurrencyResponse, int64, error) {
	currencies, total, err := s.currencyUC.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]CurrencyResponse, len(currencies))
	for i, cu := range currencies {
		items[i] = CurrencyResponse{
			ID: cu.ID, Name: cu.Name, Code: cu.Code, Symbol: cu.Symbol,
			Rate: cu.Rate, IsActive: cu.IsActive,
			CreatedAt: cu.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: cu.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// CreateCurrency 创建货币。
func (s *DirectoryService) CreateCurrency(ctx context.Context, req CreateCurrencyRequest) (*CurrencyResponse, error) {
	cu, err := s.currencyUC.Create(ctx, req.Name, req.Code, req.Symbol, req.Rate, req.IsActive)
	if err != nil {
		return nil, err
	}
	return &CurrencyResponse{
		ID: cu.ID, Name: cu.Name, Code: cu.Code, Symbol: cu.Symbol,
		Rate: cu.Rate, IsActive: cu.IsActive,
		CreatedAt: cu.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: cu.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateCurrency 更新货币。
func (s *DirectoryService) UpdateCurrency(ctx context.Context, id uint, req UpdateCurrencyRequest) (*CurrencyResponse, error) {
	cu, err := s.currencyUC.Update(ctx, id, req.Name, req.Code, req.Symbol, req.Rate, req.IsActive)
	if err != nil {
		return nil, err
	}
	return &CurrencyResponse{
		ID: cu.ID, Name: cu.Name, Code: cu.Code, Symbol: cu.Symbol,
		Rate: cu.Rate, IsActive: cu.IsActive,
		CreatedAt: cu.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: cu.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteCurrency 删除货币。
func (s *DirectoryService) DeleteCurrency(ctx context.Context, id uint) error {
	return s.currencyUC.Delete(ctx, id)
}

// ApplyRates 批量应用汇率更新。
func (s *DirectoryService) ApplyRates(ctx context.Context, rates []CurrencyRateItem) error {
	bizRates := make([]biz.CurrencyRateItem, len(rates))
	for i, r := range rates {
		bizRates[i] = biz.CurrencyRateItem{CurrencyID: r.CurrencyID, Rate: r.Rate}
	}
	return s.currencyUC.ApplyRates(ctx, bizRates)
}
