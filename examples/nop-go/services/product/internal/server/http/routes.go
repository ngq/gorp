package http

import (
	"nop-go/services/product/internal/server/http/handler"
	"nop-go/services/product/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册商品服务（catalog）的全部 HTTP 路由。
//
// 路由分组说明：
// - /api/v1/products       — 商品 CRUD + 评论 + 最近浏览 + 对比
// - /api/v1/categories     — 分类 CRUD
// - /api/v1/manufacturers  — 制造商 CRUD
// - /api/v1/search         — 商品搜索 + 自动完成
//
// 注意事项：
// - recently-viewed 和 compare 路由注册在 products 组下，
//   但不使用 :id 参数，需放在 :id 路由之前以避免路径冲突。
// - reviews 路由使用 products/:id/reviews 子路径。
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	// 初始化各处理器
	productHandler := handler.NewProductHandler(services.Product)
	categoryHandler := handler.NewCategoryHandler(services.Category)
	manufacturerHandler := handler.NewManufacturerHandler(services.Manufacturer)
	reviewHandler := handler.NewProductReviewHandler(services.Review)
	catalogHandler := handler.NewCatalogHandler(services.Product)

	// -----------------------------------------------------------------------
	// 商品路由组：/api/v1/products
	// -----------------------------------------------------------------------
	products := r.Group("/api/v1/products")
	{
		// 静态路径（不带 :id 参数）必须注册在动态路径之前，
		// 否则 recently-viewed 和 compare 会被 :id 路径匹配。
		products.GET("/recently-viewed", catalogHandler.GetRecentlyViewed)    // 最近浏览商品
		products.POST("/compare", catalogHandler.CompareProducts)              // 商品对比

		// 商品评论子路由
		products.GET("/:id/reviews", reviewHandler.ListProductReviews)         // 商品评论列表
		products.POST("/:id/reviews", reviewHandler.CreateProductReview)       // 提交商品评论

		// 商品 CRUD
		products.GET("/:id", productHandler.GetProductByID)                    // 商品详情
		products.GET("", productHandler.ListProducts)                          // 商品列表（含筛选）
		products.POST("", productHandler.CreateProduct)                        // 创建商品
		products.PUT("/:id", productHandler.UpdateProduct)                     // 更新商品
		products.DELETE("/:id", productHandler.DeleteProduct)                  // 删除商品
	}

	// -----------------------------------------------------------------------
	// 分类路由组：/api/v1/categories
	// -----------------------------------------------------------------------
	categories := r.Group("/api/v1/categories")
	{
		categories.GET("/:id", categoryHandler.GetCategoryByID)                // 分类详情
		categories.GET("", categoryHandler.ListCategories)                     // 分类列表
		categories.POST("", categoryHandler.CreateCategory)                    // 创建分类
	}

	// -----------------------------------------------------------------------
	// 制造商路由组：/api/v1/manufacturers
	// -----------------------------------------------------------------------
	manufacturers := r.Group("/api/v1/manufacturers")
	{
		manufacturers.GET("/:id", manufacturerHandler.GetManufacturerByID)     // 制造商详情
		manufacturers.GET("", manufacturerHandler.ListManufacturers)           // 制造商列表
		manufacturers.POST("", manufacturerHandler.CreateManufacturer)         // 创建制造商
	}

	// -----------------------------------------------------------------------
	// 搜索路由组：/api/v1/search
	// -----------------------------------------------------------------------
	search := r.Group("/api/v1/search")
	{
		search.GET("", catalogHandler.SearchProducts)                          // 商品搜索
		search.GET("/autocomplete", catalogHandler.SearchAutocomplete)         // 搜索自动完成
	}
}