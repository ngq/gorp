// Package service 提供 tax 服务的依赖注入与初始化
package service

import (
	"nop-go/services/tax/internal/biz"
	"nop-go/services/tax/internal/data"
	"nop-go/services/tax/internal/server/http/handler"

	"gorm.io/gorm"
)

// TaxService 税务服务，封装业务用例与 HTTP 处理器
type TaxService struct {
	uc      *biz.TaxUsecase
	Handler *handler.TaxHandler
}

// NewTaxService 创建税务服务实例
// 注入数据库连接，初始化所有子领域仓储、用例和处理器
func NewTaxService(db *gorm.DB) *TaxService {
	// 初始化数据层仓储
	providerRepo := data.NewProviderRepo(db)
	categoryRepo := data.NewCategoryRepo(db)

	// 初始化业务层用例
	uc := biz.NewTaxUsecase(providerRepo, categoryRepo)

	// 初始化 HTTP 处理器
	h := handler.NewTaxHandler(uc)

	return &TaxService{uc: uc, Handler: h}
}