// Package data 联盟推广服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/affiliate-service/internal/models"

	"gorm.io/gorm"
)

// AffiliateRepository 联盟会员仓储接口
type AffiliateRepository interface {
	Create(ctx context.Context, affiliate *models.Affiliate) error
	GetByID(ctx context.Context, id uint) (*models.Affiliate, error)
	GetByEmail(ctx context.Context, email string) (*models.Affiliate, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Affiliate, int64, error)
	ListActive(ctx context.Context) ([]*models.Affiliate, error)
	Update(ctx context.Context, affiliate *models.Affiliate) error
	Delete(ctx context.Context, id uint) error
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*models.Affiliate, int64, error)
}

type affiliateRepository struct {
	db *gorm.DB
}

func NewAffiliateRepository(db *gorm.DB) AffiliateRepository {
	return &affiliateRepository{db: db}
}

func (r *affiliateRepository) Create(ctx context.Context, affiliate *models.Affiliate) error {
	return r.db.WithContext(ctx).Create(affiliate).Error
}

func (r *affiliateRepository) GetByID(ctx context.Context, id uint) (*models.Affiliate, error) {
	var affiliate models.Affiliate
	err := r.db.WithContext(ctx).First(&affiliate, id).Error
	if err != nil {
		return nil, err
	}
	return &affiliate, nil
}

func (r *affiliateRepository) GetByEmail(ctx context.Context, email string) (*models.Affiliate, error) {
	var affiliate models.Affiliate
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&affiliate).Error
	if err != nil {
		return nil, err
	}
	return &affiliate, nil
}

func (r *affiliateRepository) List(ctx context.Context, page, pageSize int) ([]*models.Affiliate, int64, error) {
	var affiliates []*models.Affiliate
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Affiliate{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_on_utc desc").Offset(offset).Limit(pageSize).Find(&affiliates).Error; err != nil {
		return nil, 0, err
	}

	return affiliates, total, nil
}

func (r *affiliateRepository) ListActive(ctx context.Context) ([]*models.Affiliate, error) {
	var affiliates []*models.Affiliate
	err := r.db.WithContext(ctx).Where("active = ? AND deleted = ?", true, false).Find(&affiliates).Error
	return affiliates, err
}

func (r *affiliateRepository) Update(ctx context.Context, affiliate *models.Affiliate) error {
	return r.db.WithContext(ctx).Save(affiliate).Error
}

func (r *affiliateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Affiliate{}, id).Error
}

func (r *affiliateRepository) Search(ctx context.Context, keyword string, page, pageSize int) ([]*models.Affiliate, int64, error) {
	var affiliates []*models.Affiliate
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Affiliate{})
	if keyword != "" {
		db = db.Where("name LIKE ? OR email LIKE ? OR friendly_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_on_utc desc").Offset(offset).Limit(pageSize).Find(&affiliates).Error; err != nil {
		return nil, 0, err
	}

	return affiliates, total, nil
}

// AffiliateOrderRepository 联盟订单仓储接口
type AffiliateOrderRepository interface {
	Create(ctx context.Context, order *models.AffiliateOrder) error
	GetByID(ctx context.Context, id uint) (*models.AffiliateOrder, error)
	GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliateOrder, error)
	GetByOrderID(ctx context.Context, orderID uint) (*models.AffiliateOrder, error)
	List(ctx context.Context, page, pageSize int) ([]*models.AffiliateOrder, int64, error)
	Update(ctx context.Context, order *models.AffiliateOrder) error
	Delete(ctx context.Context, id uint) error
	GetUnpaidOrders(ctx context.Context, affiliateID uint) ([]*models.AffiliateOrder, error)
}

type affiliateOrderRepository struct {
	db *gorm.DB
}

func NewAffiliateOrderRepository(db *gorm.DB) AffiliateOrderRepository {
	return &affiliateOrderRepository{db: db}
}

func (r *affiliateOrderRepository) Create(ctx context.Context, order *models.AffiliateOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *affiliateOrderRepository) GetByID(ctx context.Context, id uint) (*models.AffiliateOrder, error) {
	var order models.AffiliateOrder
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *affiliateOrderRepository) GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliateOrder, error) {
	var orders []*models.AffiliateOrder
	err := r.db.WithContext(ctx).Where("affiliate_id = ?", affiliateID).Order("created_on_utc desc").Find(&orders).Error
	return orders, err
}

func (r *affiliateOrderRepository) GetByOrderID(ctx context.Context, orderID uint) (*models.AffiliateOrder, error) {
	var order models.AffiliateOrder
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *affiliateOrderRepository) List(ctx context.Context, page, pageSize int) ([]*models.AffiliateOrder, int64, error) {
	var orders []*models.AffiliateOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&models.AffiliateOrder{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_on_utc desc").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *affiliateOrderRepository) Update(ctx context.Context, order *models.AffiliateOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *affiliateOrderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AffiliateOrder{}, id).Error
}

func (r *affiliateOrderRepository) GetUnpaidOrders(ctx context.Context, affiliateID uint) ([]*models.AffiliateOrder, error) {
	var orders []*models.AffiliateOrder
	err := r.db.WithContext(ctx).Where("affiliate_id = ? AND is_paid = ?", affiliateID, false).Find(&orders).Error
	return orders, err
}

// AffiliateReferralRepository 联盟推荐仓储接口
type AffiliateReferralRepository interface {
	Create(ctx context.Context, referral *models.AffiliateReferral) error
	GetByID(ctx context.Context, id uint) (*models.AffiliateReferral, error)
	GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliateReferral, error)
	GetBySessionID(ctx context.Context, sessionID string) (*models.AffiliateReferral, error)
	Update(ctx context.Context, referral *models.AffiliateReferral) error
	Delete(ctx context.Context, id uint) error
	CountByAffiliate(ctx context.Context, affiliateID uint) (int64, error)
	CountConvertedByAffiliate(ctx context.Context, affiliateID uint) (int64, error)
}

type affiliateReferralRepository struct {
	db *gorm.DB
}

func NewAffiliateReferralRepository(db *gorm.DB) AffiliateReferralRepository {
	return &affiliateReferralRepository{db: db}
}

func (r *affiliateReferralRepository) Create(ctx context.Context, referral *models.AffiliateReferral) error {
	return r.db.WithContext(ctx).Create(referral).Error
}

func (r *affiliateReferralRepository) GetByID(ctx context.Context, id uint) (*models.AffiliateReferral, error) {
	var referral models.AffiliateReferral
	err := r.db.WithContext(ctx).First(&referral, id).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

func (r *affiliateReferralRepository) GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliateReferral, error) {
	var referrals []*models.AffiliateReferral
	err := r.db.WithContext(ctx).Where("affiliate_id = ?", affiliateID).Order("created_on_utc desc").Find(&referrals).Error
	return referrals, err
}

func (r *affiliateReferralRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.AffiliateReferral, error) {
	var referral models.AffiliateReferral
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&referral).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

func (r *affiliateReferralRepository) Update(ctx context.Context, referral *models.AffiliateReferral) error {
	return r.db.WithContext(ctx).Save(referral).Error
}

func (r *affiliateReferralRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AffiliateReferral{}, id).Error
}

func (r *affiliateReferralRepository) CountByAffiliate(ctx context.Context, affiliateID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.AffiliateReferral{}).Where("affiliate_id = ?", affiliateID).Count(&count).Error
	return count, err
}

func (r *affiliateReferralRepository) CountConvertedByAffiliate(ctx context.Context, affiliateID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.AffiliateReferral{}).Where("affiliate_id = ? AND converted = ?", affiliateID, true).Count(&count).Error
	return count, err
}

// AffiliateCommissionRepository 联盟佣金仓储接口
type AffiliateCommissionRepository interface {
	Create(ctx context.Context, commission *models.AffiliateCommission) error
	GetByID(ctx context.Context, id uint) (*models.AffiliateCommission, error)
	GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliateCommission, error)
	Update(ctx context.Context, commission *models.AffiliateCommission) error
	Delete(ctx context.Context, id uint) error
	GetPendingByAffiliate(ctx context.Context, affiliateID uint) ([]*models.AffiliateCommission, error)
	SumPendingByAffiliate(ctx context.Context, affiliateID uint) (float64, error)
	SumPaidByAffiliate(ctx context.Context, affiliateID uint) (float64, error)
	SumTotalByAffiliate(ctx context.Context, affiliateID uint) (float64, error)
}

type affiliateCommissionRepository struct {
	db *gorm.DB
}

func NewAffiliateCommissionRepository(db *gorm.DB) AffiliateCommissionRepository {
	return &affiliateCommissionRepository{db: db}
}

func (r *affiliateCommissionRepository) Create(ctx context.Context, commission *models.AffiliateCommission) error {
	return r.db.WithContext(ctx).Create(commission).Error
}

func (r *affiliateCommissionRepository) GetByID(ctx context.Context, id uint) (*models.AffiliateCommission, error) {
	var commission models.AffiliateCommission
	err := r.db.WithContext(ctx).First(&commission, id).Error
	if err != nil {
		return nil, err
	}
	return &commission, nil
}

func (r *affiliateCommissionRepository) GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliateCommission, error) {
	var commissions []*models.AffiliateCommission
	err := r.db.WithContext(ctx).Where("affiliate_id = ?", affiliateID).Order("created_on_utc desc").Find(&commissions).Error
	return commissions, err
}

func (r *affiliateCommissionRepository) Update(ctx context.Context, commission *models.AffiliateCommission) error {
	return r.db.WithContext(ctx).Save(commission).Error
}

func (r *affiliateCommissionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AffiliateCommission{}, id).Error
}

func (r *affiliateCommissionRepository) GetPendingByAffiliate(ctx context.Context, affiliateID uint) ([]*models.AffiliateCommission, error) {
	var commissions []*models.AffiliateCommission
	err := r.db.WithContext(ctx).Where("affiliate_id = ? AND status = ?", affiliateID, "pending").Find(&commissions).Error
	return commissions, err
}

func (r *affiliateCommissionRepository) SumPendingByAffiliate(ctx context.Context, affiliateID uint) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&models.AffiliateCommission{}).
		Where("affiliate_id = ? AND status = ?", affiliateID, "pending").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *affiliateCommissionRepository) SumPaidByAffiliate(ctx context.Context, affiliateID uint) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&models.AffiliateCommission{}).
		Where("affiliate_id = ? AND status = ?", affiliateID, "paid").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *affiliateCommissionRepository) SumTotalByAffiliate(ctx context.Context, affiliateID uint) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&models.AffiliateCommission{}).
		Where("affiliate_id = ?", affiliateID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

// AffiliatePayoutRepository 联盟支付仓储接口
type AffiliatePayoutRepository interface {
	Create(ctx context.Context, payout *models.AffiliatePayout) error
	GetByID(ctx context.Context, id uint) (*models.AffiliatePayout, error)
	GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliatePayout, error)
	List(ctx context.Context, page, pageSize int) ([]*models.AffiliatePayout, int64, error)
	Update(ctx context.Context, payout *models.AffiliatePayout) error
	Delete(ctx context.Context, id uint) error
	SumTotalByAffiliate(ctx context.Context, affiliateID uint) (float64, error)
}

type affiliatePayoutRepository struct {
	db *gorm.DB
}

func NewAffiliatePayoutRepository(db *gorm.DB) AffiliatePayoutRepository {
	return &affiliatePayoutRepository{db: db}
}

func (r *affiliatePayoutRepository) Create(ctx context.Context, payout *models.AffiliatePayout) error {
	return r.db.WithContext(ctx).Create(payout).Error
}

func (r *affiliatePayoutRepository) GetByID(ctx context.Context, id uint) (*models.AffiliatePayout, error) {
	var payout models.AffiliatePayout
	err := r.db.WithContext(ctx).First(&payout, id).Error
	if err != nil {
		return nil, err
	}
	return &payout, nil
}

func (r *affiliatePayoutRepository) GetByAffiliateID(ctx context.Context, affiliateID uint) ([]*models.AffiliatePayout, error) {
	var payouts []*models.AffiliatePayout
	err := r.db.WithContext(ctx).Where("affiliate_id = ?", affiliateID).Order("created_on_utc desc").Find(&payouts).Error
	return payouts, err
}

func (r *affiliatePayoutRepository) List(ctx context.Context, page, pageSize int) ([]*models.AffiliatePayout, int64, error) {
	var payouts []*models.AffiliatePayout
	var total int64

	db := r.db.WithContext(ctx).Model(&models.AffiliatePayout{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_on_utc desc").Offset(offset).Limit(pageSize).Find(&payouts).Error; err != nil {
		return nil, 0, err
	}

	return payouts, total, nil
}

func (r *affiliatePayoutRepository) Update(ctx context.Context, payout *models.AffiliatePayout) error {
	return r.db.WithContext(ctx).Save(payout).Error
}

func (r *affiliatePayoutRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AffiliatePayout{}, id).Error
}

func (r *affiliatePayoutRepository) SumTotalByAffiliate(ctx context.Context, affiliateID uint) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).Model(&models.AffiliatePayout{}).
		Where("affiliate_id = ? AND status = ?", affiliateID, "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

// 常见错误
var (
	ErrAffiliateNotFound   = errors.New("affiliate not found")
	ErrAffiliateExists     = errors.New("affiliate email already exists")
	ErrOrderNotFound       = errors.New("affiliate order not found")
	ErrReferralNotFound    = errors.New("affiliate referral not found")
	ErrCommissionNotFound  = errors.New("affiliate commission not found")
	ErrPayoutNotFound      = errors.New("affiliate payout not found")
	ErrInsufficientBalance = errors.New("insufficient pending commission balance")
)