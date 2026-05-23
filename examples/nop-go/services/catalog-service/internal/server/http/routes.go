// Package http 提供 catalog-service 的 HTTP 路由注册。
// 合并 product + directory + media + seo 四个服务的路由。
// 使用 gorp.Router 抽象接口注册路由，保持与框架解耦。
package http

import (
	svc "nop-go/services/catalog-service/internal/service"
	handler "nop-go/services/catalog-service/internal/server/http/handler"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册 catalog-service 的全部 HTTP 路由。
//
// 路由分组：
//   - /products         商品 CRUD + 搜索 + 最近浏览 + 对比
//   - /categories       分类管理
//   - /manufacturers    制造商管理
//   - /reviews          商品评论
//   - /countries        国家管理
//   - /states           省/州管理
//   - /currencies       货币管理 + 汇率应用
//   - /media            媒体文件
//   - /seo              SEO 元数据
func RegisterRoutes(
	r gorp.Router,
	services *svc.Services,
) {
	// 初始化各处理器
	productHandler := handler.NewProductHandler(services.Product)
	categoryHandler := handler.NewCategoryHandler(services.Product)
	manufacturerHandler := handler.NewManufacturerHandler(services.Product)
	reviewHandler := handler.NewReviewHandler(services.Product)
	countryHandler := handler.NewCountryHandler(services.Directory)
	stateHandler := handler.NewStateHandler(services.Directory)
	currencyHandler := handler.NewCurrencyHandler(services.Directory)
	mediaHandler := handler.NewMediaHandler(services.Media)
	seoHandler := handler.NewSeoHandler(services.Seo)

	// -----------------------------------------------------------------------
	// 商品路由组：/products
	// -----------------------------------------------------------------------
	products := r.Group("/products")
	{
		products.GET("", productHandler.ListProducts)             // 商品列表（含筛选）
		products.GET("/:id", productHandler.GetProductByID)       // 商品详情
		products.POST("", productHandler.CreateProduct)            // 创建商品
		products.PUT("/:id", productHandler.UpdateProduct)         // 更新商品
		products.DELETE("/:id", productHandler.DeleteProduct)      // 删除商品
		products.GET("/search", productHandler.SearchProducts)     // 商品搜索
		products.GET("/recently-viewed", productHandler.GetRecentlyViewed) // 最近浏览商品
		products.POST("/compare", productHandler.CompareProducts)  // 商品对比

		// 商品评论子路由
		products.GET("/:id/reviews", reviewHandler.ListProductReviews)   // 商品评论列表
		products.POST("/:id/reviews", reviewHandler.CreateProductReview) // 提交商品评论
	}

	// -----------------------------------------------------------------------
	// 分类路由组：/categories
	// -----------------------------------------------------------------------
	categories := r.Group("/categories")
	{
		categories.GET("/:id", categoryHandler.GetCategoryByID)    // 分类详情
		categories.GET("", categoryHandler.ListCategories)         // 分类列表
		categories.POST("", categoryHandler.CreateCategory)        // 创建分类
	}

	// -----------------------------------------------------------------------
	// 制造商路由组：/manufacturers
	// -----------------------------------------------------------------------
	manufacturers := r.Group("/manufacturers")
	{
		manufacturers.GET("/:id", manufacturerHandler.GetManufacturerByID)     // 制造商详情
		manufacturers.GET("", manufacturerHandler.ListManufacturers)            // 制造商列表
		manufacturers.POST("", manufacturerHandler.CreateManufacturer)         // 创建制造商
	}

	// -----------------------------------------------------------------------
	// 国家路由组：/countries
	// -----------------------------------------------------------------------
	countries := r.Group("/countries")
	{
		countries.GET("", countryHandler.List)                      // 国家列表
		countries.POST("", countryHandler.Create)                   // 创建国家
		countries.PUT("/:id", countryHandler.Update)                // 更新国家
		countries.DELETE("/:id", countryHandler.Delete)            // 删除国家

		// 国家下的省/州子路由
		countries.GET("/:country_id/states", stateHandler.ListByCountry) // 国家下的省/州列表
		countries.POST("/:country_id/states", stateHandler.Create)       // 创建省/州
	}

	// -----------------------------------------------------------------------
	// 省/州单独操作路由组：/states
	// -----------------------------------------------------------------------
	states := r.Group("/states")
	{
		states.PUT("/:id", stateHandler.Update)                    // 更新省/州
		states.DELETE("/:id", stateHandler.Delete)                  // 删除省/州
	}

	// -----------------------------------------------------------------------
	// 货币路由组：/currencies
	// -----------------------------------------------------------------------
	currencies := r.Group("/currencies")
	{
		currencies.GET("", currencyHandler.List)                    // 货币列表
		currencies.POST("", currencyHandler.Create)                 // 创建货币
		currencies.PUT("/:id", currencyHandler.Update)              // 更新货币
		currencies.DELETE("/:id", currencyHandler.Delete)          // 删除货币
		currencies.POST("/apply-rates", currencyHandler.ApplyRates) // 应用汇率更新
	}

	// -----------------------------------------------------------------------
	// 媒体路由组：/media
	// -----------------------------------------------------------------------
	media := r.Group("/media")
	{
		media.POST("/upload", mediaHandler.Upload)                 // 异步上传图片
		media.GET("/:id", mediaHandler.GetByID)                    // 获取图片
		media.DELETE("/:id", mediaHandler.Delete)                  // 删除图片
	}

	// -----------------------------------------------------------------------
	// SEO 路由组：/seo
	// -----------------------------------------------------------------------
	seo := r.Group("/seo")
	{
		seo.GET("", seoHandler.List)                                // SEO 列表
		seo.GET("/:id", seoHandler.GetByID)                        // SEO 详情
		seo.POST("", seoHandler.Create)                             // 创建 SEO
		seo.DELETE("/:id", seoHandler.Delete)                       // 删除 SEO
	}
}