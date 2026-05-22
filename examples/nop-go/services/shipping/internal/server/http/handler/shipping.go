// Package handler 提供 shipping 服务的 HTTP 请求处理器
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/shipping/internal/biz"
	"nop-go/services/shipping/internal/server/http/request"
	"nop-go/services/shipping/internal/server/http/response"
)

// ShippingHandler 配送服务 HTTP 处理器
type ShippingHandler struct {
	uc *biz.ShippingUsecase
}

// NewShippingHandler 创建配送服务处理器
func NewShippingHandler(uc *biz.ShippingUsecase) *ShippingHandler {
	return &ShippingHandler{uc: uc}
}

// ==================== 配送提供者 ====================

// ListProviders 获取配送提供者列表
// GET /api/v1/shipping/providers
func (h *ShippingHandler) ListProviders(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取配送提供者列表
	providers, total, err := h.uc.ListProviders(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListProvidersResponse{
		Total: total,
		Items: providers,
	})
}

// UpdateProvider 更新配送提供者
// PUT /api/v1/shipping/providers/:id
func (h *ShippingHandler) UpdateProvider(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateProviderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新配送提供者
	provider, err := h.uc.UpdateProvider(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, provider)
}

// ==================== 配送方式 ====================

// ListMethods 获取配送方式列表
// GET /api/v1/shipping/methods
func (h *ShippingHandler) ListMethods(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取配送方式列表
	methods, total, err := h.uc.ListMethods(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListMethodsResponse{
		Total: total,
		Items: methods,
	})
}

// CreateMethod 创建配送方式
// POST /api/v1/shipping/methods
func (h *ShippingHandler) CreateMethod(c gorp.Context) {
	// 解析请求体
	var req request.CreateMethodRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 调用业务层创建配送方式
	method, err := h.uc.CreateMethod(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, method)
}

// UpdateMethod 更新配送方式
// PUT /api/v1/shipping/methods/:id
func (h *ShippingHandler) UpdateMethod(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateMethodRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新配送方式
	method, err := h.uc.UpdateMethod(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, method)
}

// DeleteMethod 删除配送方式
// DELETE /api/v1/shipping/methods/:id
func (h *ShippingHandler) DeleteMethod(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 调用业务层删除配送方式
	if err := h.uc.DeleteMethod(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ==================== 配送日期 ====================

// ListDeliveryDates 获取配送日期列表
// GET /api/v1/shipping/delivery-dates
func (h *ShippingHandler) ListDeliveryDates(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取配送日期列表
	dates, total, err := h.uc.ListDeliveryDates(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListDeliveryDatesResponse{
		Total: total,
		Items: dates,
	})
}

// CreateDeliveryDate 创建配送日期
// POST /api/v1/shipping/delivery-dates
func (h *ShippingHandler) CreateDeliveryDate(c gorp.Context) {
	// 解析请求体
	var req request.CreateDeliveryDateRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 调用业务层创建配送日期
	date, err := h.uc.CreateDeliveryDate(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, date)
}

// UpdateDeliveryDate 更新配送日期
// PUT /api/v1/shipping/delivery-dates/:id
func (h *ShippingHandler) UpdateDeliveryDate(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateDeliveryDateRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新配送日期
	date, err := h.uc.UpdateDeliveryDate(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, date)
}

// ==================== 仓库 ====================

// ListWarehouses 获取仓库列表
// GET /api/v1/shipping/warehouses
func (h *ShippingHandler) ListWarehouses(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取仓库列表
	warehouses, total, err := h.uc.ListWarehouses(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListWarehousesResponse{
		Total: total,
		Items: warehouses,
	})
}

// CreateWarehouse 创建仓库
// POST /api/v1/shipping/warehouses
func (h *ShippingHandler) CreateWarehouse(c gorp.Context) {
	// 解析请求体
	var req request.CreateWarehouseRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 调用业务层创建仓库
	warehouse, err := h.uc.CreateWarehouse(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, warehouse)
}

// UpdateWarehouse 更新仓库
// PUT /api/v1/shipping/warehouses/:id
func (h *ShippingHandler) UpdateWarehouse(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateWarehouseRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新仓库
	warehouse, err := h.uc.UpdateWarehouse(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, warehouse)
}

// ==================== 运费估算 ====================

// EstimateShipping 估算运费
// GET /api/v1/shipping/estimate
func (h *ShippingHandler) EstimateShipping(c gorp.Context) {
	// 解析查询参数
	req := request.EstimateShippingRequest{
		WarehouseID: c.Query("warehouse_id"),
		CountryID:   c.Query("country_id"),
		StateID:     c.Query("state_id"),
		ZipCode:     c.Query("zip_code"),
		SubTotal:    c.DefaultQuery("sub_total", "0"),
		Weight:      c.DefaultQuery("weight", "0"),
	}

	// 调用业务层估算运费
	estimates, err := h.uc.EstimateShipping(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, estimates)
}

// ==================== 内部辅助函数 ====================

// parseIDParam 解析路径参数中的 ID，返回 uint64
func parseIDParam(c gorp.Context, key string) (uint, error) {
	idStr := c.Param(key)
	if idStr == "" {
		return 0, errParamRequired(key)
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, errInvalidID(key)
	}
	return uint(id), nil
}

// errParamRequired 返回参数必填错误
type paramRequiredError string

func (e paramRequiredError) Error() string { return string(e) + " 不能为空" }

func errParamRequired(key string) paramRequiredError { return paramRequiredError(key) }

// errInvalidID 返回 ID 无效错误
type invalidIDError string

func (e invalidIDError) Error() string { return "无效的 " + string(e) }

func errInvalidID(key string) invalidIDError { return invalidIDError(key) }
