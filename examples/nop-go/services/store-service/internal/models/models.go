// Package models 店铺服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// Store 店铺
type Store struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	URL            string         `gorm:"size:255" json:"url"`
	SSL            bool           `json:"ssl"`
	Hosts          string         `gorm:"type:text" json:"hosts"`          // 多个域名用逗号分隔
	DefaultLanguageID uint        `json:"default_language_id"`
	DisplayOrder   int            `json:"display_order"`
	CompanyName    string         `gorm:"size:255" json:"company_name"`
	CompanyAddress string         `gorm:"size:500" json:"company_address"`
	CompanyPhone   string         `gorm:"size:50" json:"company_phone"`
	CompanyEmail   string         `gorm:"size:100" json:"company_email"`
	Active         bool           `json:"active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Store) TableName() string {
	return "stores"
}

// Vendor 供应商
type Vendor struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:255;not null" json:"name"`
	Email           string         `gorm:"size:100" json:"email"`
	Description     string         `gorm:"type:text" json:"description"`
	AdminComment    string         `gorm:"type:text" json:"admin_comment"`
	AddressID       uint           `json:"address_id"`
	Active          bool           `json:"active"`
	DisplayOrder    int            `json:"display_order"`
	AllowCustomersToContactVendors bool `json:"allow_customers_to_contact"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Vendor) TableName() string {
	return "vendors"
}

// StoreVendor 店铺-供应商关联
type StoreVendor struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StoreID   uint      `gorm:"not null;index" json:"store_id"`
	VendorID  uint      `gorm:"not null;index" json:"vendor_id"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (StoreVendor) TableName() string {
	return "store_vendors"
}

// VendorNote 供应商备注
type VendorNote struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VendorID  uint      `gorm:"not null;index" json:"vendor_id"`
	Note      string    `gorm:"type:text;not null" json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (VendorNote) TableName() string {
	return "vendor_notes"
}

// StoreCreateRequest 店铺创建请求
type StoreCreateRequest struct {
	Name             string `json:"name" binding:"required"`
	URL              string `json:"url"`
	SSL              bool   `json:"ssl"`
	Hosts            string `json:"hosts"`
	DefaultLanguageID uint  `json:"default_language_id"`
	DisplayOrder     int    `json:"display_order"`
	CompanyName      string `json:"company_name"`
	CompanyAddress   string `json:"company_address"`
	CompanyPhone     string `json:"company_phone"`
	CompanyEmail     string `json:"company_email"`
	Active           bool   `json:"active"`
}

// StoreUpdateRequest 店铺更新请求
type StoreUpdateRequest struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	SSL              *bool  `json:"ssl"`
	Hosts            string `json:"hosts"`
	DefaultLanguageID uint  `json:"default_language_id"`
	DisplayOrder     *int   `json:"display_order"`
	CompanyName      string `json:"company_name"`
	CompanyAddress   string `json:"company_address"`
	CompanyPhone     string `json:"company_phone"`
	CompanyEmail     string `json:"company_email"`
	Active           *bool  `json:"active"`
}

// VendorCreateRequest 供应商创建请求
type VendorCreateRequest struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"email"`
	Description     string `json:"description"`
	AdminComment    string `json:"admin_comment"`
	AddressID       uint   `json:"address_id"`
	Active          bool   `json:"active"`
	DisplayOrder    int    `json:"display_order"`
	AllowCustomersToContactVendors bool `json:"allow_customers_to_contact"`
}

// VendorUpdateRequest 供应商更新请求
type VendorUpdateRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Description     string `json:"description"`
	AdminComment    string `json:"admin_comment"`
	AddressID       uint   `json:"address_id"`
	Active          *bool  `json:"active"`
	DisplayOrder    *int   `json:"display_order"`
	AllowCustomersToContactVendors *bool `json:"allow_customers_to_contact"`
}