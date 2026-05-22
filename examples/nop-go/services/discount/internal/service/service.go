// Package service 提供 discount 服务的依赖注入与初始化
package service

import (
	"nop-go/services/discount/internal/biz"
	"nop-go/services/discount/internal/data"
	"nop-go/services/discount/internal/server/http/handler"

	"gorm.io/gorm"
)

// DiscountService 折扣服务，封装业务用例与 HTTP 处理器
type DiscountService struct {
	uc      *biz.DiscountUsecase
	Handler *handler.DiscountHandler
}

// NewDiscountService 创建折扣服务实例
// 注入数据库连接，初始化所有子领域仓储、用例和处理器
func NewDiscountService(db *gorm.DB) *DiscountService {
	// 初始化数据层仓储
	discountRepo := data.NewDiscountRepo(db)
	productRepo := data.NewDiscountProductRepo(db)
	categoryRepo := data.NewDiscountCategoryRepo(db)
	manufacturerRepo := data.NewDiscountManufacturerRepo(db)
	usageHistoryRepo := data.NewDiscountUsageHistoryRepo(db)

	// 初始化业务层用例
	uc := biz.NewDiscountUsecase(discountRepo, productRepo, categoryRepo, manufacturerRepo, usageHistoryRepo)

	// 初始化 HTTP 处理器
	h := handler.NewDiscountHandler(uc)

	return &DiscountService{uc: uc, Handler: h}
}