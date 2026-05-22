// Package service 提供 payment 服务的依赖注入与初始化
package service

import (
	"context"

	"nop-go/services/payment/internal/biz"
	"nop-go/services/payment/internal/data"
	"nop-go/services/payment/internal/server/http/handler"
	"nop-go/services/payment/internal/server/http/request"
	"nop-go/services/payment/internal/server/http/response"

	"gorm.io/gorm"
)

// Services 支付服务集合，持有所有子服务引用
type Services struct {
	Payment *PaymentService
}

// NewServices 创建支付服务集合
func NewServices(db *gorm.DB) *Services {
	// 初始化数据层仓储
	methodRepo := data.NewPaymentMethodRepo(db)
	restrictionRepo := data.NewMethodRestrictionRepo(db)

	// 初始化业务层用例
	uc := biz.NewPaymentUsecase(methodRepo, restrictionRepo)

	return &Services{
		Payment: &PaymentService{uc: uc},
	}
}

// PaymentService 支付服务，封装业务用例
type PaymentService struct {
	uc *biz.PaymentUsecase
}

// NewPaymentHandler 创建支付服务 HTTP 处理器
func NewPaymentHandler(svc *PaymentService) *handler.PaymentHandler {
	return handler.NewPaymentHandler(svc.uc)
}

// ListPaymentMethods 获取支付方式列表
func (s *PaymentService) ListPaymentMethods(ctx context.Context, page, pageSize int) ([]*response.PaymentMethodResponse, int64, error) {
	return s.uc.ListPaymentMethods(ctx, page, pageSize)
}

// UpdatePaymentMethod 更新支付方式
func (s *PaymentService) UpdatePaymentMethod(ctx context.Context, req request.UpdatePaymentMethodRequest) (*response.PaymentMethodResponse, error) {
	return s.uc.UpdatePaymentMethod(ctx, req)
}

// ListMethodRestrictions 获取支付方式限制列表
func (s *PaymentService) ListMethodRestrictions(ctx context.Context, page, pageSize int) ([]*response.MethodRestrictionResponse, int64, error) {
	return s.uc.ListMethodRestrictions(ctx, page, pageSize)
}

// UpdateMethodRestrictions 更新支付方式限制
func (s *PaymentService) UpdateMethodRestrictions(ctx context.Context, req request.UpdateMethodRestrictionsRequest) (*response.MethodRestrictionResponse, error) {
	return s.uc.UpdateMethodRestrictions(ctx, req)
}
