// Package biz 提供 payment 服务的业务逻辑层
//
// 支付服务包含两大业务领域：
// 1. 支付方式（PaymentMethod）— 管理系统支持的支付方式及其属性
// 2. 支付方式限制（MethodRestriction）— 管理支付方式的使用限制规则
package biz

import (
	"context"
	"time"

	"nop-go/services/payment/internal/server/http/request"
	"nop-go/services/payment/internal/server/http/response"
)

// ==================== 领域实体定义 ====================

// PaymentMethod 支付方式领域实体
type PaymentMethod struct {
	ID                    uint      // 支付方式 ID
	Name                  string    // 支付方式名称
	SystemKeyword         string    // 系统关键字标识
	DisplayOrder          int       // 显示排序
	IsActive              bool      // 是否启用
	LogoURL               string    // Logo 地址
	SupportsRefund        bool      // 是否支持退款
	SupportsPartialRefund bool      // 是否支持部分退款
	CreatedAt             time.Time // 创建时间
	UpdatedAt             time.Time // 更新时间
}

// MethodRestriction 支付方式限制领域实体
type MethodRestriction struct {
	ID               uint      // 限制规则 ID
	PaymentMethodID  uint      // 关联的支付方式 ID
	MinOrderAmount   float64   // 最小订单金额
	MaxOrderAmount   float64   // 最大订单金额
	RestrictionType  string    // 限制类型
	RestrictionValue string    // 限制值
	IsActive         bool      // 是否启用
	CreatedAt        time.Time // 创建时间
	UpdatedAt        time.Time // 更新时间
}

// ==================== 仓储接口定义 ====================

// PaymentMethodRepository 支付方式仓储接口
type PaymentMethodRepository interface {
	// List 获取支付方式列表（分页）
	List(ctx context.Context, page, pageSize int) ([]*PaymentMethod, int64, error)
	// Update 更新支付方式
	Update(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
}

// MethodRestrictionRepository 支付方式限制仓储接口
type MethodRestrictionRepository interface {
	// List 获取支付方式限制列表（分页）
	List(ctx context.Context, page, pageSize int) ([]*MethodRestriction, int64, error)
	// Update 更新支付方式限制
	Update(ctx context.Context, restriction *MethodRestriction) (*MethodRestriction, error)
}

// ==================== 用例实现 ====================

// PaymentUsecase 支付服务业务用例
type PaymentUsecase struct {
	methodRepo      PaymentMethodRepository
	restrictionRepo MethodRestrictionRepository
}

// NewPaymentUsecase 创建支付服务业务用例
func NewPaymentUsecase(methodRepo PaymentMethodRepository, restrictionRepo MethodRestrictionRepository) *PaymentUsecase {
	return &PaymentUsecase{
		methodRepo:      methodRepo,
		restrictionRepo: restrictionRepo,
	}
}

// ListPaymentMethods 获取支付方式列表
func (uc *PaymentUsecase) ListPaymentMethods(ctx context.Context, page, pageSize int) ([]*response.PaymentMethodResponse, int64, error) {
	// 调用仓储获取领域实体列表
	methods, total, err := uc.methodRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 领域实体转换为响应 DTO
	items := make([]*response.PaymentMethodResponse, len(methods))
	for i, m := range methods {
		items[i] = toPaymentMethodResponse(m)
	}
	return items, total, nil
}

// UpdatePaymentMethod 更新支付方式
func (uc *PaymentUsecase) UpdatePaymentMethod(ctx context.Context, req request.UpdatePaymentMethodRequest) (*response.PaymentMethodResponse, error) {
	// 请求 DTO 转换为领域实体
	method := &PaymentMethod{
		ID:                    req.ID,
		Name:                  req.Name,
		SystemKeyword:         req.SystemKeyword,
		DisplayOrder:          req.DisplayOrder,
		IsActive:              req.IsActive,
		LogoURL:               req.LogoURL,
		SupportsRefund:        req.SupportsRefund,
		SupportsPartialRefund: req.SupportsPartialRefund,
	}

	// 调用仓储更新
	updated, err := uc.methodRepo.Update(ctx, method)
	if err != nil {
		return nil, err
	}

	return toPaymentMethodResponse(updated), nil
}

// ListMethodRestrictions 获取支付方式限制列表
func (uc *PaymentUsecase) ListMethodRestrictions(ctx context.Context, page, pageSize int) ([]*response.MethodRestrictionResponse, int64, error) {
	// 调用仓储获取领域实体列表
	restrictions, total, err := uc.restrictionRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 领域实体转换为响应 DTO
	items := make([]*response.MethodRestrictionResponse, len(restrictions))
	for i, r := range restrictions {
		items[i] = toMethodRestrictionResponse(r)
	}
	return items, total, nil
}

// UpdateMethodRestrictions 更新支付方式限制
func (uc *PaymentUsecase) UpdateMethodRestrictions(ctx context.Context, req request.UpdateMethodRestrictionsRequest) (*response.MethodRestrictionResponse, error) {
	// 请求 DTO 转换为领域实体
	restriction := &MethodRestriction{
		ID:               req.ID,
		PaymentMethodID:  req.PaymentMethodID,
		MinOrderAmount:   req.MinOrderAmount,
		MaxOrderAmount:   req.MaxOrderAmount,
		RestrictionType:  req.RestrictionType,
		RestrictionValue: req.RestrictionValue,
		IsActive:         req.IsActive,
	}

	// 调用仓储更新
	updated, err := uc.restrictionRepo.Update(ctx, restriction)
	if err != nil {
		return nil, err
	}

	return toMethodRestrictionResponse(updated), nil
}

// ==================== 内部转换函数 ====================

// toPaymentMethodResponse 领域实体转换为支付方式响应 DTO
func toPaymentMethodResponse(m *PaymentMethod) *response.PaymentMethodResponse {
	return &response.PaymentMethodResponse{
		ID:                    m.ID,
		Name:                  m.Name,
		SystemKeyword:         m.SystemKeyword,
		DisplayOrder:          m.DisplayOrder,
		IsActive:              m.IsActive,
		LogoURL:               m.LogoURL,
		SupportsRefund:        m.SupportsRefund,
		SupportsPartialRefund: m.SupportsPartialRefund,
		CreatedAt:             m.CreatedAt.Unix(),
		UpdatedAt:             m.UpdatedAt.Unix(),
	}
}

// toMethodRestrictionResponse 领域实体转换为支付方式限制响应 DTO
func toMethodRestrictionResponse(r *MethodRestriction) *response.MethodRestrictionResponse {
	return &response.MethodRestrictionResponse{
		ID:               r.ID,
		PaymentMethodID:  r.PaymentMethodID,
		MinOrderAmount:   r.MinOrderAmount,
		MaxOrderAmount:   r.MaxOrderAmount,
		RestrictionType:  r.RestrictionType,
		RestrictionValue: r.RestrictionValue,
		IsActive:         r.IsActive,
		CreatedAt:        r.CreatedAt.Unix(),
		UpdatedAt:        r.UpdatedAt.Unix(),
	}
}
