//go:build wireinject

package main

import (
	"nop-go/services/cart-service/internal/biz"
	"nop-go/services/cart-service/internal/data"
	"nop-go/services/cart-service/internal/service"

	"github.com/google/wire"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"gorm.io/gorm"
)

func wireCartService(db *gorm.DB, jwtSvc securitycontract.JWTService) (*service.CartService, error) {
	panic(wire.Build(
		data.NewShoppingCartRepository,
		data.NewCartItemRepository,
		data.NewWishlistRepository,
		data.NewWishlistItemRepository,
		biz.NewCartUseCase,
		biz.NewWishlistUseCase,
		service.NewCartService,
	))
}
