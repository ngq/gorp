// Package handler 目录相关 HTTP 处理器。
// 包含国家、省/州、货币管理接口。
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/catalog-service/internal/server/http/request"
	"nop-go/services/catalog-service/internal/server/http/response"
	"nop-go/services/catalog-service/internal/service"
)

// ---------------------------------------------------------------------------
// 国家处理器
// ---------------------------------------------------------------------------

// CountryHandler 国家 HTTP 处理器。
type CountryHandler struct {
	directory *service.DirectoryService
}

// NewCountryHandler 创建国家处理器。
func NewCountryHandler(directory *service.DirectoryService) *CountryHandler {
	return &CountryHandler{directory: directory}
}

// List 获取国家列表。
// 路由：GET /countries
func (h *CountryHandler) List(c gorp.Context) {
	var req request.ListCountryRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.directory.ListCountries(c.Context(), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Country, len(items))
	for i, item := range items {
		respItems[i] = response.Country{
			ID:               item.ID,
			Name:             item.Name,
			IsoCode2:         item.IsoCode2,
			IsoCode3:         item.IsoCode3,
			AddressFormat:    item.AddressFormat,
			PostcodeRequired: item.PostcodeRequired,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		}
	}

	gorp.Success(c, response.CountryList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// Create 创建国家。
// 路由：POST /countries
func (h *CountryHandler) Create(c gorp.Context) {
	var req request.CreateCountry
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	country, err := h.directory.CreateCountry(c.Context(), service.CreateCountryRequest{
		Name: req.Name, IsoCode2: req.IsoCode2, IsoCode3: req.IsoCode3,
		AddressFormat: req.AddressFormat, PostcodeRequired: req.PostcodeRequired,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Country{
		ID:               country.ID,
		Name:             country.Name,
		IsoCode2:         country.IsoCode2,
		IsoCode3:         country.IsoCode3,
		AddressFormat:    country.AddressFormat,
		PostcodeRequired: country.PostcodeRequired,
		CreatedAt:        country.CreatedAt,
		UpdatedAt:        country.UpdatedAt,
	})
}

// Update 更新国家。
// 路由：PUT /countries/:id
func (h *CountryHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的国家ID")
		return
	}

	var req request.UpdateCountry
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	country, err := h.directory.UpdateCountry(c.Context(), uint(id), service.UpdateCountryRequest{
		Name: req.Name, IsoCode2: req.IsoCode2, IsoCode3: req.IsoCode3,
		AddressFormat: req.AddressFormat, PostcodeRequired: req.PostcodeRequired,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Country{
		ID:               country.ID,
		Name:             country.Name,
		IsoCode2:         country.IsoCode2,
		IsoCode3:         country.IsoCode3,
		AddressFormat:    country.AddressFormat,
		PostcodeRequired: country.PostcodeRequired,
		CreatedAt:        country.CreatedAt,
		UpdatedAt:        country.UpdatedAt,
	})
}

// Delete 删除国家。
// 路由：DELETE /countries/:id
func (h *CountryHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的国家ID")
		return
	}

	if err := h.directory.DeleteCountry(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"deleted": true})
}

// ---------------------------------------------------------------------------
// 省/州处理器
// ---------------------------------------------------------------------------

// StateHandler 省/州 HTTP 处理器。
type StateHandler struct {
	directory *service.DirectoryService
}

// NewStateHandler 创建省/州处理器。
func NewStateHandler(directory *service.DirectoryService) *StateHandler {
	return &StateHandler{directory: directory}
}

// ListByCountry 获取指定国家下的省/州列表。
// 路由：GET /countries/:country_id/states
func (h *StateHandler) ListByCountry(c gorp.Context) {
	countryID, err := strconv.ParseUint(c.Param("country_id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的国家ID")
		return
	}

	var req request.ListStateRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.directory.ListStates(c.Context(), uint(countryID), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.State, len(items))
	for i, item := range items {
		respItems[i] = response.State{
			ID:        item.ID,
			CountryID: item.CountryID,
			Name:      item.Name,
			IsoCode:   item.IsoCode,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	gorp.Success(c, response.StateList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// Create 创建省/州。
// 路由：POST /countries/:country_id/states
func (h *StateHandler) Create(c gorp.Context) {
	countryID, err := strconv.ParseUint(c.Param("country_id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的国家ID")
		return
	}

	var req request.CreateState
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	state, err := h.directory.CreateState(c.Context(), service.CreateStateRequest{
		CountryID: uint(countryID), Name: req.Name, IsoCode: req.IsoCode,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.State{
		ID:        state.ID,
		CountryID: state.CountryID,
		Name:      state.Name,
		IsoCode:   state.IsoCode,
		CreatedAt: state.CreatedAt,
		UpdatedAt: state.UpdatedAt,
	})
}

// Update 更新省/州。
// 路由：PUT /states/:id
func (h *StateHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的省/州ID")
		return
	}

	var req request.UpdateState
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	state, err := h.directory.UpdateState(c.Context(), uint(id), service.UpdateStateRequest{
		Name: req.Name, IsoCode: req.IsoCode,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.State{
		ID:        state.ID,
		CountryID: state.CountryID,
		Name:      state.Name,
		IsoCode:   state.IsoCode,
		CreatedAt: state.CreatedAt,
		UpdatedAt: state.UpdatedAt,
	})
}

// Delete 删除省/州。
// 路由：DELETE /states/:id
func (h *StateHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的省/州ID")
		return
	}

	if err := h.directory.DeleteState(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"deleted": true})
}

// ---------------------------------------------------------------------------
// 货币处理器
// ---------------------------------------------------------------------------

// CurrencyHandler 货币 HTTP 处理器。
type CurrencyHandler struct {
	directory *service.DirectoryService
}

// NewCurrencyHandler 创建货币处理器。
func NewCurrencyHandler(directory *service.DirectoryService) *CurrencyHandler {
	return &CurrencyHandler{directory: directory}
}

// List 获取货币列表。
// 路由：GET /currencies
func (h *CurrencyHandler) List(c gorp.Context) {
	var req request.ListCurrencyRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.directory.ListCurrencies(c.Context(), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Currency, len(items))
	for i, item := range items {
		respItems[i] = response.Currency{
			ID:        item.ID,
			Name:      item.Name,
			Code:      item.Code,
			Symbol:    item.Symbol,
			Rate:      item.Rate,
			IsActive:  item.IsActive,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	gorp.Success(c, response.CurrencyList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// Create 创建货币。
// 路由：POST /currencies
func (h *CurrencyHandler) Create(c gorp.Context) {
	var req request.CreateCurrency
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	currency, err := h.directory.CreateCurrency(c.Context(), service.CreateCurrencyRequest{
		Name: req.Name, Code: req.Code, Symbol: req.Symbol, Rate: req.Rate, IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Currency{
		ID:        currency.ID,
		Name:      currency.Name,
		Code:      currency.Code,
		Symbol:    currency.Symbol,
		Rate:      currency.Rate,
		IsActive:  currency.IsActive,
		CreatedAt: currency.CreatedAt,
		UpdatedAt: currency.UpdatedAt,
	})
}

// Update 更新货币。
// 路由：PUT /currencies/:id
func (h *CurrencyHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的货币ID")
		return
	}

	var req request.UpdateCurrency
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	currency, err := h.directory.UpdateCurrency(c.Context(), uint(id), service.UpdateCurrencyRequest{
		Name: req.Name, Code: req.Code, Symbol: req.Symbol, Rate: req.Rate, IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Currency{
		ID:        currency.ID,
		Name:      currency.Name,
		Code:      currency.Code,
		Symbol:    currency.Symbol,
		Rate:      currency.Rate,
		IsActive:  currency.IsActive,
		CreatedAt: currency.CreatedAt,
		UpdatedAt: currency.UpdatedAt,
	})
}

// Delete 删除货币。
// 路由：DELETE /currencies/:id
func (h *CurrencyHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的货币ID")
		return
	}

	if err := h.directory.DeleteCurrency(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"deleted": true})
}

// ApplyRates 批量应用汇率更新。
// 路由：POST /currencies/apply-rates
func (h *CurrencyHandler) ApplyRates(c gorp.Context) {
	var req request.ApplyRatesRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	rates := make([]service.CurrencyRateItem, len(req.Rates))
	for i, r := range req.Rates {
		rates[i] = service.CurrencyRateItem{CurrencyID: r.CurrencyID, Rate: r.Rate}
	}

	if err := h.directory.ApplyRates(c.Context(), rates); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "汇率已更新"})
}