// Package service 提供 shipping 服务的依赖注入与初始化
package service

import (
	"nop-go/services/shipping/internal/biz"
	"nop-go/services/shipping/internal/data"
	"nop-go/services/shipping/internal/server/http/handler"

	"gorm.io/gorm"
)

// ShippingService 配送服务，封装业务用例与 HTTP 处理器
type ShippingService struct {
	uc      *biz.ShippingUsecase
	Handler *handler.ShippingHandler
}

// NewShippingService 创建配送服务实例
// 注入数据库连接，初始化所有子领域仓储、用例和处理器
func NewShippingService(db *gorm.DB) *ShippingService {
	// 初始化数据层仓储
	providerRepo := data.NewProviderRepo(db)
	methodRepo := data.NewMethodRepo(db)
	deliveryDateRepo := data.NewDeliveryDateRepo(db)
	warehouseRepo := data.NewWarehouseRepo(db)

	// 初始化业务层用例
	uc := biz.NewShippingUsecase(providerRepo, methodRepo, deliveryDateRepo, warehouseRepo)

	// 初始化 HTTP 处理器
	h := handler.NewShippingHandler(uc)

	return &ShippingService{uc: uc, Handler: h}
}
