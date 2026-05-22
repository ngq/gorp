package http

import (
	"nop-go/services/directory/internal/server/http/handler"
	"nop-go/services/directory/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册 HTTP 路由。
//
// 路由分组：
//   - /api/v1/countries — 国家 CRUD
//   - /api/v1/countries/:id/states — 国家下的省/州 CRUD
//   - /api/v1/states/:id — 省/州单独更新/删除
//   - /api/v1/currencies — 货币 CRUD + 汇率应用
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	countryHandler := handler.NewCountryHandler(services.Directory)
	stateHandler := handler.NewStateHandler(services.Directory)
	currencyHandler := handler.NewCurrencyHandler(services.Directory)

	// 国家路由
	countries := r.Group("/api/v1/countries")
	{
		countries.GET("", countryHandler.List)
		countries.POST("", countryHandler.Create)
		countries.PUT("/:id", countryHandler.Update)
		countries.DELETE("/:id", countryHandler.Delete)
	}

	// 国家下的省/州路由
	countryStates := r.Group("/api/v1/countries/:country_id/states")
	{
		countryStates.GET("", stateHandler.ListByCountry)
		countryStates.POST("", stateHandler.Create)
	}

	// 省/州单独操作路由
	states := r.Group("/api/v1/states")
	{
		states.PUT("/:id", stateHandler.Update)
		states.DELETE("/:id", stateHandler.Delete)
	}

	// 货币路由
	currencies := r.Group("/api/v1/currencies")
	{
		currencies.GET("", currencyHandler.List)
		currencies.POST("", currencyHandler.Create)
		currencies.PUT("/:id", currencyHandler.Update)
		currencies.DELETE("/:id", currencyHandler.Delete)
		currencies.POST("/apply-rates", currencyHandler.ApplyRates)
	}
}
