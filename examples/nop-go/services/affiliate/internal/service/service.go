package service

import (
	"context"

	"nop-go/services/affiliate/internal/biz"
	"nop-go/services/affiliate/internal/data"

	"gorm.io/gorm"
)

// Services 联盟服务集合。
type Services struct {
	Affiliate *AffiliateService
}

// NewServices 创建联盟服务集合。
func NewServices(db *gorm.DB) *Services {
	affRepo := data.NewAffiliateRepo(db)
	orderRepo := data.NewAffiliateOrderRepo(db)
	customerRepo := data.NewAffiliateCustomerRepo(db)
	affUC := biz.NewAffiliateUseCase(affRepo, orderRepo, customerRepo)
	return &Services{
		Affiliate: &AffiliateService{uc: affUC},
	}
}

// AffiliateService 联盟服务。
type AffiliateService struct {
	uc *biz.AffiliateUseCase
}

// CreateAffiliateRequest 创建联盟请求。
type CreateAffiliateRequest struct {
	Name   string `json:"name" binding:"required"`     // 联盟名称
	Url    string `json:"url" binding:"required"`      // 联盟URL
	Active bool   `json:"active"`                      // 是否启用
}

// UpdateAffiliateRequest 更新联盟请求。
type UpdateAffiliateRequest struct {
	Name   string `json:"name" binding:"required"`     // 联盟名称
	Url    string `json:"url" binding:"required"`      // 联盟URL
	Active bool   `json:"active"`                      // 是否启用
}

// AffiliateResponse 联盟响应。
type AffiliateResponse struct {
	ID        uint   `json:"id"`         // 联盟ID
	Name      string `json:"name"`       // 联盟名称
	Url       string `json:"url"`        // 联盟URL
	Active    bool   `json:"active"`     // 是否启用
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// AffiliateOrderResponse 联盟订单响应。
type AffiliateOrderResponse struct {
	ID          uint    `json:"id"`           // 订单ID
	OrderNo     string  `json:"order_no"`     // 订单编号
	CustomerID  uint    `json:"customer_id"`  // 客户ID
	TotalAmount float64 `json:"total_amount"` // 订单总金额
	Status      string  `json:"status"`       // 订单状态
	CreatedAt   string  `json:"created_at"`   // 创建时间
}

// AffiliateCustomerResponse 联盟客户响应。
type AffiliateCustomerResponse struct {
	ID        uint   `json:"id"`         // 客户ID
	Username  string `json:"username"`   // 客户用户名
	Email     string `json:"email"`      // 客户邮箱
	CreatedAt string `json:"created_at"` // 注册时间
}

// toAffiliateResponse 将联盟领域实体转换为响应结构体。
func toAffiliateResponse(aff *biz.Affiliate) *AffiliateResponse {
	return &AffiliateResponse{
		ID:        aff.ID,
		Name:      aff.Name,
		Url:       aff.Url,
		Active:    aff.Active,
		CreatedAt: aff.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: aff.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List 获取联盟列表。
func (s *AffiliateService) List(ctx context.Context, page, size int) ([]AffiliateResponse, int64, error) {
	affs, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]AffiliateResponse, len(affs))
	for i, aff := range affs {
		items[i] = *toAffiliateResponse(aff)
	}
	return items, total, nil
}

// Create 创建联盟。
func (s *AffiliateService) Create(ctx context.Context, req CreateAffiliateRequest) (*AffiliateResponse, error) {
	aff, err := s.uc.Create(ctx, req.Name, req.Url, req.Active)
	if err != nil {
		return nil, err
	}
	return toAffiliateResponse(aff), nil
}

// Update 更新联盟。
func (s *AffiliateService) Update(ctx context.Context, id uint, req UpdateAffiliateRequest) (*AffiliateResponse, error) {
	aff, err := s.uc.Update(ctx, id, req.Name, req.Url, req.Active)
	if err != nil {
		return nil, err
	}
	return toAffiliateResponse(aff), nil
}

// Delete 删除联盟。
func (s *AffiliateService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// ListOrders 获取联盟关联订单。
func (s *AffiliateService) ListOrders(ctx context.Context, affiliateID uint, page, size int) ([]AffiliateOrderResponse, int64, error) {
	orders, total, err := s.uc.ListOrders(ctx, affiliateID, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]AffiliateOrderResponse, len(orders))
	for i, o := range orders {
		items[i] = AffiliateOrderResponse{
			ID:          o.ID,
			OrderNo:     o.OrderNo,
			CustomerID:  o.CustomerID,
			TotalAmount: o.TotalAmount,
			Status:      o.Status,
			CreatedAt:   o.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// ListCustomers 获取联盟关联客户。
func (s *AffiliateService) ListCustomers(ctx context.Context, affiliateID uint, page, size int) ([]AffiliateCustomerResponse, int64, error) {
	customers, total, err := s.uc.ListCustomers(ctx, affiliateID, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]AffiliateCustomerResponse, len(customers))
	for i, c := range customers {
		items[i] = AffiliateCustomerResponse{
			ID:        c.ID,
			Username:  c.Username,
			Email:     c.Email,
			CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}