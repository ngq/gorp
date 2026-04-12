// Package service 购物车服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/serviceauth/token"
	"nop-go/services/cart-service/internal/biz"
	"nop-go/services/cart-service/internal/models"
)

type CartService struct {
	cartUC     *biz.CartUseCase
	wishlistUC *biz.WishlistUseCase
	jwtSvc     contract.JWTService
}

// NewCartService 创建购物车服务。
//
// 中文说明：
// - 使用 framework 级 JWTService，替代项目层 jwtSecret；
// - 中间件改用 framework 提供的 AuthMiddleware。
func NewCartService(cartUC *biz.CartUseCase, wishlistUC *biz.WishlistUseCase, jwtSvc contract.JWTService) *CartService {
	return &CartService{cartUC: cartUC, wishlistUC: wishlistUC, jwtSvc: jwtSvc}
}

func (s *CartService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	// 使用 framework JWT middleware
	customerAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "customer")
	{
		cart := api.Group("/cart")
		{
			cart.GET("", s.GetCart)
			cart.POST("/items", s.AddToCart)
			cart.PUT("/items/:item_id", s.UpdateCartItem)
			cart.DELETE("/items/:item_id", s.RemoveFromCart)
			cart.DELETE("", s.ClearCart)
			cart.POST("/coupon", s.ApplyCoupon)
			cart.DELETE("/coupon", s.RemoveCoupon)
		}

		wishlist := api.Group("/wishlist")
		wishlist.Use(customerAuth)
		{
			wishlist.GET("", s.GetWishlist)
			wishlist.POST("/items", s.AddToWishlist)
			wishlist.DELETE("/items/:item_id", s.RemoveFromWishlist)
		}
	}
}

func (s *CartService) GetCart(c *gin.Context) {
	customerID, _ := jwtmiddleware.SubjectIDFromContext(c)
	sessionID := c.GetHeader("X-Session-ID")

	cart, err := s.cartUC.GetOrCreateCart(c.Request.Context(), uint64(customerID), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToCartResponse(cart))
}

func (s *CartService) AddToCart(c *gin.Context) {
	customerID, _ := jwtmiddleware.SubjectIDFromContext(c)
	sessionID := c.GetHeader("X-Session-ID")

	cart, err := s.cartUC.GetOrCreateCart(c.Request.Context(), uint64(customerID), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.cartUC.AddToCart(c.Request.Context(), cart.ID, req.ProductID, "Product", "SKU", 100.0, req.Quantity, req.Attributes, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, _ = s.cartUC.GetOrCreateCart(c.Request.Context(), uint64(customerID), sessionID)
	c.JSON(http.StatusOK, models.ToCartResponse(cart))
}

func (s *CartService) UpdateCartItem(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.Param("item_id"), 10, 64)
	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.cartUC.UpdateCartItem(c.Request.Context(), int(itemID), req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cart item updated"})
}

func (s *CartService) RemoveFromCart(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err := s.cartUC.RemoveFromCart(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CartService) ClearCart(c *gin.Context) {
	customerID, _ := jwtmiddleware.SubjectIDFromContext(c)
	sessionID := c.GetHeader("X-Session-ID")

	cart, err := s.cartUC.GetOrCreateCart(c.Request.Context(), uint64(customerID), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.cartUC.ClearCart(c.Request.Context(), cart.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CartService) ApplyCoupon(c *gin.Context) {
	customerID, _ := jwtmiddleware.SubjectIDFromContext(c)
	sessionID := c.GetHeader("X-Session-ID")

	cart, err := s.cartUC.GetOrCreateCart(c.Request.Context(), uint64(customerID), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.cartUC.ApplyCoupon(c.Request.Context(), cart.ID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "coupon applied"})
}

func (s *CartService) RemoveCoupon(c *gin.Context) {
	customerID, _ := jwtmiddleware.SubjectIDFromContext(c)
	sessionID := c.GetHeader("X-Session-ID")

	cart, err := s.cartUC.GetOrCreateCart(c.Request.Context(), uint64(customerID), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.cartUC.RemoveCoupon(c.Request.Context(), cart.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "coupon removed"})
}

func (s *CartService) GetWishlist(c *gin.Context) {
	customerID, ok := jwtmiddleware.SubjectIDFromContext(c)
	if !ok || customerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}

	wishlist, err := s.wishlistUC.GetWishlist(c.Request.Context(), uint64(customerID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"items": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, wishlist)
}

func (s *CartService) AddToWishlist(c *gin.Context) {
	customerID, ok := jwtmiddleware.SubjectIDFromContext(c)
	if !ok || customerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}

	var req struct {
		ProductID uint64 `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.wishlistUC.AddToWishlist(c.Request.Context(), uint64(customerID), req.ProductID, "Product", ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "added to wishlist"})
}

func (s *CartService) RemoveFromWishlist(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err := s.wishlistUC.RemoveFromWishlist(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}