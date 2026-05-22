package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/directory/internal/server/http/request"
	"nop-go/services/directory/internal/server/http/response"
	"nop-go/services/directory/internal/service"
)

// ==================== 国家处理器 ====================

// CountryHandler 国家 HTTP 处理器。
type CountryHandler struct {
	directory *service.DirectoryService
}

// NewCountryHandler 创建国家处理器。
func NewCountryHandler(directory *service.DirectoryService) *CountryHandler {
	return &CountryHandler{directory: directory}
}

// List 国家列表。
func (h *CountryHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.directory.ListCountries(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Country, len(items))
	for i, item := range items {
		respItems[i] = toCountryResponse(&item)
	}

	gorp.Success(c, response.CountryList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建国家。
func (h *CountryHandler) Create(c gorp.Context) {
	var req request.CreateCountry
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	item, err := h.directory.CreateCountry(c.Context(), service.CreateCountryRequest{
		Name:           req.Name,
		IsoCode2:       req.IsoCode2,
		IsoCode3:       req.IsoCode3,
		AddressFormat:  req.AddressFormat,
		PostcodeRequired: req.PostcodeRequired,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, toCountryResponse(item))
}

// Update 更新国家。
func (h *CountryHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateCountry
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	item, err := h.directory.UpdateCountry(c.Context(), uint(id), service.UpdateCountryRequest{
		Name:           req.Name,
		IsoCode2:       req.IsoCode2,
		IsoCode3:       req.IsoCode3,
		AddressFormat:  req.AddressFormat,
		PostcodeRequired: req.PostcodeRequired,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, toCountryResponse(item))
}

// Delete 删除国家。
func (h *CountryHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.directory.DeleteCountry(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// toCountryResponse 将 service.CountryResponse 转换为 response.Country。
func toCountryResponse(src *service.CountryResponse) response.Country {
	return response.Country{
		ID:               src.ID,
		Name:             src.Name,
		IsoCode2:         src.IsoCode2,
		IsoCode3:         src.IsoCode3,
		AddressFormat:    src.AddressFormat,
		PostcodeRequired: src.PostcodeRequired,
		CreatedAt:        src.CreatedAt,
		UpdatedAt:        src.UpdatedAt,
	}
}

// ==================== 省/州处理器 ====================

// StateHandler 省/州 HTTP 处理器。
type StateHandler struct {
	directory *service.DirectoryService
}

// NewStateHandler 创建省/州处理器。
func NewStateHandler(directory *service.DirectoryService) *StateHandler {
	return &StateHandler{directory: directory}
}

// ListByCountry 获取指定国家下的省/州列表。
func (h *StateHandler) ListByCountry(c gorp.Context) {
	countryID, err := strconv.ParseUint(c.Param("country_id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的国家ID参数")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.directory.ListStates(c.Context(), uint(countryID), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.State, len(items))
	for i, item := range items {
		respItems[i] = toStateResponse(&item)
	}

	gorp.Success(c, response.StateList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建省/州。
func (h *StateHandler) Create(c gorp.Context) {
	countryID, err := strconv.ParseUint(c.Param("country_id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的国家ID参数")
		return
	}

	var req request.CreateState
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	item, err := h.directory.CreateState(c.Context(), service.CreateStateRequest{
		CountryID: uint(countryID),
		Name:      req.Name,
		IsoCode:   req.IsoCode,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, toStateResponse(item))
}

// Update 更新省/州。
func (h *StateHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateState
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	item, err := h.directory.UpdateState(c.Context(), uint(id), service.UpdateStateRequest{
		Name:    req.Name,
		IsoCode: req.IsoCode,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, toStateResponse(item))
}

// Delete 删除省/州。
func (h *StateHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.directory.DeleteState(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// toStateResponse 将 service.StateResponse 转换为 response.State。
func toStateResponse(src *service.StateResponse) response.State {
	return response.State{
		ID:        src.ID,
		CountryID: src.CountryID,
		Name:      src.Name,
		IsoCode:   src.IsoCode,
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
	}
}

// ==================== 货币处理器 ====================

// CurrencyHandler 货币 HTTP 处理器。
type CurrencyHandler struct {
	directory *service.DirectoryService
}

// NewCurrencyHandler 创建货币处理器。
func NewCurrencyHandler(directory *service.DirectoryService) *CurrencyHandler {
	return &CurrencyHandler{directory: directory}
}

// List 货币列表。
func (h *CurrencyHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.directory.ListCurrencies(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Currency, len(items))
	for i, item := range items {
		respItems[i] = toCurrencyResponse(&item)
	}

	gorp.Success(c, response.CurrencyList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建货币。
func (h *CurrencyHandler) Create(c gorp.Context) {
	var req request.CreateCurrency
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	item, err := h.directory.CreateCurrency(c.Context(), service.CreateCurrencyRequest{
		Name:     req.Name,
		Code:     req.Code,
		Symbol:   req.Symbol,
		Rate:     req.Rate,
		IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, toCurrencyResponse(item))
}

// Update 更新货币。
func (h *CurrencyHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateCurrency
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	item, err := h.directory.UpdateCurrency(c.Context(), uint(id), service.UpdateCurrencyRequest{
		Name:     req.Name,
		Code:     req.Code,
		Symbol:   req.Symbol,
		Rate:     req.Rate,
		IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, toCurrencyResponse(item))
}

// Delete 删除货币。
func (h *CurrencyHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.directory.DeleteCurrency(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ApplyRates 应用汇率更新。
func (h *CurrencyHandler) ApplyRates(c gorp.Context) {
	var req request.ApplyRates
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	rates := make([]service.CurrencyRateItem, len(req.Rates))
	for i, r := range req.Rates {
		rates[i] = service.CurrencyRateItem{
			CurrencyID: r.CurrencyID,
			Rate:       r.Rate,
		}
	}

	if err := h.directory.ApplyRates(c.Context(), rates); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]string{"message": "汇率已更新"})
}

// toCurrencyResponse 将 service.CurrencyResponse 转换为 response.Currency。
func toCurrencyResponse(src *service.CurrencyResponse) response.Currency {
	return response.Currency{
		ID:        src.ID,
		Name:      src.Name,
		Code:      src.Code,
		Symbol:    src.Symbol,
		Rate:      src.Rate,
		IsActive:  src.IsActive,
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
	}
}
