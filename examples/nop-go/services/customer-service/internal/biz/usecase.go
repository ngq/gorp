// Package biz 客户服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"nop-go/services/customer-service/internal/data"
	"nop-go/services/customer-service/internal/models"
	shareErrors "nop-go/shared/errors"

	"golang.org/x/crypto/bcrypt"
)

// CustomerUseCase 客户用例
type CustomerUseCase struct {
	customerRepo data.CustomerRepository
	addressRepo  data.AddressRepository
	roleRepo     data.CustomerRoleRepository
	jwtSvc       contract.JWTService
}

// NewCustomerUseCase 创建客户用例。
//
// 中文说明：
// - 使用 framework 级 JWTService，替代项目层 jwtSecret/jwtExpire；
// - JWTService 统一处理签发/验证，配置从 auth.jwt.* 读取。
func NewCustomerUseCase(
	customerRepo data.CustomerRepository,
	addressRepo data.AddressRepository,
	roleRepo data.CustomerRoleRepository,
	jwtSvc contract.JWTService,
) *CustomerUseCase {
	return &CustomerUseCase{
		customerRepo: customerRepo,
		addressRepo:  addressRepo,
		roleRepo:     roleRepo,
		jwtSvc:       jwtSvc,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// Register 注册客户
func (uc *CustomerUseCase) Register(ctx context.Context, req *RegisterRequest) (*models.Customer, error) {
	// 检查用户名是否已存在
	if existing, _ := uc.customerRepo.GetByUsername(ctx, req.Username); existing != nil {
		return nil, shareErrors.ErrCustomerAlreadyExists
	}

	// 检查邮箱是否已存在
	if existing, _ := uc.customerRepo.GetByEmail(ctx, req.Email); existing != nil {
		return nil, shareErrors.ErrCustomerAlreadyExists
	}

	// 密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	customer := &models.Customer{
		Username:      req.Username,
		Email:         req.Email,
		Phone:         req.Phone,
		PasswordHash:  string(passwordHash),
		Gender:        "unknown",
		IsActive:      true,
		EmailVerified: false,
		PhoneVerified: false,
	}

	if err := uc.customerRepo.Create(ctx, customer); err != nil {
		return nil, err
	}

	// 添加默认角色
	if role, _ := uc.roleRepo.GetBySystemName(ctx, models.RoleRegistered); role != nil {
		uc.customerRepo.AddRole(ctx, customer.ID, role.ID)
	}

	return customer, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResult 登录结果
type LoginResult struct {
	Customer *models.Customer
	Token    string
}

// Login 登录
func (uc *CustomerUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResult, error) {
	// 尝试通过用户名查找
	customer, err := uc.customerRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// 尝试通过邮箱查找
		customer, err = uc.customerRepo.GetByEmail(ctx, req.Username)
		if err != nil {
			return nil, shareErrors.ErrInvalidCredentials
		}
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password)); err != nil {
		return nil, shareErrors.ErrInvalidCredentials
	}

	// 检查账户状态
	if !customer.IsActive {
		return nil, shareErrors.ErrCustomerDisabled
	}

	// 更新最后登录信息
	now := time.Now()
	customer.LastLoginAt = &now
	uc.customerRepo.Update(ctx, customer)

	roles := make([]string, 0, len(customer.Roles))
	for _, role := range customer.Roles {
		roles = append(roles, role.SystemName)
	}

	// 使用 framework JWTService 签发 token
	claims := uc.jwtSvc.NewClaims(int64(customer.ID), "customer", customer.Username, roles, 86400)
	token, err := uc.jwtSvc.Sign(claims)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Customer: customer,
		Token:    token,
	}, nil
}

// GetByID 根据ID获取客户
func (uc *CustomerUseCase) GetByID(ctx context.Context, id uint64) (*models.Customer, error) {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrCustomerNotFound
	}
	return customer, nil
}

// GetByEmail 根据邮箱获取客户
func (uc *CustomerUseCase) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	customer, err := uc.customerRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, shareErrors.ErrCustomerNotFound
	}
	return customer, nil
}

// UpdateProfile 更新客户资料
func (uc *CustomerUseCase) UpdateProfile(ctx context.Context, id uint64, firstName, lastName, gender string, birthday *time.Time) (*models.Customer, error) {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrCustomerNotFound
	}

	if firstName != "" {
		customer.FirstName = firstName
	}
	if lastName != "" {
		customer.LastName = lastName
	}
	if gender != "" {
		customer.Gender = gender
	}
	if birthday != nil {
		customer.Birthday = birthday
	}

	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// ChangePassword 修改密码
func (uc *CustomerUseCase) ChangePassword(ctx context.Context, id uint64, oldPassword, newPassword string) error {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrCustomerNotFound
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(oldPassword)); err != nil {
		return shareErrors.ErrPasswordMismatch
	}

	// 哈希新密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	customer.PasswordHash = string(passwordHash)
	return uc.customerRepo.Update(ctx, customer)
}

// List 客户列表
func (uc *CustomerUseCase) List(ctx context.Context, page, pageSize int) ([]*models.Customer, int64, error) {
	return uc.customerRepo.List(ctx, page, pageSize)
}

// ValidateCustomer 验证客户（供其他服务调用）
func (uc *CustomerUseCase) ValidateCustomer(ctx context.Context, id uint64) error {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrCustomerNotFound
	}

	if !customer.IsActive {
		return shareErrors.ErrCustomerDisabled
	}

	return nil
}

// AddressUseCase 地址用例
type AddressUseCase struct {
	addressRepo data.AddressRepository
}

// NewAddressUseCase 创建地址用例
func NewAddressUseCase(addressRepo data.AddressRepository) *AddressUseCase {
	return &AddressUseCase{addressRepo: addressRepo}
}

// CreateAddress 创建地址
func (uc *AddressUseCase) CreateAddress(ctx context.Context, address *models.Address) error {
	return uc.addressRepo.Create(ctx, address)
}

// GetAddressByID 获取地址
func (uc *AddressUseCase) GetAddressByID(ctx context.Context, id uint64) (*models.Address, error) {
	address, err := uc.addressRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrAddressNotFound
	}
	return address, nil
}

// GetCustomerAddresses 获取客户所有地址
func (uc *AddressUseCase) GetCustomerAddresses(ctx context.Context, customerID uint64) ([]*models.Address, error) {
	return uc.addressRepo.GetByCustomerID(ctx, customerID)
}

// UpdateAddress 更新地址
func (uc *AddressUseCase) UpdateAddress(ctx context.Context, address *models.Address) error {
	_, err := uc.addressRepo.GetByID(ctx, address.ID)
	if err != nil {
		return shareErrors.ErrAddressNotFound
	}
	return uc.addressRepo.Update(ctx, address)
}

// DeleteAddress 删除地址
func (uc *AddressUseCase) DeleteAddress(ctx context.Context, id uint64) error {
	return uc.addressRepo.Delete(ctx, id)
}

// SetDefaultBilling 设置默认账单地址
func (uc *AddressUseCase) SetDefaultBilling(ctx context.Context, customerID, addressID uint64) error {
	return uc.addressRepo.SetDefaultBilling(ctx, customerID, addressID)
}

// SetDefaultShipping 设置默认配送地址
func (uc *AddressUseCase) SetDefaultShipping(ctx context.Context, customerID, addressID uint64) error {
	return uc.addressRepo.SetDefaultShipping(ctx, customerID, addressID)
}

// IsErrCustomerNotFound 判断是否为客户不存在错误
func IsErrCustomerNotFound(err error) bool {
	return errors.Is(err, shareErrors.ErrCustomerNotFound)
}

// ========== GDPR 用例 ==========

// GdprUseCase GDPR用例
type GdprUseCase struct {
	consentRepo     data.GdprConsentRepository
	logRepo         data.GdprLogRepository
	requestRepo     data.GdprRequestRepository
	customerConsent data.CustomerConsentRepository
	customerRepo    data.CustomerRepository
	addressRepo     data.AddressRepository
}

// NewGdprUseCase 创建GDPR用例
func NewGdprUseCase(
	consentRepo data.GdprConsentRepository,
	logRepo data.GdprLogRepository,
	requestRepo data.GdprRequestRepository,
	customerConsent data.CustomerConsentRepository,
	customerRepo data.CustomerRepository,
	addressRepo data.AddressRepository,
) *GdprUseCase {
	return &GdprUseCase{
		consentRepo:     consentRepo,
		logRepo:         logRepo,
		requestRepo:     requestRepo,
		customerConsent: customerConsent,
		customerRepo:    customerRepo,
		addressRepo:     addressRepo,
	}
}

// CreateConsent 创建GDPR同意项
func (uc *GdprUseCase) CreateConsent(ctx context.Context, req *models.GdprConsentCreateRequest) (*models.GdprConsent, error) {
	consent := &models.GdprConsent{
		Message:         req.Message,
		IsRequired:      req.IsRequired,
		RequiredMessage: req.RequiredMessage,
		DisplayOrder:    req.DisplayOrder,
		IsActive:        true,
	}
	if err := uc.consentRepo.Create(ctx, consent); err != nil {
		return nil, err
	}
	return consent, nil
}

// GetConsent 获取GDPR同意项
func (uc *GdprUseCase) GetConsent(ctx context.Context, id uint64) (*models.GdprConsent, error) {
	return uc.consentRepo.GetByID(ctx, id)
}

// ListConsents GDPR同意项列表
func (uc *GdprUseCase) ListConsents(ctx context.Context) ([]*models.GdprConsent, error) {
	return uc.consentRepo.List(ctx)
}

// ListActiveConsents 获取活动的GDPR同意项
func (uc *GdprUseCase) ListActiveConsents(ctx context.Context) ([]*models.GdprConsent, error) {
	return uc.consentRepo.ListActive(ctx)
}

// UpdateConsent 更新GDPR同意项
func (uc *GdprUseCase) UpdateConsent(ctx context.Context, id uint64, req *models.GdprConsentUpdateRequest) (*models.GdprConsent, error) {
	consent, err := uc.consentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Message != "" {
		consent.Message = req.Message
	}
	consent.IsRequired = req.IsRequired
	consent.RequiredMessage = req.RequiredMessage
	consent.DisplayOrder = req.DisplayOrder
	consent.IsActive = req.IsActive
	if err := uc.consentRepo.Update(ctx, consent); err != nil {
		return nil, err
	}
	return consent, nil
}

// DeleteConsent 删除GDPR同意项
func (uc *GdprUseCase) DeleteConsent(ctx context.Context, id uint64) error {
	return uc.consentRepo.Delete(ctx, id)
}

// AcceptConsent 客户接受同意项
func (uc *GdprUseCase) AcceptConsent(ctx context.Context, req *models.CustomerConsentRequest) error {
	// 检查同意项是否存在
	consent, err := uc.consentRepo.GetByID(ctx, req.ConsentID)
	if err != nil {
		return err
	}

	// 检查是否已有记录
	existing, err := uc.customerConsent.GetByCustomerAndConsent(ctx, req.CustomerID, req.ConsentID)
	if err == nil && existing != nil {
		// 更新现有记录
		existing.IsAccepted = req.IsAccepted
		if req.IsAccepted {
			existing.AcceptedAt = time.Now()
		}
		return uc.customerConsent.Update(ctx, existing)
	}

	// 创建新记录
	cc := &models.CustomerConsent{
		CustomerID: req.CustomerID,
		ConsentID:  req.ConsentID,
		IsAccepted: req.IsAccepted,
	}
	if req.IsAccepted {
		cc.AcceptedAt = time.Now()
	}
	if err := uc.customerConsent.Create(ctx, cc); err != nil {
		return err
	}

	// 记录日志
	logType := 1 // 同意
	if !req.IsAccepted {
		logType = 2 // 撤回
	}
	log := &models.GdprLog{
		CustomerID:  req.CustomerID,
		ConsentID:   req.ConsentID,
		RequestType: logType,
		IpAddress:   req.IpAddress,
		CreatedOnUtc: time.Now().UTC(),
	}
	uc.logRepo.Create(ctx, log)

	_ = consent // 避免未使用警告
	return nil
}

// GetCustomerConsents 获取客户的同意记录
func (uc *GdprUseCase) GetCustomerConsents(ctx context.Context, customerID uint64) ([]*models.CustomerConsent, error) {
	return uc.customerConsent.GetByCustomerID(ctx, customerID)
}

// RequestDataExport 请求数据导出
func (uc *GdprUseCase) RequestDataExport(ctx context.Context, req *models.GdprExportRequest) (*models.GdprRequest, error) {
	gdprReq := &models.GdprRequest{
		CustomerID:   req.CustomerID,
		RequestType:  1, // 导出
		Status:       0, // 待处理
		CreatedOnUtc: time.Now().UTC(),
	}
	if err := uc.requestRepo.Create(ctx, gdprReq); err != nil {
		return nil, err
	}
	return gdprReq, nil
}

// RequestDataDeletion 请求数据删除
func (uc *GdprUseCase) RequestDataDeletion(ctx context.Context, req *models.GdprDeleteRequest) (*models.GdprRequest, error) {
	gdprReq := &models.GdprRequest{
		CustomerID:     req.CustomerID,
		RequestType:    2, // 删除
		RequestDetails: req.RequestDetails,
		Status:         0, // 待处理
		CreatedOnUtc:   time.Now().UTC(),
	}
	if err := uc.requestRepo.Create(ctx, gdprReq); err != nil {
		return nil, err
	}
	return gdprReq, nil
}

// ExportCustomerData 导出客户数据
func (uc *GdprUseCase) ExportCustomerData(ctx context.Context, customerID uint64) (*models.CustomerDataExport, error) {
	// 获取客户信息
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 获取地址
	addresses, err := uc.addressRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 获取同意记录
	consents, err := uc.customerConsent.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 获取日志
	logs, err := uc.logRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return &models.CustomerDataExport{
		Customer:   customer,
		Addresses:  addresses,
		Consents:   consents,
		Logs:       logs,
		ExportedAt: time.Now(),
	}, nil
}

// DeleteCustomerData 删除客户数据（GDPR合规）
func (uc *GdprUseCase) DeleteCustomerData(ctx context.Context, customerID uint64) error {
	// 删除同意记录
	if err := uc.customerConsent.DeleteByCustomerID(ctx, customerID); err != nil {
		return err
	}

	// 删除客户（软删除）
	if err := uc.customerRepo.Delete(ctx, customerID); err != nil {
		return err
	}

	return nil
}

// GetGdprRequests 获取GDPR请求列表
func (uc *GdprUseCase) GetGdprRequests(ctx context.Context, page, pageSize int) ([]*models.GdprRequest, int64, error) {
	return uc.requestRepo.List(ctx, page, pageSize)
}

// ProcessGdprRequest 处理GDPR请求
func (uc *GdprUseCase) ProcessGdprRequest(ctx context.Context, requestID uint64, approve bool) error {
	req, err := uc.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	if approve {
		req.Status = 1 // 已处理
		if req.RequestType == 2 {
			// 如果是删除请求，执行删除
			if err := uc.DeleteCustomerData(ctx, req.CustomerID); err != nil {
				return err
			}
		}
	} else {
		req.Status = 2 // 已拒绝
	}

	req.ProcessedOnUtc = time.Now().UTC()
	return uc.requestRepo.Update(ctx, req)
}

// GetGdprLogs 获取客户的GDPR日志
func (uc *GdprUseCase) GetGdprLogs(ctx context.Context, customerID uint64) ([]*models.GdprLog, error) {
	return uc.logRepo.GetByCustomerID(ctx, customerID)
}