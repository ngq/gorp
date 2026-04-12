//go:build wireinject

package main

import (
	"nop-go/services/order-service/internal/biz"
	"nop-go/services/order-service/internal/data"
	"nop-go/services/order-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireOrderService(db *gorm.DB) (*service.OrderService, error) {
	panic(wire.Build(
		data.NewOrderRepository,
		data.NewOrderItemRepository,
		data.NewOrderAddressRepository,
		data.NewGiftCardRepository,
		data.NewReturnRequestRepository,
		biz.NewOrderUseCase,
		biz.NewGiftCardUseCase,
		biz.NewReturnRequestUseCase,
		service.NewOrderService,
	))
}
