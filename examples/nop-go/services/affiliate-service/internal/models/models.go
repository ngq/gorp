// Package models 联盟推广服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// Affiliate 联盟会员
type Affiliate struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:255;not null" json:"name"`                    // 联盟会员名称
	Email            string         `gorm:"size:255;not null;uniqueIndex" json:"email"`      // 邮箱
	URL              string         `gorm:"size:255" json:"url"`                              // 网站URL
	FriendlyName     string         `gorm:"size:255" json:"friendly_name"`                   // 友好名称
	AdminComment     string         `gorm:"type:text" json:"admin_comment"`                  // 管理员备注
	Active           bool           `gorm:"default:false" json:"active"`                     // 是否激活
	Deleted          bool           `gorm:"default:false" json:"deleted"`                    // 是否删除
	CreatedOnUtc     time.Time      `json:"created_on_utc"`
	UpdatedOnUtc     time.Time      `json:"updated_on_utc"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Affiliate) TableName() string {
	return "affiliates"
}

// AffiliateOrder 联盟订单
type AffiliateOrder struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	AffiliateID      uint      `gorm:"not null;index" json:"affiliate_id"`           // 联盟会员ID
	OrderID          uint      `gorm:"not null;index" json:"order_id"`              // 订单ID
	CommissionRate   float64   `gorm:"type:decimal(18,8)" json:"commission_rate"`    // 佣金比例
	CommissionAmount float64   `gorm:"type:decimal(18,8)" json:"commission_amount"` // 佣金金额
	IsPaid           bool      `gorm:"default:false" json:"is_paid"`                // 是否已支付佣金
	CreatedOnUtc     time.Time `json:"created_on_utc"`
	UpdatedOnUtc     time.Time `json:"updated_on_utc"`
}

// TableName 指定表名
func (AffiliateOrder) TableName() string {
	return "affiliate_orders"
}

// AffiliateReferral 联盟推荐记录
type AffiliateReferral struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AffiliateID   uint      `gorm:"not null;index" json:"affiliate_id"`     // 联盟会员ID
	CustomerID    uint      `gorm:"index" json:"customer_id"`              // 客户ID
	SessionID     string    `gorm:"size:100;index" json:"session_id"`      // 会话ID
	ReferrerURL   string    `gorm:"size:500" json:"referrer_url"`          // 来源URL
	IPAddress     string    `gorm:"size:50" json:"ip_address"`             // IP地址
	CreatedOnUtc  time.Time `json:"created_on_utc"`                        // 访问时间
	Converted     bool      `gorm:"default:false" json:"converted"`        // 是否转化
	ConvertedOn   time.Time `json:"converted_on"`                          // 转化时间
}

// TableName 指定表名
func (AffiliateReferral) TableName() string {
	return "affiliate_referrals"
}

// AffiliateCommission 联盟佣金记录
type AffiliateCommission struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	AffiliateID      uint      `gorm:"not null;index" json:"affiliate_id"`       // 联盟会员ID
	OrderID          uint      `gorm:"index" json:"order_id"`                    // 关联订单ID
	Amount           float64   `gorm:"type:decimal(18,8);not null" json:"amount"` // 佣金金额
	Status           string    `gorm:"size:20;not null" json:"status"`           // 状态: pending/paid/cancelled
	Description      string    `gorm:"type:text" json:"description"`             // 描述
	CreatedOnUtc     time.Time `json:"created_on_utc"`
	UpdatedOnUtc     time.Time `json:"updated_on_utc"`
	PaidOnUtc        time.Time `json:"paid_on_utc"`                             // 支付时间
}

// TableName 指定表名
func (AffiliateCommission) TableName() string {
	return "affiliate_commissions"
}

// AffiliatePayout 联盟佣金支付记录
type AffiliatePayout struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	AffiliateID      uint           `gorm:"not null;index" json:"affiliate_id"`       // 联盟会员ID
	Amount           float64        `gorm:"type:decimal(18,8);not null" json:"amount"` // 支付金额
	PaymentMethod    string         `gorm:"size:50;not null" json:"payment_method"`   // 支付方式
	PaymentDetails   string         `gorm:"type:text" json:"payment_details"`         // 支付详情
	Status           string         `gorm:"size:20;not null" json:"status"`           // 状态: pending/completed/cancelled
	CreatedOnUtc     time.Time      `json:"created_on_utc"`
	ProcessedOnUtc   time.Time      `json:"processed_on_utc"`                         // 处理时间
	AdminComment     string         `gorm:"type:text" json:"admin_comment"`          // 管理员备注
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AffiliatePayout) TableName() string {
	return "affiliate_payouts"
}

// AffiliateCreateRequest 联盟会员创建请求
type AffiliateCreateRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	URL          string `json:"url"`
	FriendlyName string `json:"friendly_name"`
	AdminComment string `json:"admin_comment"`
	Active       bool   `json:"active"`
}

// AffiliateUpdateRequest 联盟会员更新请求
type AffiliateUpdateRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	URL          string `json:"url"`
	FriendlyName string `json:"friendly_name"`
	AdminComment string `json:"admin_comment"`
	Active       bool   `json:"active"`
}

// AffiliateOrderCreateRequest 联盟订单创建请求
type AffiliateOrderCreateRequest struct {
	AffiliateID    uint    `json:"affiliate_id" binding:"required"`
	OrderID        uint    `json:"order_id" binding:"required"`
	CommissionRate float64 `json:"commission_rate"`
}

// CommissionCalculateRequest 佣金计算请求
type CommissionCalculateRequest struct {
	AffiliateID uint    `json:"affiliate_id" binding:"required"`
	OrderID     uint    `json:"order_id" binding:"required"`
	OrderAmount float64 `json:"order_amount" binding:"required"`
}

// PayoutCreateRequest 支付创建请求
type PayoutCreateRequest struct {
	AffiliateID    uint    `json:"affiliate_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	PaymentMethod  string  `json:"payment_method" binding:"required"`
	PaymentDetails string  `json:"payment_details"`
	AdminComment   string  `json:"admin_comment"`
}

// AffiliateStats 联盟会员统计
type AffiliateStats struct {
	AffiliateID        uint    `json:"affiliate_id"`
	TotalReferrals     int64   `json:"total_referrals"`     // 总推荐数
	ConvertedReferrals int64   `json:"converted_referrals"` // 转化推荐数
	TotalOrders        int64   `json:"total_orders"`        // 总订单数
	TotalCommission    float64 `json:"total_commission"`    // 总佣金
	PendingCommission  float64 `json:"pending_commission"`  // 待支付佣金
	PaidCommission     float64 `json:"paid_commission"`     // 已支付佣金
	ConversionRate     float64 `json:"conversion_rate"`     // 转化率
}