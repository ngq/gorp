//go:build wireinject

package main

import (
	"nop-go/services/catalog-service/internal/biz"
	"nop-go/services/catalog-service/internal/data"
	"nop-go/services/catalog-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireCatalogService 使用 Wire 生成 catalog-service 的装配代码。
func wireCatalogService(db *gorm.DB) (*service.CatalogService, error) {
	panic(wire.Build(
		data.NewProductRepository,
		data.NewCategoryRepository,
		data.NewManufacturerRepository,
		data.NewProductPictureRepository,
		data.NewProductReviewRepository,
		biz.NewProductUseCase,
		biz.NewCategoryUseCase,
		biz.NewManufacturerUseCase,
		biz.NewProductPictureUseCase,
		biz.NewProductReviewUseCase,
		provideCatalogService,
	))
}

func provideCatalogService(productUC *biz.ProductUseCase, categoryUC *biz.CategoryUseCase, manufacturerUC *biz.ManufacturerUseCase, pictureUC *biz.ProductPictureUseCase, reviewUC *biz.ProductReviewUseCase) *service.CatalogService {
	return service.NewCatalogService(productUC, categoryUC, manufacturerUC, pictureUC, reviewUC, service.Options{
		EnableMedia:  true,
		EnableReview: true,
	})
}
