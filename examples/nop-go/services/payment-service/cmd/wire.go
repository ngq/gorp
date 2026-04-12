//go:build wireinject

package main

import (
	"nop-go/shared/plugin"
	"nop-go/services/payment-service/internal/biz"
	"nop-go/services/payment-service/internal/data"
	"nop-go/services/payment-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wirePaymentService(db *gorm.DB, registry *plugin.Registry) (*service.PaymentService, error) {
	panic(wire.Build(
		data.NewPaymentRepository,
		data.NewPaymentTransactionRepository,
		data.NewRefundRepository,
		biz.NewPaymentUseCase,
		service.NewPaymentService,
	))
}
