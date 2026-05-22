// Package handler 提供 payment 服务的 HTTP 请求处理器
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/payment/internal/biz"
	"nop-go/services/payment/internal/server/http/request"
	"nop-go/services/payment/internal/server/http/response"
)

// PaymentHandler 支付服务 HTTP 处理器
type PaymentHandler struct {
	uc *biz.PaymentUsecase
}

// NewPaymentHandler 创建支付服务处理器
func NewPaymentHandler(uc *biz.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

// ListPaymentMethods 获取支付方式列表
// GET /api/v1/payment/methods
func (h *PaymentHandler) ListPaymentMethods(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取支付方式列表
	methods, total, err := h.uc.ListPaymentMethods(c, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 构建分页响应
	gorp.Success(c, response.ListPaymentMethodsResponse{
		Total: total,
		Items: methods,
	})
}

// UpdatePaymentMethod 更新支付方式
// PUT /api/v1/payment/methods/:id
func (h *PaymentHandler) UpdatePaymentMethod(c gorp.Context) {
	// 获取路径参数 ID
	idStr := c.Param("id")
	if idStr == "" {
		gorp.BadRequest(c, "支付方式 ID 不能为空")
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的支付方式 ID")
		return
	}

	// 解析请求体
	var req request.UpdatePaymentMethodRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = uint(id)

	// 调用业务层更新支付方式
	method, err := h.uc.UpdatePaymentMethod(c, req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, method)
}

// ListMethodRestrictions 获取支付方式限制列表
// GET /api/v1/payment/method-restrictions
func (h *PaymentHandler) ListMethodRestrictions(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取支付方式限制列表
	restrictions, total, err := h.uc.ListMethodRestrictions(c, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 构建分页响应
	gorp.Success(c, response.ListMethodRestrictionsResponse{
		Total: total,
		Items: restrictions,
	})
}

// UpdateMethodRestrictions 更新支付方式限制
// PUT /api/v1/payment/method-restrictions
func (h *PaymentHandler) UpdateMethodRestrictions(c gorp.Context) {
	// 解析请求体
	var req request.UpdateMethodRestrictionsRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 调用业务层更新支付方式限制
	restriction, err := h.uc.UpdateMethodRestrictions(c, req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, restriction)
}
