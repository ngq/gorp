// Package data 数据访问层。
// 包含用户服务的持久化对象（PO）定义和仓储实现。
// PO 结构体同时包含 gorm 和 db(sqlx) tag，支持双 ORM 切换。
package data

import (
	"context"
	"time"

	"nop-go/services/user/internal/biz"

	"gorm.io/gorm"
)

// ============================================================
// 持久化对象（PO）
// ============================================================

// UserPO 用户持久化对象。
// 对应数据库 users 表，同时支持 gorm 和 sqlx 映射。
type UserPO struct {
	ID                    uint       `gorm:"primaryKey;column:id"                db:"id"                     json:"id"`
	Username              string     `gorm:"size:64;uniqueIndex;column:username" db:"username"               json:"username"`
	Email                 string     `gorm:"size:128;column:email"               db:"email"                  json:"email"`
	PasswordHash          string     `gorm:"size:256;column:password_hash"       db:"password_hash"          json:"-"`
	PasswordSalt          string     `gorm:"size:64;column:password_salt"        db:"password_salt"          json:"-"`
	Active                bool       `gorm:"column:active;default:true"          db:"active"                 json:"active"`
	IsSystemAccount       bool       `gorm:"column:is_system_account;default:false" db:"is_system_account"   json:"is_system_account"`
	SystemName            string     `gorm:"size:128;column:system_name"         db:"system_name"            json:"system_name"`
	LastLoginAt           *time.Time `gorm:"column:last_login_at"                db:"last_login_at"          json:"last_login_at"`
	LastActivityAt        *time.Time `gorm:"column:last_activity_at"             db:"last_activity_at"       json:"last_activity_at"`
	AvatarURL             string     `gorm:"size:512;column:avatar_url"          db:"avatar_url"             json:"avatar_url"`
	Phone                 string     `gorm:"size:32;column:phone"                db:"phone"                  json:"phone"`
	PasswordRecoveryToken string     `gorm:"size:128;column:password_recovery_token" db:"password_recovery_token" json:"-"`
	MFASecret             string     `gorm:"size:256;column:mfa_secret"          db:"mfa_secret"             json:"-"`
	MFAEnabled            bool       `gorm:"column:mfa_enabled;default:false"    db:"mfa_enabled"            json:"mfa_enabled"`
	CreatedAt             time.Time  `gorm:"autoCreateTime;column:created_at"    db:"created_at"             json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime;column:updated_at"    db:"updated_at"             json:"updated_at"`
}

// TableName 表名。
func (UserPO) TableName() string {
	return "users"
}

// ToEntity 转换为领域实体。
func (po *UserPO) ToEntity() *biz.User {
	return &biz.User{
		ID:                    po.ID,
		Username:              po.Username,
		Email:                 po.Email,
		PasswordHash:          po.PasswordHash,
		PasswordSalt:          po.PasswordSalt,
		Active:                po.Active,
		IsSystemAccount:       po.IsSystemAccount,
		SystemName:            po.SystemName,
		LastLoginAt:           po.LastLoginAt,
		LastActivityAt:        po.LastActivityAt,
		AvatarURL:             po.AvatarURL,
		Phone:                 po.Phone,
		PasswordRecoveryToken: po.PasswordRecoveryToken,
		MFASecret:             po.MFASecret,
		MFAEnabled:            po.MFAEnabled,
		CreatedAt:             po.CreatedAt,
		UpdatedAt:             po.UpdatedAt,
	}
}

// AddressPO 地址持久化对象。
// 对应数据库 addresses 表，同时支持 gorm 和 sqlx 映射。
type AddressPO struct {
	ID              uint      `gorm:"primaryKey;column:id"                    db:"id"               json:"id"`
	UserID          uint      `gorm:"index;column:user_id;not null"           db:"user_id"          json:"user_id"`
	FirstName       string    `gorm:"size:64;column:first_name"               db:"first_name"       json:"first_name"`
	LastName        string    `gorm:"size:64;column:last_name"                db:"last_name"        json:"last_name"`
	Email           string    `gorm:"size:128;column:email"                   db:"email"            json:"email"`
	Phone           string    `gorm:"size:32;column:phone"                    db:"phone"            json:"phone"`
	Fax             string    `gorm:"size:32;column:fax"                      db:"fax"              json:"fax"`
	Company         string    `gorm:"size:128;column:company"                 db:"company"          json:"company"`
	CountryID       uint      `gorm:"column:country_id"                       db:"country_id"       json:"country_id"`
	StateProvinceID uint      `gorm:"column:state_province_id"                db:"state_province_id" json:"state_province_id"`
	City            string    `gorm:"size:64;column:city"                     db:"city"             json:"city"`
	Address1        string    `gorm:"size:256;column:address1"                db:"address1"         json:"address1"`
	Address2        string    `gorm:"size:256;column:address2"                db:"address2"         json:"address2"`
	ZipPostalCode   string    `gorm:"size:32;column:zip_postal_code"          db:"zip_postal_code"  json:"zip_postal_code"`
	IsDefault       bool      `gorm:"column:is_default;default:false"         db:"is_default"       json:"is_default"`
	CreatedAt       time.Time `gorm:"autoCreateTime;column:created_at"        db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime;column:updated_at"        db:"updated_at"       json:"updated_at"`
}

// TableName 表名。
func (AddressPO) TableName() string {
	return "addresses"
}

// ToEntity 转换为领域实体。
func (po *AddressPO) ToEntity() *biz.Address {
	return &biz.Address{
		ID:              po.ID,
		UserID:          po.UserID,
		FirstName:       po.FirstName,
		LastName:        po.LastName,
		Email:           po.Email,
		Phone:           po.Phone,
		Fax:             po.Fax,
		Company:         po.Company,
		CountryID:       po.CountryID,
		StateProvinceID: po.StateProvinceID,
		City:            po.City,
		Address1:        po.Address1,
		Address2:        po.Address2,
		ZipPostalCode:   po.ZipPostalCode,
		IsDefault:       po.IsDefault,
		CreatedAt:       po.CreatedAt,
		UpdatedAt:       po.UpdatedAt,
	}
}

// ExternalAssociationPO 外部关联持久化对象。
// 对应数据库 external_associations 表，同时支持 gorm 和 sqlx 映射。
type ExternalAssociationPO struct {
	ID          uint      `gorm:"primaryKey;column:id"              db:"id"           json:"id"`
	UserID      uint      `gorm:"index;column:user_id;not null"     db:"user_id"      json:"user_id"`
	Provider    string    `gorm:"size:64;column:provider;not null"  db:"provider"     json:"provider"`
	ProviderUID string    `gorm:"size:128;column:provider_uid"      db:"provider_uid" json:"provider_uid"`
	AccessToken string    `gorm:"size:512;column:access_token"      db:"access_token" json:"-"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at"  db:"created_at"   json:"created_at"`
}

// TableName 表名。
func (ExternalAssociationPO) TableName() string {
	return "external_associations"
}

// DownloadableProductPO 可下载产品持久化对象。
// 对应数据库 downloadable_products 表，同时支持 gorm 和 sqlx 映射。
type DownloadableProductPO struct {
	ID            uint       `gorm:"primaryKey;column:id"              db:"id"            json:"id"`
	UserID        uint       `gorm:"index;column:user_id;not null"     db:"user_id"       json:"user_id"`
	OrderItemID   uint       `gorm:"column:order_item_id"              db:"order_item_id" json:"order_item_id"`
	ProductID     uint       `gorm:"column:product_id"                 db:"product_id"    json:"product_id"`
	DownloadCount int        `gorm:"column:download_count;default:0"   db:"download_count" json:"download_count"`
	MaxDownloads  int        `gorm:"column:max_downloads;default:0"    db:"max_downloads" json:"max_downloads"`
	IsActivated   bool       `gorm:"column:is_activated;default:false" db:"is_activated"  json:"is_activated"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"                 db:"expires_at"    json:"expires_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;column:created_at"  db:"created_at"    json:"created_at"`
}

// TableName 表名。
func (DownloadableProductPO) TableName() string {
	return "downloadable_products"
}

// ToEntity 转换为领域实体。
func (po *DownloadableProductPO) ToEntity() *biz.DownloadableProduct {
	return &biz.DownloadableProduct{
		ID:            po.ID,
		UserID:        po.UserID,
		OrderItemID:   po.OrderItemID,
		ProductID:     po.ProductID,
		DownloadCount: po.DownloadCount,
		MaxDownloads:  po.MaxDownloads,
		IsActivated:   po.IsActivated,
		ExpiresAt:     po.ExpiresAt,
		CreatedAt:     po.CreatedAt,
	}
}

// ============================================================
// 仓储实现
// ============================================================

// userRepo 用户仓储实现。
type userRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储。
func NewUserRepo(db *gorm.DB) biz.UserRepository {
	return &userRepo{db: db}
}

// Create 创建用户。
func (r *userRepo) Create(ctx context.Context, user *biz.User) error {
	po := &UserPO{
		Username:              user.Username,
		Email:                 user.Email,
		PasswordHash:          user.PasswordHash,
		PasswordSalt:          user.PasswordSalt,
		Active:                user.Active,
		IsSystemAccount:       user.IsSystemAccount,
		SystemName:            user.SystemName,
		LastLoginAt:           user.LastLoginAt,
		LastActivityAt:        user.LastActivityAt,
		AvatarURL:             user.AvatarURL,
		Phone:                 user.Phone,
		PasswordRecoveryToken: user.PasswordRecoveryToken,
		MFASecret:             user.MFASecret,
		MFAEnabled:            user.MFAEnabled,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据 ID 获取用户。
func (r *userRepo) GetByID(ctx context.Context, id uint) (*biz.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByUsername 根据用户名获取用户。
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByEmail 根据邮箱获取用户。
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新用户信息。
func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	return r.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", user.ID).Updates(map[string]any{
		"email":    user.Email,
		"phone":    user.Phone,
		"avatar_url": user.AvatarURL,
		"updated_at": user.UpdatedAt,
	}).Error
}

// UpdatePassword 更新用户密码。
func (r *userRepo) UpdatePassword(ctx context.Context, id uint, passwordHash, passwordSalt string) error {
	return r.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", id).Updates(map[string]any{
		"password_hash": passwordHash,
		"password_salt": passwordSalt,
		"updated_at":    time.Now(),
	}).Error
}

// SetPasswordRecoveryToken 设置密码恢复令牌。
func (r *userRepo) SetPasswordRecoveryToken(ctx context.Context, id uint, token string) error {
	return r.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", id).Updates(map[string]any{
		"password_recovery_token": token,
		"updated_at":              time.Now(),
	}).Error
}

// GetByPasswordRecoveryToken 根据密码恢复令牌获取用户。
func (r *userRepo) GetByPasswordRecoveryToken(ctx context.Context, token string) (*biz.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).Where("password_recovery_token = ?", token).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// UpdateLastLogin 更新最后登录时间。
func (r *userRepo) UpdateLastLogin(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", id).Updates(map[string]any{
		"last_login_at":    &now,
		"last_activity_at": &now,
	}).Error
}

// CheckUsernameAvailable 检查用户名是否可用。
func (r *userRepo) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

// UpdateAvatar 更新用户头像。
func (r *userRepo) UpdateAvatar(ctx context.Context, id uint, avatarURL string) error {
	return r.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", id).Updates(map[string]any{
		"avatar_url":  avatarURL,
		"updated_at":  time.Now(),
	}).Error
}

// RemoveAvatar 移除用户头像。
func (r *userRepo) RemoveAvatar(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", id).Updates(map[string]any{
		"avatar_url":  "",
		"updated_at":  time.Now(),
	}).Error
}

// ---------- 地址仓储实现 ----------

// addressRepo 地址仓储实现。
type addressRepo struct {
	db *gorm.DB
}

// NewAddressRepo 创建地址仓储。
func NewAddressRepo(db *gorm.DB) biz.AddressRepository {
	return &addressRepo{db: db}
}

// ListByUserID 获取用户的所有地址。
func (r *addressRepo) ListByUserID(ctx context.Context, userID uint) ([]*biz.Address, error) {
	var pos []AddressPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}

	addresses := make([]*biz.Address, len(pos))
	for i, po := range pos {
		addresses[i] = po.ToEntity()
	}
	return addresses, nil
}

// GetByID 根据 ID 获取地址。
func (r *addressRepo) GetByID(ctx context.Context, id uint) (*biz.Address, error) {
	var po AddressPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Create 创建地址。
func (r *addressRepo) Create(ctx context.Context, address *biz.Address) error {
	po := &AddressPO{
		UserID:          address.UserID,
		FirstName:       address.FirstName,
		LastName:        address.LastName,
		Email:           address.Email,
		Phone:           address.Phone,
		Fax:             address.Fax,
		Company:         address.Company,
		CountryID:       address.CountryID,
		StateProvinceID: address.StateProvinceID,
		City:            address.City,
		Address1:        address.Address1,
		Address2:        address.Address2,
		ZipPostalCode:   address.ZipPostalCode,
		IsDefault:       address.IsDefault,
		CreatedAt:       address.CreatedAt,
		UpdatedAt:       address.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// Update 更新地址。
func (r *addressRepo) Update(ctx context.Context, address *biz.Address) error {
	return r.db.WithContext(ctx).Model(&AddressPO{}).Where("id = ?", address.ID).Updates(map[string]any{
		"first_name":        address.FirstName,
		"last_name":         address.LastName,
		"email":             address.Email,
		"phone":             address.Phone,
		"fax":               address.Fax,
		"company":           address.Company,
		"country_id":        address.CountryID,
		"state_province_id": address.StateProvinceID,
		"city":              address.City,
		"address1":          address.Address1,
		"address2":          address.Address2,
		"zip_postal_code":   address.ZipPostalCode,
		"is_default":        address.IsDefault,
		"updated_at":        address.UpdatedAt,
	}).Error
}

// Delete 删除地址。
func (r *addressRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AddressPO{}, id).Error
}

// ---------- 外部关联仓储实现 ----------

// externalAssociationRepo 外部关联仓储实现。
type externalAssociationRepo struct {
	db *gorm.DB
}

// NewExternalAssociationRepo 创建外部关联仓储。
func NewExternalAssociationRepo(db *gorm.DB) biz.ExternalAssociationRepository {
	return &externalAssociationRepo{db: db}
}

// Remove 移除外部关联。
func (r *externalAssociationRepo) Remove(ctx context.Context, userID uint, provider string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&ExternalAssociationPO{}).Error
}

// ---------- 可下载产品仓储实现 ----------

// downloadableProductRepo 可下载产品仓储实现。
type downloadableProductRepo struct {
	db *gorm.DB
}

// NewDownloadableProductRepo 创建可下载产品仓储。
func NewDownloadableProductRepo(db *gorm.DB) biz.DownloadableProductRepository {
	return &downloadableProductRepo{db: db}
}

// ListByUserID 获取用户的可下载产品列表。
func (r *downloadableProductRepo) ListByUserID(ctx context.Context, userID uint) ([]*biz.DownloadableProduct, error) {
	var pos []DownloadableProductPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}

	products := make([]*biz.DownloadableProduct, len(pos))
	for i, po := range pos {
		products[i] = po.ToEntity()
	}
	return products, nil
}
