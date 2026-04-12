// Package biz 联盟推广服务业务逻辑层
package biz

import (
	"context"
	"fmt"
	"time"

	"nop-go/services/affiliate-service/internal/data"
	"nop-go/services/affiliate-service/internal/models"
)

// AffiliateConfig 联盟配置
type AffiliateConfig struct {
	DefaultCommissionRate float64 // 默认佣金比例
	MinCommissionAmount   float64 // 最小佣金金额
	MinPayoutAmount       float64 // 最小提现金额
	CookieExpirationDays  int     // Cookie 有效期（天）
}

// AffiliateUseCase 联盟用例
type AffiliateUseCase struct {
	affiliateRepo   data.AffiliateRepository
	orderRepo       data.AffiliateOrderRepository
	referralRepo    data.AffiliateReferralRepository
	commissionRepo  data.AffiliateCommissionRepository
	payoutRepo      data.AffiliatePayoutRepository
	config          AffiliateConfig
}

// NewAffiliateUseCase 创建联盟用例
func NewAffiliateUseCase(
	affiliateRepo data.AffiliateRepository,
	orderRepo data.AffiliateOrderRepository,
	referralRepo data.AffiliateReferralRepository,
	commissionRepo data.AffiliateCommissionRepository,
	payoutRepo data.AffiliatePayoutRepository,
	config AffiliateConfig,
) *AffiliateUseCase {
	return &AffiliateUseCase{
		affiliateRepo:   affiliateRepo,
		orderRepo:       orderRepo,
		referralRepo:    referralRepo,
		commissionRepo:  commissionRepo,
		payoutRepo:      payoutRepo,
		config:          config,
	}
}

// ========== 联盟会员管理 ==========

// CreateAffiliate 创建联盟会员
func (uc *AffiliateUseCase) CreateAffiliate(ctx context.Context, req *models.AffiliateCreateRequest) (*models.Affiliate, error) {
	// 检查邮箱是否已存在
	existing, err := uc.affiliateRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, data.ErrAffiliateExists
	}

	now := time.Now().UTC()
	affiliate := &models.Affiliate{
		Name:         req.Name,
		Email:        req.Email,
		URL:          req.URL,
		FriendlyName: req.FriendlyName,
		AdminComment: req.AdminComment,
		Active:       req.Active,
		Deleted:      false,
		CreatedOnUtc: now,
		UpdatedOnUtc: now,
	}

	if err := uc.affiliateRepo.Create(ctx, affiliate); err != nil {
		return nil, err
	}

	return affiliate, nil
}

// GetAffiliate 获取联盟会员
func (uc *AffiliateUseCase) GetAffiliate(ctx context.Context, id uint) (*models.Affiliate, error) {
	return uc.affiliateRepo.GetByID(ctx, id)
}

// GetAffiliateByEmail 通过邮箱获取联盟会员
func (uc *AffiliateUseCase) GetAffiliateByEmail(ctx context.Context, email string) (*models.Affiliate, error) {
	return uc.affiliateRepo.GetByEmail(ctx, email)
}

// ListAffiliates 联盟会员列表
func (uc *AffiliateUseCase) ListAffiliates(ctx context.Context, page, pageSize int) ([]*models.Affiliate, int64, error) {
	return uc.affiliateRepo.List(ctx, page, pageSize)
}

// SearchAffiliates 搜索联盟会员
func (uc *AffiliateUseCase) SearchAffiliates(ctx context.Context, keyword string, page, pageSize int) ([]*models.Affiliate, int64, error) {
	return uc.affiliateRepo.Search(ctx, keyword, page, pageSize)
}

// UpdateAffiliate 更新联盟会员
func (uc *AffiliateUseCase) UpdateAffiliate(ctx context.Context, id uint, req *models.AffiliateUpdateRequest) (*models.Affiliate, error) {
	affiliate, err := uc.affiliateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrAffiliateNotFound
	}

	if req.Name != "" {
		affiliate.Name = req.Name
	}
	if req.Email != "" {
		// 检查新邮箱是否已被其他联盟会员使用
		existing, err := uc.affiliateRepo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil && existing.ID != id {
			return nil, data.ErrAffiliateExists
		}
		affiliate.Email = req.Email
	}
	if req.URL != "" {
		affiliate.URL = req.URL
	}
	if req.FriendlyName != "" {
		affiliate.FriendlyName = req.FriendlyName
	}
	affiliate.AdminComment = req.AdminComment
	affiliate.Active = req.Active
	affiliate.UpdatedOnUtc = time.Now().UTC()

	if err := uc.affiliateRepo.Update(ctx, affiliate); err != nil {
		return nil, err
	}

	return affiliate, nil
}

// DeleteAffiliate 删除联盟会员
func (uc *AffiliateUseCase) DeleteAffiliate(ctx context.Context, id uint) error {
	return uc.affiliateRepo.Delete(ctx, id)
}

// ActivateAffiliate 激活联盟会员
func (uc *AffiliateUseCase) ActivateAffiliate(ctx context.Context, id uint) error {
	affiliate, err := uc.affiliateRepo.GetByID(ctx, id)
	if err != nil {
		return data.ErrAffiliateNotFound
	}

	affiliate.Active = true
	affiliate.UpdatedOnUtc = time.Now().UTC()
	return uc.affiliateRepo.Update(ctx, affiliate)
}

// DeactivateAffiliate 禁用联盟会员
func (uc *AffiliateUseCase) DeactivateAffiliate(ctx context.Context, id uint) error {
	affiliate, err := uc.affiliateRepo.GetByID(ctx, id)
	if err != nil {
		return data.ErrAffiliateNotFound
	}

	affiliate.Active = false
	affiliate.UpdatedOnUtc = time.Now().UTC()
	return uc.affiliateRepo.Update(ctx, affiliate)
}

// ========== 联盟推荐追踪 ==========

// TrackReferral 追踪推荐访问
func (uc *AffiliateUseCase) TrackReferral(ctx context.Context, affiliateID uint, sessionID string, referrerURL, ipAddress string, customerID uint) (*models.AffiliateReferral, error) {
	// 检查联盟会员是否存在且激活
	affiliate, err := uc.affiliateRepo.GetByID(ctx, affiliateID)
	if err != nil {
		return nil, data.ErrAffiliateNotFound
	}
	if !affiliate.Active {
		return nil, fmt.Errorf("affiliate is not active")
	}

	referral := &models.AffiliateReferral{
		AffiliateID:  affiliateID,
		CustomerID:   customerID,
		SessionID:    sessionID,
		ReferrerURL:  referrerURL,
		IPAddress:    ipAddress,
		CreatedOnUtc: time.Now().UTC(),
		Converted:    false,
	}

	if err := uc.referralRepo.Create(ctx, referral); err != nil {
		return nil, err
	}

	return referral, nil
}

// ConvertReferral 转化推荐
func (uc *AffiliateUseCase) ConvertReferral(ctx context.Context, sessionID string) error {
	referral, err := uc.referralRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil // 未找到推荐记录，忽略
	}

	now := time.Now().UTC()
	referral.Converted = true
	referral.ConvertedOn = now

	return uc.referralRepo.Update(ctx, referral)
}

// GetAffiliateReferrals 获取联盟会员的推荐记录
func (uc *AffiliateUseCase) GetAffiliateReferrals(ctx context.Context, affiliateID uint) ([]*models.AffiliateReferral, error) {
	return uc.referralRepo.GetByAffiliateID(ctx, affiliateID)
}

// ========== 联盟订单管理 ==========

// CreateAffiliateOrder 创建联盟订单
func (uc *AffiliateUseCase) CreateAffiliateOrder(ctx context.Context, req *models.AffiliateOrderCreateRequest) (*models.AffiliateOrder, error) {
	// 检查联盟会员是否存在
	affiliate, err := uc.affiliateRepo.GetByID(ctx, req.AffiliateID)
	if err != nil {
		return nil, data.ErrAffiliateNotFound
	}
	if !affiliate.Active {
		return nil, fmt.Errorf("affiliate is not active")
	}

	// 检查订单是否已关联
	existing, err := uc.orderRepo.GetByOrderID(ctx, req.OrderID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("order already associated with an affiliate")
	}

	// 计算佣金比例
	commissionRate := req.CommissionRate
	if commissionRate <= 0 {
		commissionRate = uc.config.DefaultCommissionRate
	}

	now := time.Now().UTC()
	order := &models.AffiliateOrder{
		AffiliateID:    req.AffiliateID,
		OrderID:        req.OrderID,
		CommissionRate: commissionRate,
		IsPaid:         false,
		CreatedOnUtc:   now,
		UpdatedOnUtc:   now,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetAffiliateOrders 获取联盟会员的订单
func (uc *AffiliateUseCase) GetAffiliateOrders(ctx context.Context, affiliateID uint) ([]*models.AffiliateOrder, error) {
	return uc.orderRepo.GetByAffiliateID(ctx, affiliateID)
}

// CalculateCommission 计算佣金
func (uc *AffiliateUseCase) CalculateCommission(ctx context.Context, req *models.CommissionCalculateRequest) (*models.AffiliateCommission, error) {
	// 检查联盟会员是否存在
	affiliate, err := uc.affiliateRepo.GetByID(ctx, req.AffiliateID)
	if err != nil {
		return nil, data.ErrAffiliateNotFound
	}
	if !affiliate.Active {
		return nil, fmt.Errorf("affiliate is not active")
	}

	// 检查联盟订单是否存在
	affOrder, err := uc.orderRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, data.ErrOrderNotFound
	}

	// 计算佣金金额
	commissionAmount := req.OrderAmount * affOrder.CommissionRate / 100
	if commissionAmount < uc.config.MinCommissionAmount {
		return nil, nil // 佣金金额太小，不记录
	}

	now := time.Now().UTC()
	commission := &models.AffiliateCommission{
		AffiliateID:  affiliate.ID,
		OrderID:      req.OrderID,
		Amount:       commissionAmount,
		Status:       "pending",
		Description:  fmt.Sprintf("Commission for order %d", req.OrderID),
		CreatedOnUtc: now,
		UpdatedOnUtc: now,
	}

	if err := uc.commissionRepo.Create(ctx, commission); err != nil {
		return nil, err
	}

	// 标记联盟订单佣金已计算（但不支付）
	affOrder.CommissionAmount = commissionAmount
	affOrder.UpdatedOnUtc = now
	uc.orderRepo.Update(ctx, affOrder)

	return commission, nil
}

// GetAffiliateCommissions 获取联盟会员的佣金记录
func (uc *AffiliateUseCase) GetAffiliateCommissions(ctx context.Context, affiliateID uint) ([]*models.AffiliateCommission, error) {
	return uc.commissionRepo.GetByAffiliateID(ctx, affiliateID)
}

// ========== 佣金支付管理 ==========

// CreatePayout 创建佣金支付
func (uc *AffiliateUseCase) CreatePayout(ctx context.Context, req *models.PayoutCreateRequest) (*models.AffiliatePayout, error) {
	// 检查联盟会员是否存在
	_, err := uc.affiliateRepo.GetByID(ctx, req.AffiliateID)
	if err != nil {
		return nil, data.ErrAffiliateNotFound
	}

	// 检查待支付佣金余额
	pendingBalance, err := uc.commissionRepo.SumPendingByAffiliate(ctx, req.AffiliateID)
	if err != nil {
		return nil, err
	}
	if pendingBalance < req.Amount {
		return nil, data.ErrInsufficientBalance
	}

	// 检查最小提现金额
	if req.Amount < uc.config.MinPayoutAmount {
		return nil, fmt.Errorf("payout amount must be at least %.2f", uc.config.MinPayoutAmount)
	}

	now := time.Now().UTC()
	payout := &models.AffiliatePayout{
		AffiliateID:    req.AffiliateID,
		Amount:         req.Amount,
		PaymentMethod:  req.PaymentMethod,
		PaymentDetails: req.PaymentDetails,
		Status:         "pending",
		CreatedOnUtc:   now,
		AdminComment:   req.AdminComment,
	}

	if err := uc.payoutRepo.Create(ctx, payout); err != nil {
		return nil, err
	}

	return payout, nil
}

// ProcessPayout 处理支付
func (uc *AffiliateUseCase) ProcessPayout(ctx context.Context, payoutID uint) error {
	payout, err := uc.payoutRepo.GetByID(ctx, payoutID)
	if err != nil {
		return data.ErrPayoutNotFound
	}

	now := time.Now().UTC()
	payout.Status = "completed"
	payout.ProcessedOnUtc = now

	if err := uc.payoutRepo.Update(ctx, payout); err != nil {
		return err
	}

	// 更新关联的佣金记录状态
	pendingCommissions, err := uc.commissionRepo.GetPendingByAffiliate(ctx, payout.AffiliateID)
	if err != nil {
		return err
	}

	// 计算需要标记为已支付的佣金总额
	var paidAmount float64
	for _, c := range pendingCommissions {
		if paidAmount >= payout.Amount {
			break
		}
		c.Status = "paid"
		c.PaidOnUtc = now
		c.UpdatedOnUtc = now
		uc.commissionRepo.Update(ctx, c)
		paidAmount += c.Amount
	}

	return nil
}

// GetAffiliatePayouts 获取联盟会员的支付记录
func (uc *AffiliateUseCase) GetAffiliatePayouts(ctx context.Context, affiliateID uint) ([]*models.AffiliatePayout, error) {
	return uc.payoutRepo.GetByAffiliateID(ctx, affiliateID)
}

// GetPayout 获取支付记录
func (uc *AffiliateUseCase) GetPayout(ctx context.Context, payoutID uint) (*models.AffiliatePayout, error) {
	return uc.payoutRepo.GetByID(ctx, payoutID)
}

// ========== 统计信息 ==========

// GetAffiliateStats 获取联盟会员统计信息
func (uc *AffiliateUseCase) GetAffiliateStats(ctx context.Context, affiliateID uint) (*models.AffiliateStats, error) {
	// 验证联盟会员存在
	_, err := uc.affiliateRepo.GetByID(ctx, affiliateID)
	if err != nil {
		return nil, data.ErrAffiliateNotFound
	}

	// 获取推荐统计
	totalReferrals, err := uc.referralRepo.CountByAffiliate(ctx, affiliateID)
	if err != nil {
		return nil, err
	}
	convertedReferrals, err := uc.referralRepo.CountConvertedByAffiliate(ctx, affiliateID)
	if err != nil {
		return nil, err
	}

	// 获取订单统计
	orders, err := uc.orderRepo.GetByAffiliateID(ctx, affiliateID)
	if err != nil {
		return nil, err
	}
	totalOrders := int64(len(orders))

	// 获取佣金统计
	totalCommission, err := uc.commissionRepo.SumTotalByAffiliate(ctx, affiliateID)
	if err != nil {
		return nil, err
	}
	pendingCommission, err := uc.commissionRepo.SumPendingByAffiliate(ctx, affiliateID)
	if err != nil {
		return nil, err
	}
	paidCommission, err := uc.commissionRepo.SumPaidByAffiliate(ctx, affiliateID)
	if err != nil {
		return nil, err
	}

	// 计算转化率
	var conversionRate float64
	if totalReferrals > 0 {
		conversionRate = float64(convertedReferrals) / float64(totalReferrals) * 100
	}

	return &models.AffiliateStats{
		AffiliateID:        affiliateID,
		TotalReferrals:     totalReferrals,
		ConvertedReferrals: convertedReferrals,
		TotalOrders:        totalOrders,
		TotalCommission:    totalCommission,
		PendingCommission:  pendingCommission,
		PaidCommission:     paidCommission,
		ConversionRate:     conversionRate,
	}, nil
}

// GetPendingBalance 获取待支付佣金余额
func (uc *AffiliateUseCase) GetPendingBalance(ctx context.Context, affiliateID uint) (float64, error) {
	return uc.commissionRepo.SumPendingByAffiliate(ctx, affiliateID)
}