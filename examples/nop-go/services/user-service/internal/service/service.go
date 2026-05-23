// Package service 定义用户服务的服务层，作为 handler 和 biz 之间的桥梁。
// 本文件合并了原 user 和 gdpr 服务的服务层，统一管理所有子服务。
// 服务层负责：
// 1. 组装 biz 层用例的返回值为 handler 需要的响应格式
// 2. 处理跨领域编排逻辑
// 3. 隔离 handler 对 biz 层领域实体的直接依赖
package service

import (
	"context"
	"time"

	"nop-go/services/user-service/internal/biz"
	"nop-go/services/user-service/internal/data"

	"gorm.io/gorm"
)

// ======================== 服务容器 ========================

// Services 服务容器，聚合所有子服务实例
// 通过 NewServices 统一初始化，完成所有仓储、用例和服务的依赖注入
type Services struct {
	User *UserService // 用户子服务
	Gdpr *GdprService // GDPR 子服务
}

// NewServices 创建服务容器
// 初始化所有仓储、用例和子服务实例，完成完整依赖注入链路
func NewServices(db *gorm.DB) *Services {
	// ---- 用户子服务初始化 ----
	userRepo := data.NewUserRepo(db)
	addressRepo := data.NewAddressRepo(db)
	extRepo := data.NewExternalAssociationRepo(db)
	dlRepo := data.NewDownloadableProductRepo(db)
	userUC := biz.NewUserUseCase(userRepo, addressRepo, extRepo, dlRepo)

	// ---- GDPR 子服务初始化 ----
	gdprRepo := data.NewGdprRepo(db)
	gdprUC := biz.NewGdprUseCase(gdprRepo)

	return &Services{
		User: &UserService{uc: userUC},
		Gdpr: &GdprService{uc: gdprUC},
	}
}

// ======================== UserService 用户子服务 ========================

// UserService 用户子服务，封装用户相关的业务编排逻辑
type UserService struct {
	uc *biz.UserUseCase
}

// ---- 用户 CRUD ----

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, user *biz.User) (*UserDTO, error) {
	created, err := s.uc.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return toUserDTO(created), nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, user *biz.User) (*UserDTO, error) {
	updated, err := s.uc.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return toUserDTO(updated), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.uc.DeleteUser(ctx, id)
}

// GetUser 根据 ID 获取用户
func (s *UserService) GetUser(ctx context.Context, id uint) (*UserDTO, error) {
	user, err := s.uc.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserDTO(user), nil
}

// ListUsers 获取用户列表（分页）
func (s *UserService) ListUsers(ctx context.Context, page, size int) ([]*UserDTO, int64, error) {
	offset := (page - 1) * size
	users, total, err := s.uc.ListUsers(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*UserDTO, 0, len(users))
	for _, u := range users {
		items = append(items, toUserDTO(u))
	}
	return items, total, nil
}

// ---- 地址管理 ----

// CreateAddress 创建地址
func (s *UserService) CreateAddress(ctx context.Context, address *biz.Address) (*AddressDTO, error) {
	created, err := s.uc.CreateAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return toAddressDTO(created), nil
}

// UpdateAddress 更新地址
func (s *UserService) UpdateAddress(ctx context.Context, address *biz.Address) (*AddressDTO, error) {
	updated, err := s.uc.UpdateAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	return toAddressDTO(updated), nil
}

// DeleteAddress 删除地址
func (s *UserService) DeleteAddress(ctx context.Context, id uint) error {
	return s.uc.DeleteAddress(ctx, id)
}

// ListAddresses 获取用户地址列表
func (s *UserService) ListAddresses(ctx context.Context, userID uint) ([]*AddressDTO, error) {
	addrs, err := s.uc.ListAddresses(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]*AddressDTO, 0, len(addrs))
	for _, a := range addrs {
		items = append(items, toAddressDTO(a))
	}
	return items, nil
}

// ---- 外部关联管理 ----

// CreateExternalAssociation 创建外部关联
func (s *UserService) CreateExternalAssociation(ctx context.Context, ea *biz.ExternalAssociation) (*ExternalAssociationDTO, error) {
	created, err := s.uc.CreateExternalAssociation(ctx, ea)
	if err != nil {
		return nil, err
	}
	return toExternalAssociationDTO(created), nil
}

// DeleteExternalAssociation 删除外部关联
func (s *UserService) DeleteExternalAssociation(ctx context.Context, id uint) error {
	return s.uc.DeleteExternalAssociation(ctx, id)
}

// ListExternalAssociations 获取用户外部关联列表
func (s *UserService) ListExternalAssociations(ctx context.Context, userID uint) ([]*ExternalAssociationDTO, error) {
	eas, err := s.uc.ListExternalAssociations(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]*ExternalAssociationDTO, 0, len(eas))
	for _, e := range eas {
		items = append(items, toExternalAssociationDTO(e))
	}
	return items, nil
}

// ---- 可下载产品管理 ----

// CreateDownloadableProduct 创建可下载产品
func (s *UserService) CreateDownloadableProduct(ctx context.Context, dp *biz.DownloadableProduct) (*DownloadableProductDTO, error) {
	created, err := s.uc.CreateDownloadableProduct(ctx, dp)
	if err != nil {
		return nil, err
	}
	return toDownloadableProductDTO(created), nil
}

// DeleteDownloadableProduct 删除可下载产品
func (s *UserService) DeleteDownloadableProduct(ctx context.Context, id uint) error {
	return s.uc.DeleteDownloadableProduct(ctx, id)
}

// ListDownloadableProducts 获取用户可下载产品列表
func (s *UserService) ListDownloadableProducts(ctx context.Context, userID uint) ([]*DownloadableProductDTO, error) {
	dps, err := s.uc.ListDownloadableProducts(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]*DownloadableProductDTO, 0, len(dps))
	for _, d := range dps {
		items = append(items, toDownloadableProductDTO(d))
	}
	return items, nil
}

// ======================== GdprService GDPR 子服务 ========================

// GdprService GDPR 子服务，封装 GDPR 请求相关的业务编排逻辑
type GdprService struct {
	uc *biz.GdprUseCase
}

// CreateGdpr 创建 GDPR 请求
func (s *GdprService) CreateGdpr(ctx context.Context, gdpr *biz.Gdpr) (*GdprDTO, error) {
	created, err := s.uc.CreateGdpr(ctx, gdpr)
	if err != nil {
		return nil, err
	}
	return toGdprDTO(created), nil
}

// UpdateGdpr 更新 GDPR 请求
func (s *GdprService) UpdateGdpr(ctx context.Context, gdpr *biz.Gdpr) (*GdprDTO, error) {
	updated, err := s.uc.UpdateGdpr(ctx, gdpr)
	if err != nil {
		return nil, err
	}
	return toGdprDTO(updated), nil
}

// DeleteGdpr 删除 GDPR 请求
func (s *GdprService) DeleteGdpr(ctx context.Context, id uint) error {
	return s.uc.DeleteGdpr(ctx, id)
}

// GetGdpr 根据 ID 获取 GDPR 请求
func (s *GdprService) GetGdpr(ctx context.Context, id uint) (*GdprDTO, error) {
	gdpr, err := s.uc.GetGdpr(ctx, id)
	if err != nil {
		return nil, err
	}
	return toGdprDTO(gdpr), nil
}

// ListGdprsByUserID 获取用户的 GDPR 请求列表
func (s *GdprService) ListGdprsByUserID(ctx context.Context, userID uint) ([]*GdprDTO, error) {
	gdprs, err := s.uc.ListGdprsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]*GdprDTO, 0, len(gdprs))
	for _, g := range gdprs {
		items = append(items, toGdprDTO(g))
	}
	return items, nil
}

// ListGdprs 获取 GDPR 请求列表（管理端分页查询）
func (s *GdprService) ListGdprs(ctx context.Context, page, size int) ([]*GdprDTO, int64, error) {
	offset := (page - 1) * size
	gdprs, total, err := s.uc.ListGdprs(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*GdprDTO, 0, len(gdprs))
	for _, g := range gdprs {
		items = append(items, toGdprDTO(g))
	}
	return items, total, nil
}

// ReviewGdpr 审核 GDPR 请求
func (s *GdprService) ReviewGdpr(ctx context.Context, id uint, status string, reviewedBy uint) (*GdprDTO, error) {
	updated, err := s.uc.ReviewGdpr(ctx, id, status, reviewedBy)
	if err != nil {
		return nil, err
	}
	return toGdprDTO(updated), nil
}

// CompleteGdpr 完成 GDPR 请求
func (s *GdprService) CompleteGdpr(ctx context.Context, id uint) (*GdprDTO, error) {
	updated, err := s.uc.CompleteGdpr(ctx, id)
	if err != nil {
		return nil, err
	}
	return toGdprDTO(updated), nil
}

// ======================== DTO 定义 ========================

// UserDTO 用户数据传输对象，对外暴露的响应格式（不含密码等敏感字段）
type UserDTO struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Status    int        `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// AddressDTO 地址数据传输对象
type AddressDTO struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	Detail        string `json:"detail"`
	IsDefault     bool   `json:"is_default"`
}

// ExternalAssociationDTO 外部关联数据传输对象
type ExternalAssociationDTO struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Platform     string `json:"platform"`
	ExternalID   string `json:"external_id"`
	ExternalData string `json:"external_data"`
}

// DownloadableProductDTO 可下载产品数据传输对象
type DownloadableProductDTO struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	ProductID   string     `json:"product_id"`
	ProductName string     `json:"product_name"`
	DownloadURL string     `json:"download_url"`
	ExpireAt    *time.Time `json:"expire_at"`
}

// GdprDTO GDPR 请求数据传输对象
type GdprDTO struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	RequestType string     `json:"request_type"`
	Status      string     `json:"status"`
	Reason      string     `json:"reason"`
	ReviewedBy  *uint      `json:"reviewed_by"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ======================== 领域模型 -> DTO 转换 ========================

// toUserDTO 将用户领域模型转换为 DTO（剥离密码等敏感字段）
func toUserDTO(user *biz.User) *UserDTO {
	return &UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// toAddressDTO 将地址领域模型转换为 DTO
func toAddressDTO(addr *biz.Address) *AddressDTO {
	return &AddressDTO{
		ID:            addr.ID,
		UserID:        addr.UserID,
		RecipientName: addr.RecipientName,
		Phone:         addr.Phone,
		Province:      addr.Province,
		City:          addr.City,
		District:      addr.District,
		Detail:        addr.Detail,
		IsDefault:     addr.IsDefault,
	}
}

// toExternalAssociationDTO 将外部关联领域模型转换为 DTO
func toExternalAssociationDTO(ea *biz.ExternalAssociation) *ExternalAssociationDTO {
	return &ExternalAssociationDTO{
		ID:           ea.ID,
		UserID:       ea.UserID,
		Platform:     ea.Platform,
		ExternalID:   ea.ExternalID,
		ExternalData: ea.ExternalData,
	}
}

// toDownloadableProductDTO 将可下载产品领域模型转换为 DTO
func toDownloadableProductDTO(dp *biz.DownloadableProduct) *DownloadableProductDTO {
	return &DownloadableProductDTO{
		ID:          dp.ID,
		UserID:      dp.UserID,
		ProductID:   dp.ProductID,
		ProductName: dp.ProductName,
		DownloadURL: dp.DownloadURL,
		ExpireAt:    dp.ExpireAt,
	}
}

// toGdprDTO 将 GDPR 领域模型转换为 DTO
func toGdprDTO(g *biz.Gdpr) *GdprDTO {
	return &GdprDTO{
		ID:          g.ID,
		UserID:      g.UserID,
		RequestType: g.RequestType,
		Status:      g.Status,
		Reason:      g.Reason,
		ReviewedBy:  g.ReviewedBy,
		ReviewedAt:  g.ReviewedAt,
		CompletedAt: g.CompletedAt,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}