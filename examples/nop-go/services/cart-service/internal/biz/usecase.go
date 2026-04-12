// Package biz 购物车服务业务逻辑层
package biz

import (
	"context"

	"nop-go/services/cart-service/internal/data"
	"nop-go/services/cart-service/internal/models"
)

type CartUseCase struct {
	cartRepo data.ShoppingCartRepository
	itemRepo data.CartItemRepository
}

func NewCartUseCase(cartRepo data.ShoppingCartRepository, itemRepo data.CartItemRepository) *CartUseCase {
	return &CartUseCase{cartRepo: cartRepo, itemRepo: itemRepo}
}

func (uc *CartUseCase) GetOrCreateCart(ctx context.Context, customerID uint64, sessionID string) (*models.ShoppingCart, error) {
	if customerID > 0 {
		cart, err := uc.cartRepo.GetByCustomerID(ctx, customerID)
		if err == nil {
			return cart, nil
		}
	}

	if sessionID != "" {
		cart, err := uc.cartRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			return cart, nil
		}
	}

	cart := &models.ShoppingCart{}
	if customerID > 0 {
		cart.CustomerID = customerID
	}
	if sessionID != "" {
		cart.SessionID = sessionID
	}

	if err := uc.cartRepo.Create(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (uc *CartUseCase) AddToCart(ctx context.Context, cartID, productID uint64, productName, sku string, price float64, quantity int, attributes, imageURL string) error {
	existingItem, err := uc.itemRepo.GetByCartAndProduct(ctx, cartID, productID)
	if err == nil {
		existingItem.Quantity += quantity
		return uc.itemRepo.Update(ctx, existingItem)
	}

	item := &models.CartItem{
		CartID:      cartID,
		ProductID:   productID,
		ProductName: productName,
		SKU:         sku,
		Quantity:    quantity,
		UnitPrice:   price,
		Attributes:  attributes,
		ImageURL:    imageURL,
	}

	return uc.itemRepo.Create(ctx, item)
}

func (uc *CartUseCase) UpdateCartItem(ctx context.Context, itemID, quantity int) error {
	item, err := uc.itemRepo.GetByID(ctx, uint64(itemID))
	if err != nil {
		return err
	}

	if quantity <= 0 {
		return uc.itemRepo.Delete(ctx, uint64(itemID))
	}

	item.Quantity = quantity
	return uc.itemRepo.Update(ctx, item)
}

func (uc *CartUseCase) RemoveFromCart(ctx context.Context, itemID uint64) error {
	return uc.itemRepo.Delete(ctx, itemID)
}

func (uc *CartUseCase) ClearCart(ctx context.Context, cartID uint64) error {
	items, err := uc.itemRepo.GetByCartID(ctx, cartID)
	if err != nil {
		return err
	}

	for _, item := range items {
		uc.itemRepo.Delete(ctx, item.ID)
	}

	return nil
}

func (uc *CartUseCase) ApplyCoupon(ctx context.Context, cartID uint64, couponCode string) error {
	cart, err := uc.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return err
	}

	cart.CouponCode = couponCode
	return uc.cartRepo.Update(ctx, cart)
}

func (uc *CartUseCase) RemoveCoupon(ctx context.Context, cartID uint64) error {
	cart, err := uc.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return err
	}

	cart.CouponCode = ""
	return uc.cartRepo.Update(ctx, cart)
}

type WishlistUseCase struct {
	wishlistRepo data.WishlistRepository
	itemRepo     data.WishlistItemRepository
}

func NewWishlistUseCase(wishlistRepo data.WishlistRepository, itemRepo data.WishlistItemRepository) *WishlistUseCase {
	return &WishlistUseCase{wishlistRepo: wishlistRepo, itemRepo: itemRepo}
}

func (uc *WishlistUseCase) GetOrCreateWishlist(ctx context.Context, customerID uint64) (*models.Wishlist, error) {
	wishlist, err := uc.wishlistRepo.GetByCustomerID(ctx, customerID)
	if err == nil {
		return wishlist, nil
	}

	wishlist = &models.Wishlist{CustomerID: customerID}
	if err := uc.wishlistRepo.Create(ctx, wishlist); err != nil {
		return nil, err
	}

	return wishlist, nil
}

func (uc *WishlistUseCase) AddToWishlist(ctx context.Context, customerID, productID uint64, productName, imageURL string) error {
	wishlist, err := uc.GetOrCreateWishlist(ctx, customerID)
	if err != nil {
		return err
	}

	item := &models.WishlistItem{
		WishlistID:  wishlist.ID,
		ProductID:   productID,
		ProductName: productName,
		ImageURL:    imageURL,
	}

	return uc.itemRepo.Create(ctx, item)
}

func (uc *WishlistUseCase) RemoveFromWishlist(ctx context.Context, itemID uint64) error {
	return uc.itemRepo.Delete(ctx, itemID)
}

func (uc *WishlistUseCase) GetWishlist(ctx context.Context, customerID uint64) (*models.Wishlist, error) {
	return uc.wishlistRepo.GetByCustomerID(ctx, customerID)
}