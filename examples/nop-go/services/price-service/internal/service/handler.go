// Package service 价格服务HTTP层
package service

import (
	"net/http"

	"nop-go/services/price-service/internal/biz"
	"nop-go/services/price-service/internal/models"

	"github.com/gin-gonic/gin"
)

type PriceService struct {
	priceUC    *biz.PriceUseCase
	taxUC      *biz.TaxUseCase
	discountUC *biz.DiscountUseCase
}

func NewPriceService(priceUC *biz.PriceUseCase, taxUC *biz.TaxUseCase, discountUC *biz.DiscountUseCase) *PriceService {
	return &PriceService{priceUC: priceUC, taxUC: taxUC, discountUC: discountUC}
}

func (s *PriceService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/price/calculate", s.CalculatePrice)
		api.POST("/price/coupon/apply", s.ApplyCoupon)

		tax := api.Group("/tax")
		{
			tax.GET("", s.ListTaxRates)
			tax.POST("", s.CreateTaxRate)
		}

		discounts := api.Group("/discounts")
		{
			discounts.GET("", s.ListDiscounts)
			discounts.POST("", s.CreateDiscount)
			discounts.GET("/:id", s.GetDiscount)
		}
	}
}

func (s *PriceService) CalculatePrice(c *gin.Context) {
	var req models.CalculatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.priceUC.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *PriceService) ApplyCoupon(c *gin.Context) {
	var req models.ApplyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.priceUC.ApplyCoupon(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *PriceService) ListTaxRates(c *gin.Context) {
	list, err := s.taxUC.ListTaxRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *PriceService) CreateTaxRate(c *gin.Context) {
	var rate models.TaxRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.taxUC.CreateTaxRate(c.Request.Context(), &rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rate)
}

func (s *PriceService) ListDiscounts(c *gin.Context) {
	list, err := s.discountUC.ListDiscounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *PriceService) CreateDiscount(c *gin.Context) {
	var discount models.Discount
	if err := c.ShouldBindJSON(&discount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.discountUC.CreateDiscount(c.Request.Context(), &discount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, discount)
}

func (s *PriceService) GetDiscount(c *gin.Context) {
	id := c.Param("id")
	discount, err := s.discountUC.GetDiscount(c.Request.Context(), 0)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	_ = id
	c.JSON(http.StatusOK, discount)
}