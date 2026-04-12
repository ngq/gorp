// Package models 客户服务数据模型
//
// 中文说明：
// - 定义客户相关的数据模型；
// - 对应 nopCommerce 的 Customer、CustomerRole、Address 等实体。
package models

import (
	"time"

	"nop-go/shared/models"
)

// Customer 客户实体
//
// 中文说明：
// - 对应 nopCommerce Customer 表；
// - 支持用户名、邮箱、手机号登录；
// - 包含客户基本信息和认证信息。
type Customer struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email         string    `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Phone         string    `gorm:"size:32;index" json:"phone"`
	PasswordHash  string    `gorm:"size:256;not null" json:"-"`
	FirstName     string    `gorm:"size:64" json:"first_name"`
	LastName      string    `gorm:"size:64" json:"last_name"`
	Gender        string    `gorm:"size:16;default:'unknown'" json:"gender"`
	Birthday      *time.Time `json:"birthday"`
	AvatarURL     string    `gorm:"size:512" json:"avatar_url"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	PhoneVerified bool      `gorm:"default:false" json:"phone_verified"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastLoginIP   string    `gorm:"size:64" json:"last_login_ip"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Roles    []CustomerRole    `gorm:"many2many:customer_role_mappings;" json:"roles,omitempty"`
	Addresses []Address         `gorm:"foreignKey:CustomerID" json:"addresses,omitempty"`
}

// TableName 表名
func (Customer) TableName() string {
	return "customers"
}

// FullName 返回完整姓名
func (c *Customer) FullName() string {
	if c.FirstName != "" && c.LastName != "" {
		return c.FirstName + " " + c.LastName
	}
	if c.FirstName != "" {
		return c.FirstName
	}
	return c.LastName
}

// CustomerRole 客户角色
//
// 中文说明：
// - 对应 nopCommerce CustomerRole 表；
// - 支持角色继承和权限控制。
type CustomerRole struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	SystemName  string    `gorm:"size:64;uniqueIndex" json:"system_name"`
	Description string    `gorm:"type:text" json:"description"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (CustomerRole) TableName() string {
	return "customer_roles"
}

// 预定义角色
const (
	RoleGuests     = "Guests"
	RoleRegistered = "Registered"
	RoleVendors    = "Vendors"
	RoleAdmins     = "Administrators"
	RoleForumMods  = "ForumModerators"
)

// Address 客户地址
//
// 中文说明：
// - 对应 nopCommerce Address 表；
// - 支持多个地址，区分账单地址和配送地址。
type Address struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	CustomerID       uint64    `gorm:"not null;index" json:"customer_id"`
	FirstName        string    `gorm:"size:64" json:"first_name"`
	LastName         string    `gorm:"size:64" json:"last_name"`
	Email            string    `gorm:"size:128" json:"email"`
	Phone            string    `gorm:"size:32" json:"phone"`
	Company          string    `gorm:"size:128" json:"company"`
	Country          string    `gorm:"size:64" json:"country"`
	CountryCode      string    `gorm:"size:8" json:"country_code"`
	State            string    `gorm:"size:64" json:"state"`
	StateCode        string    `gorm:"size:8" json:"state_code"`
	City             string    `gorm:"size:64" json:"city"`
	Address1         string    `gorm:"size:256;not null" json:"address1"`
	Address2         string    `gorm:"size:256" json:"address2"`
	ZipCode          string    `gorm:"size:16" json:"zip_code"`
	IsDefaultBilling bool      `gorm:"default:false" json:"is_default_billing"`
	IsDefaultShipping bool     `gorm:"default:false" json:"is_default_shipping"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (Address) TableName() string {
	return "addresses"
}

// ToModel 转换为共享模型
func (a *Address) ToModel() *models.Address {
	return &models.Address{
		ID:          a.ID,
		FirstName:   a.FirstName,
		LastName:    a.LastName,
		Email:       a.Email,
		Phone:       a.Phone,
		Company:     a.Company,
		Country:     a.Country,
		CountryCode: a.CountryCode,
		State:       a.State,
		StateCode:   a.StateCode,
		City:        a.City,
		Address1:    a.Address1,
		Address2:    a.Address2,
		ZipCode:     a.ZipCode,
	}
}

// CustomerPassword 客户密码历史
type CustomerPassword struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	CustomerID   uint64    `gorm:"not null;index" json:"customer_id"`
	PasswordHash string    `gorm:"size:256;not null" json:"-"`
	PasswordFormat int     `gorm:"default:1" json:"password_format"` // 0=Clear, 1=Hashed, 2=Encrypted
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (CustomerPassword) TableName() string {
	return "customer_passwords"
}

// ExternalAuthenticationRecord 外部认证记录
//
// 中文说明：
// - 用于第三方登录（微信、QQ、微博等）。
type ExternalAuthenticationRecord struct {
	ID                 uint64    `gorm:"primaryKey" json:"id"`
	CustomerID         uint64    `gorm:"not null;index" json:"customer_id"`
	ExternalIdentifier string    `gorm:"size:256;not null" json:"external_identifier"`
	Provider           string    `gorm:"size:64;not null" json:"provider"` // WeChat, QQ, Weibo, etc.
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (ExternalAuthenticationRecord) TableName() string {
	return "external_authentication_records"
}

// RewardPointsHistory 积分历史
type RewardPointsHistory struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	CustomerID uint64    `gorm:"not null;index" json:"customer_id"`
	Points     int       `gorm:"not null" json:"points"`          // 正数=获得，负数=消费
	Balance    int       `gorm:"not null" json:"balance"`         // 变动后余额
	Message    string    `gorm:"size:256" json:"message"`
	OrderID    *uint64   `json:"order_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (RewardPointsHistory) TableName() string {
	return "reward_points_history"
}

// ========== GDPR 合规 ==========

// GdprConsent GDPR 同意项
type GdprConsent struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	Message         string    `gorm:"type:text;not null" json:"message"`           // 同意说明文本
	IsRequired      bool      `gorm:"default:false" json:"is_required"`            // 是否必须同意
	RequiredMessage string    `gorm:"size:256" json:"required_message"`            // 必须同意时的提示
	DisplayOrder    int       `gorm:"default:0" json:"display_order"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (GdprConsent) TableName() string {
	return "gdpr_consents"
}

// GdprLog GDPR 操作日志
type GdprLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	CustomerID   uint64    `gorm:"not null;index" json:"customer_id"`
	ConsentID    uint64    `gorm:"index" json:"consent_id"`           // 关联的同意项ID
	RequestType  int       `gorm:"not null" json:"request_type"`      // 请求类型：1=同意, 2=撤回, 3=导出, 4=删除
	IpAddress    string    `gorm:"size:50" json:"ip_address"`
	CreatedOnUtc time.Time `json:"created_on_utc"`
}

// TableName 表名
func (GdprLog) TableName() string {
	return "gdpr_logs"
}

// GdprRequest GDPR 数据请求
type GdprRequest struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	CustomerID     uint64    `gorm:"not null;index" json:"customer_id"`
	RequestType    int       `gorm:"not null" json:"request_type"`       // 1=导出个人数据, 2=删除个人数据
	RequestDetails string    `gorm:"type:text" json:"request_details"`   // 请求详情
	Status         int       `gorm:"default:0" json:"status"`            // 0=待处理, 1=已处理, 2=已拒绝
	CreatedOnUtc   time.Time `json:"created_on_utc"`
	ProcessedOnUtc time.Time `json:"processed_on_utc"`                   // 处理时间
}

// TableName 表名
func (GdprRequest) TableName() string {
	return "gdpr_requests"
}

// CustomerConsent 客户同意记录
type CustomerConsent struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	CustomerID  uint64    `gorm:"not null;index" json:"customer_id"`
	ConsentID   uint64    `gorm:"not null;index" json:"consent_id"`
	IsAccepted  bool      `gorm:"default:false" json:"is_accepted"`
	AcceptedAt  time.Time `json:"accepted_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 表名
func (CustomerConsent) TableName() string {
	return "customer_consents"
}

// ========== GDPR DTO ==========

// GdprConsentCreateRequest GDPR同意项创建请求
type GdprConsentCreateRequest struct {
	Message         string `json:"message" binding:"required"`
	IsRequired      bool   `json:"is_required"`
	RequiredMessage string `json:"required_message"`
	DisplayOrder    int    `json:"display_order"`
}

// GdprConsentUpdateRequest GDPR同意项更新请求
type GdprConsentUpdateRequest struct {
	Message         string `json:"message"`
	IsRequired      bool   `json:"is_required"`
	RequiredMessage string `json:"required_message"`
	DisplayOrder    int    `json:"display_order"`
	IsActive        bool   `json:"is_active"`
}

// CustomerConsentRequest 客户同意请求
type CustomerConsentRequest struct {
	CustomerID uint64 `json:"customer_id" binding:"required"`
	ConsentID  uint64 `json:"consent_id" binding:"required"`
	IsAccepted bool   `json:"is_accepted"`
	IpAddress  string `json:"ip_address"`
}

// GdprExportRequest 数据导出请求
type GdprExportRequest struct {
	CustomerID uint64 `json:"customer_id" binding:"required"`
}

// GdprDeleteRequest 数据删除请求
type GdprDeleteRequest struct {
	CustomerID     uint64 `json:"customer_id" binding:"required"`
	RequestDetails string `json:"request_details"`
}

// CustomerDataExport 客户数据导出结果
type CustomerDataExport struct {
	Customer   *Customer             `json:"customer"`
	Addresses  []*Address            `json:"addresses"`
	Consents   []*CustomerConsent    `json:"consents"`
	Logs       []*GdprLog            `json:"logs"`
	ExportedAt time.Time             `json:"exported_at"`
}