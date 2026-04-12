// Package data 购物车服务数据访问层
package data

import (
	"context"

	"nop-go/services/cart-service/internal/models"

	"gorm.io/gorm"
)

type ShoppingCartRepository interface {
	Create(ctx context.Context, cart *models.ShoppingCart) error
	GetByID(ctx context.Context, id uint64) (*models.ShoppingCart, error)
	GetByCustomerID(ctx context.Context, customerID uint64) (*models.ShoppingCart, error)
	GetBySessionID(ctx context.Context, sessionID string) (*models.ShoppingCart, error)
	Update(ctx context.Context, cart *models.ShoppingCart) error
	Delete(ctx context.Context, id uint64) error
}

type CartItemRepository interface {
	Create(ctx context.Context, item *models.CartItem) error
	GetByID(ctx context.Context, id uint64) (*models.CartItem, error)
	GetByCartID(ctx context.Context, cartID uint64) ([]*models.CartItem, error)
	GetByCartAndProduct(ctx context.Context, cartID, productID uint64) (*models.CartItem, error)
	Update(ctx context.Context, item *models.CartItem) error
	Delete(ctx context.Context, id uint64) error
}

type WishlistRepository interface {
	Create(ctx context.Context, wishlist *models.Wishlist) error
	GetByCustomerID(ctx context.Context, customerID uint64) (*models.Wishlist, error)
	Update(ctx context.Context, wishlist *models.Wishlist) error
}

type WishlistItemRepository interface {
	Create(ctx context.Context, item *models.WishlistItem) error
	GetByWishlistID(ctx context.Context, wishlistID uint64) ([]*models.WishlistItem, error)
	Delete(ctx context.Context, id uint64) error
}

type shoppingCartRepo struct{ db *gorm.DB }

func NewShoppingCartRepository(db *gorm.DB) ShoppingCartRepository {
	return &shoppingCartRepo{db: db}
}

func (r *shoppingCartRepo) Create(ctx context.Context, c *models.ShoppingCart) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *shoppingCartRepo) GetByID(ctx context.Context, id uint64) (*models.ShoppingCart, error) {
	var c models.ShoppingCart
	err := r.db.WithContext(ctx).Preload("Items").First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *shoppingCartRepo) GetByCustomerID(ctx context.Context, customerID uint64) (*models.ShoppingCart, error) {
	var c models.ShoppingCart
	err := r.db.WithContext(ctx).Preload("Items").Where("customer_id = ?", customerID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *shoppingCartRepo) GetBySessionID(ctx context.Context, sessionID string) (*models.ShoppingCart, error) {
	var c models.ShoppingCart
	err := r.db.WithContext(ctx).Preload("Items").Where("session_id = ?", sessionID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *shoppingCartRepo) Update(ctx context.Context, c *models.ShoppingCart) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *shoppingCartRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.ShoppingCart{}, id).Error
}

type cartItemRepo struct{ db *gorm.DB }

func NewCartItemRepository(db *gorm.DB) CartItemRepository {
	return &cartItemRepo{db: db}
}

func (r *cartItemRepo) Create(ctx context.Context, i *models.CartItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *cartItemRepo) GetByID(ctx context.Context, id uint64) (*models.CartItem, error) {
	var i models.CartItem
	err := r.db.WithContext(ctx).First(&i, id).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *cartItemRepo) GetByCartID(ctx context.Context, cartID uint64) ([]*models.CartItem, error) {
	var items []*models.CartItem
	err := r.db.WithContext(ctx).Where("cart_id = ?", cartID).Find(&items).Error
	return items, err
}

func (r *cartItemRepo) GetByCartAndProduct(ctx context.Context, cartID, productID uint64) (*models.CartItem, error) {
	var i models.CartItem
	err := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&i).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *cartItemRepo) Update(ctx context.Context, i *models.CartItem) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *cartItemRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.CartItem{}, id).Error
}

type wishlistRepo struct{ db *gorm.DB }

func NewWishlistRepository(db *gorm.DB) WishlistRepository {
	return &wishlistRepo{db: db}
}

func (r *wishlistRepo) Create(ctx context.Context, w *models.Wishlist) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *wishlistRepo) GetByCustomerID(ctx context.Context, customerID uint64) (*models.Wishlist, error) {
	var w models.Wishlist
	err := r.db.WithContext(ctx).Preload("Items").Where("customer_id = ?", customerID).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *wishlistRepo) Update(ctx context.Context, w *models.Wishlist) error {
	return r.db.WithContext(ctx).Save(w).Error
}

type wishlistItemRepo struct{ db *gorm.DB }

func NewWishlistItemRepository(db *gorm.DB) WishlistItemRepository {
	return &wishlistItemRepo{db: db}
}

func (r *wishlistItemRepo) Create(ctx context.Context, i *models.WishlistItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *wishlistItemRepo) GetByWishlistID(ctx context.Context, wishlistID uint64) ([]*models.WishlistItem, error) {
	var items []*models.WishlistItem
	err := r.db.WithContext(ctx).Where("wishlist_id = ?", wishlistID).Find(&items).Error
	return items, err
}

func (r *wishlistItemRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.WishlistItem{}, id).Error
}