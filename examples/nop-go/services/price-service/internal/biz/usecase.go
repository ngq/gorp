// Package biz 价格服务业务逻辑层
package biz

import (
	"context"
	"time"

	"nop-go/services/price-service/internal/data"
	"nop-go/services/price-service/internal/models"
)

type PriceUseCase struct {
	priceRepo  data.ProductPriceRepository
	taxRepo    data.TaxRateRepository
	discountRepo data.DiscountRepository
}

func NewPriceUseCase(
	priceRepo data.ProductPriceRepository,
	taxRepo data.TaxRateRepository,
	discountRepo data.DiscountRepository,
) *PriceUseCase {
	return &PriceUseCase{
		priceRepo:  priceRepo,
		taxRepo:    taxRepo,
		discountRepo: discountRepo,
	}
}

func (uc *PriceUseCase) CalculatePrice(ctx context.Context, req *models.CalculatePriceRequest) (*models.PriceResult, error) {
	result := &models.PriceResult{
		ProductID: req.ProductID,
	}

	price, err := uc.priceRepo.GetByProductID(ctx, req.ProductID)
	if err != nil {
		result.BasePrice = 0
		result.FinalPrice = 0
		return result, nil
	}

	result.BasePrice = price.BasePrice
	result.FinalPrice = price.BasePrice

	if price.SpecialPrice > 0 && price.SpecialStart != nil && price.SpecialEnd != nil {
		now := time.Now()
		if now.After(*price.SpecialStart) && now.Before(*price.SpecialEnd) {
			result.FinalPrice = price.SpecialPrice
		}
	} else if price.SalePrice > 0 {
		result.FinalPrice = price.SalePrice
	}

	if req.CouponCode != "" {
		discount, err := uc.discountRepo.GetByCouponCode(ctx, req.CouponCode)
		if err == nil && discount.IsActive {
			var discountAmount float64
			if discount.DiscountType == "percentage" {
				discountAmount = result.FinalPrice * discount.DiscountAmount / 100
			} else {
				discountAmount = discount.DiscountAmount
			}

			if discount.MaxDiscountAmount > 0 && discountAmount > discount.MaxDiscountAmount {
				discountAmount = discount.MaxDiscountAmount
			}

			result.DiscountAmount = discountAmount
			result.FinalPrice -= discountAmount
		}
	}

	if req.CountryCode != "" {
		taxRate, err := uc.taxRepo.GetByLocation(ctx, req.CountryCode, req.StateCode, "")
		if err == nil {
			result.TaxAmount = result.FinalPrice * taxRate.Rate
		}
	}

	result.Total = result.FinalPrice + result.TaxAmount

	return result, nil
}

func (uc *PriceUseCase) ApplyCoupon(ctx context.Context, req *models.ApplyCouponRequest) (*models.CouponResponse, error) {
	discount, err := uc.discountRepo.GetByCouponCode(ctx, req.CouponCode)
	if err != nil {
		return &models.CouponResponse{
			Code:    req.CouponCode,
			IsValid: false,
			Message: "Invalid coupon code",
		}, nil
	}

	if !discount.IsActive {
		return &models.CouponResponse{
			Code:    req.CouponCode,
			IsValid: false,
			Message: "Coupon is not active",
		}, nil
	}

	if discount.StartDate != nil && time.Now().Before(*discount.StartDate) {
		return &models.CouponResponse{
			Code:    req.CouponCode,
			IsValid: false,
			Message: "Coupon is not yet valid",
		}, nil
	}

	if discount.EndDate != nil && time.Now().After(*discount.EndDate) {
		return &models.CouponResponse{
			Code:    req.CouponCode,
			IsValid: false,
			Message: "Coupon has expired",
		}, nil
	}

	if discount.MinOrderAmount > 0 && req.Subtotal < discount.MinOrderAmount {
		return &models.CouponResponse{
			Code:           req.CouponCode,
			IsValid:        false,
			Message:        "Order amount does not meet minimum requirement",
			MinOrderAmount: discount.MinOrderAmount,
		}, nil
	}

	if discount.UsageLimit > 0 && discount.UsedCount >= discount.UsageLimit {
		return &models.CouponResponse{
			Code:    req.CouponCode,
			IsValid: false,
			Message: "Coupon usage limit reached",
		}, nil
	}

	return &models.CouponResponse{
		Code:           req.CouponCode,
		Name:           discount.Name,
		DiscountType:   discount.DiscountType,
		DiscountAmount: discount.DiscountAmount,
		MinOrderAmount: discount.MinOrderAmount,
		IsValid:        true,
	}, nil
}

type TaxUseCase struct {
	taxRepo data.TaxRateRepository
}

func NewTaxUseCase(taxRepo data.TaxRateRepository) *TaxUseCase {
	return &TaxUseCase{taxRepo: taxRepo}
}

func (uc *TaxUseCase) GetTaxRate(ctx context.Context, countryCode, stateCode, zipCode string) (*models.TaxRate, error) {
	return uc.taxRepo.GetByLocation(ctx, countryCode, stateCode, zipCode)
}

func (uc *TaxUseCase) CreateTaxRate(ctx context.Context, rate *models.TaxRate) error {
	return uc.taxRepo.Create(ctx, rate)
}

func (uc *TaxUseCase) ListTaxRates(ctx context.Context) ([]*models.TaxRate, error) {
	return uc.taxRepo.List(ctx)
}

type DiscountUseCase struct {
	discountRepo data.DiscountRepository
}

func NewDiscountUseCase(discountRepo data.DiscountRepository) *DiscountUseCase {
	return &DiscountUseCase{discountRepo: discountRepo}
}

func (uc *DiscountUseCase) CreateDiscount(ctx context.Context, discount *models.Discount) error {
	return uc.discountRepo.Create(ctx, discount)
}

func (uc *DiscountUseCase) GetDiscount(ctx context.Context, id uint64) (*models.Discount, error) {
	return uc.discountRepo.GetByID(ctx, id)
}

func (uc *DiscountUseCase) ListDiscounts(ctx context.Context) ([]*models.Discount, error) {
	return uc.discountRepo.List(ctx)
}

func (uc *DiscountUseCase) IncrementUsage(ctx context.Context, id uint64) error {
	return uc.discountRepo.IncrementUsage(ctx, id)
}