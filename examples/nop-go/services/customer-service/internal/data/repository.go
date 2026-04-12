// Package data 客户服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/customer-service/internal/models"

	"gorm.io/gorm"
)

// CustomerRepository 客户仓储接口
type CustomerRepository interface {
	// Create 创建客户
	Create(ctx context.Context, customer *models.Customer) error
	// GetByID 根据ID获取客户
	GetByID(ctx context.Context, id uint64) (*models.Customer, error)
	// GetByUsername 根据用户名获取客户
	GetByUsername(ctx context.Context, username string) (*models.Customer, error)
	// GetByEmail 根据邮箱获取客户
	GetByEmail(ctx context.Context, email string) (*models.Customer, error)
	// GetByPhone 根据手机号获取客户
	GetByPhone(ctx context.Context, phone string) (*models.Customer, error)
	// Update 更新客户
	Update(ctx context.Context, customer *models.Customer) error
	// Delete 删除客户
	Delete(ctx context.Context, id uint64) error
	// List 客户列表
	List(ctx context.Context, page, pageSize int) ([]*models.Customer, int64, error)
	// AddRole 添加角色
	AddRole(ctx context.Context, customerID, roleID uint64) error
	// RemoveRole 移除角色
	RemoveRole(ctx context.Context, customerID, roleID uint64) error
}

// AddressRepository 地址仓储接口
type AddressRepository interface {
	// Create 创建地址
	Create(ctx context.Context, address *models.Address) error
	// GetByID 根据ID获取地址
	GetByID(ctx context.Context, id uint64) (*models.Address, error)
	// GetByCustomerID 获取客户所有地址
	GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.Address, error)
	// Update 更新地址
	Update(ctx context.Context, address *models.Address) error
	// Delete 删除地址
	Delete(ctx context.Context, id uint64) error
	// SetDefaultBilling 设置默认账单地址
	SetDefaultBilling(ctx context.Context, customerID, addressID uint64) error
	// SetDefaultShipping 设置默认配送地址
	SetDefaultShipping(ctx context.Context, customerID, addressID uint64) error
}

// customerRepo 客户仓储实现
type customerRepo struct {
	db *gorm.DB
}

// NewCustomerRepository 创建客户仓储
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepo{db: db}
}

func (r *customerRepo) Create(ctx context.Context, customer *models.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *customerRepo) GetByID(ctx context.Context, id uint64) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.WithContext(ctx).Preload("Roles").First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) GetByUsername(ctx context.Context, username string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) GetByPhone(ctx context.Context, phone string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.WithContext(ctx).Preload("Roles").Where("phone = ?", phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepo) Update(ctx context.Context, customer *models.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *customerRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Customer{}, id).Error
}

func (r *customerRepo) List(ctx context.Context, page, pageSize int) ([]*models.Customer, int64, error) {
	var customers []*models.Customer
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Customer{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *customerRepo) AddRole(ctx context.Context, customerID, roleID uint64) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO customer_role_mappings (customer_id, role_id) VALUES (?, ?)",
		customerID, roleID,
	).Error
}

func (r *customerRepo) RemoveRole(ctx context.Context, customerID, roleID uint64) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM customer_role_mappings WHERE customer_id = ? AND role_id = ?",
		customerID, roleID,
	).Error
}

// addressRepo 地址仓储实现
type addressRepo struct {
	db *gorm.DB
}

// NewAddressRepository 创建地址仓储
func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepo{db: db}
}

func (r *addressRepo) Create(ctx context.Context, address *models.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}

func (r *addressRepo) GetByID(ctx context.Context, id uint64) (*models.Address, error) {
	var address models.Address
	err := r.db.WithContext(ctx).First(&address, id).Error
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *addressRepo) GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.Address, error) {
	var addresses []*models.Address
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *addressRepo) Update(ctx context.Context, address *models.Address) error {
	return r.db.WithContext(ctx).Save(address).Error
}

func (r *addressRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Address{}, id).Error
}

func (r *addressRepo) SetDefaultBilling(ctx context.Context, customerID, addressID uint64) error {
	// 先清除所有默认
	if err := r.db.WithContext(ctx).Model(&models.Address{}).
		Where("customer_id = ?", customerID).
		Update("is_default_billing", false).Error; err != nil {
		return err
	}
	// 设置指定地址为默认
	return r.db.WithContext(ctx).Model(&models.Address{}).
		Where("id = ?", addressID).
		Update("is_default_billing", true).Error
}

func (r *addressRepo) SetDefaultShipping(ctx context.Context, customerID, addressID uint64) error {
	// 先清除所有默认
	if err := r.db.WithContext(ctx).Model(&models.Address{}).
		Where("customer_id = ?", customerID).
		Update("is_default_shipping", false).Error; err != nil {
		return err
	}
	// 设置指定地址为默认
	return r.db.WithContext(ctx).Model(&models.Address{}).
		Where("id = ?", addressID).
		Update("is_default_shipping", true).Error
}

// CustomerRoleRepository 角色仓储接口
type CustomerRoleRepository interface {
	GetByID(ctx context.Context, id uint64) (*models.CustomerRole, error)
	GetBySystemName(ctx context.Context, name string) (*models.CustomerRole, error)
	List(ctx context.Context) ([]*models.CustomerRole, error)
}

type customerRoleRepo struct {
	db *gorm.DB
}

func NewCustomerRoleRepository(db *gorm.DB) CustomerRoleRepository {
	return &customerRoleRepo{db: db}
}

func (r *customerRoleRepo) GetByID(ctx context.Context, id uint64) (*models.CustomerRole, error) {
	var role models.CustomerRole
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *customerRoleRepo) GetBySystemName(ctx context.Context, name string) (*models.CustomerRole, error) {
	var role models.CustomerRole
	err := r.db.WithContext(ctx).Where("system_name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *customerRoleRepo) List(ctx context.Context) ([]*models.CustomerRole, error) {
	var roles []*models.CustomerRole
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

// ========== GDPR 仓储 ==========

// GdprConsentRepository GDPR同意项仓储接口
type GdprConsentRepository interface {
	Create(ctx context.Context, consent *models.GdprConsent) error
	GetByID(ctx context.Context, id uint64) (*models.GdprConsent, error)
	List(ctx context.Context) ([]*models.GdprConsent, error)
	ListActive(ctx context.Context) ([]*models.GdprConsent, error)
	Update(ctx context.Context, consent *models.GdprConsent) error
	Delete(ctx context.Context, id uint64) error
}

type gdprConsentRepo struct{ db *gorm.DB }

func NewGdprConsentRepository(db *gorm.DB) GdprConsentRepository {
	return &gdprConsentRepo{db: db}
}

func (r *gdprConsentRepo) Create(ctx context.Context, c *models.GdprConsent) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *gdprConsentRepo) GetByID(ctx context.Context, id uint64) (*models.GdprConsent, error) {
	var c models.GdprConsent
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *gdprConsentRepo) List(ctx context.Context) ([]*models.GdprConsent, error) {
	var list []*models.GdprConsent
	err := r.db.WithContext(ctx).Order("display_order").Find(&list).Error
	return list, err
}

func (r *gdprConsentRepo) ListActive(ctx context.Context) ([]*models.GdprConsent, error) {
	var list []*models.GdprConsent
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("display_order").Find(&list).Error
	return list, err
}

func (r *gdprConsentRepo) Update(ctx context.Context, c *models.GdprConsent) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *gdprConsentRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.GdprConsent{}, id).Error
}

// GdprLogRepository GDPR日志仓储接口
type GdprLogRepository interface {
	Create(ctx context.Context, log *models.GdprLog) error
	GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.GdprLog, error)
}

type gdprLogRepo struct{ db *gorm.DB }

func NewGdprLogRepository(db *gorm.DB) GdprLogRepository {
	return &gdprLogRepo{db: db}
}

func (r *gdprLogRepo) Create(ctx context.Context, l *models.GdprLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *gdprLogRepo) GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.GdprLog, error) {
	var logs []*models.GdprLog
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_on_utc desc").Find(&logs).Error
	return logs, err
}

// GdprRequestRepository GDPR请求仓储接口
type GdprRequestRepository interface {
	Create(ctx context.Context, req *models.GdprRequest) error
	GetByID(ctx context.Context, id uint64) (*models.GdprRequest, error)
	GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.GdprRequest, error)
	List(ctx context.Context, page, pageSize int) ([]*models.GdprRequest, int64, error)
	Update(ctx context.Context, req *models.GdprRequest) error
}

type gdprRequestRepo struct{ db *gorm.DB }

func NewGdprRequestRepository(db *gorm.DB) GdprRequestRepository {
	return &gdprRequestRepo{db: db}
}

func (r *gdprRequestRepo) Create(ctx context.Context, req *models.GdprRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *gdprRequestRepo) GetByID(ctx context.Context, id uint64) (*models.GdprRequest, error) {
	var req models.GdprRequest
	err := r.db.WithContext(ctx).First(&req, id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *gdprRequestRepo) GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.GdprRequest, error) {
	var list []*models.GdprRequest
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_on_utc desc").Find(&list).Error
	return list, err
}

func (r *gdprRequestRepo) List(ctx context.Context, page, pageSize int) ([]*models.GdprRequest, int64, error) {
	var list []*models.GdprRequest
	var total int64
	db := r.db.WithContext(ctx).Model(&models.GdprRequest{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_on_utc desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *gdprRequestRepo) Update(ctx context.Context, req *models.GdprRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

// CustomerConsentRepository 客户同意记录仓储接口
type CustomerConsentRepository interface {
	Create(ctx context.Context, cc *models.CustomerConsent) error
	GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.CustomerConsent, error)
	GetByCustomerAndConsent(ctx context.Context, customerID, consentID uint64) (*models.CustomerConsent, error)
	Update(ctx context.Context, cc *models.CustomerConsent) error
	DeleteByCustomerID(ctx context.Context, customerID uint64) error
}

type customerConsentRepo struct{ db *gorm.DB }

func NewCustomerConsentRepository(db *gorm.DB) CustomerConsentRepository {
	return &customerConsentRepo{db: db}
}

func (r *customerConsentRepo) Create(ctx context.Context, cc *models.CustomerConsent) error {
	return r.db.WithContext(ctx).Create(cc).Error
}

func (r *customerConsentRepo) GetByCustomerID(ctx context.Context, customerID uint64) ([]*models.CustomerConsent, error) {
	var list []*models.CustomerConsent
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Find(&list).Error
	return list, err
}

func (r *customerConsentRepo) GetByCustomerAndConsent(ctx context.Context, customerID, consentID uint64) (*models.CustomerConsent, error) {
	var cc models.CustomerConsent
	err := r.db.WithContext(ctx).Where("customer_id = ? AND consent_id = ?", customerID, consentID).First(&cc).Error
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

func (r *customerConsentRepo) Update(ctx context.Context, cc *models.CustomerConsent) error {
	return r.db.WithContext(ctx).Save(cc).Error
}

func (r *customerConsentRepo) DeleteByCustomerID(ctx context.Context, customerID uint64) error {
	return r.db.WithContext(ctx).Where("customer_id = ?", customerID).Delete(&models.CustomerConsent{}).Error
}