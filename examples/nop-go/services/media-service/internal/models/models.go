// Package models 媒体服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// Picture 图片
type Picture struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	MimeType       string         `gorm:"size:50;not null" json:"mime_type"`
	SeoFilename    string         `gorm:"size:255" json:"seo_filename"`
	AltAttribute   string         `gorm:"size:255" json:"alt_attribute"`
	TitleAttribute string         `gorm:"size:255" json:"title_attribute"`
	IsNew          bool           `json:"is_new"`
	Path           string         `gorm:"size:500;not null" json:"path"`
	Width          int            `json:"width"`
	Height         int            `json:"height"`
	Size           int64          `json:"size"` // 文件大小(字节)
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Picture) TableName() string {
	return "pictures"
}

// ProductPicture 商品图片关联
type ProductPicture struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProductID    uint      `gorm:"not null;index" json:"product_id"`
	PictureID    uint      `gorm:"not null;index" json:"picture_id"`
	DisplayOrder int       `json:"display_order"`
	IsMain       bool      `json:"is_main"` // 是否主图
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProductPicture) TableName() string {
	return "product_pictures"
}

// CategoryPicture 分类图片关联
type CategoryPicture struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CategoryID   uint      `gorm:"not null;index" json:"category_id"`
	PictureID    uint      `gorm:"not null;index" json:"picture_id"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (CategoryPicture) TableName() string {
	return "category_pictures"
}

// ManufacturerPicture 品牌图片关联
type ManufacturerPicture struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ManufacturerID uint      `gorm:"not null;index" json:"manufacturer_id"`
	PictureID      uint      `gorm:"not null;index" json:"picture_id"`
	DisplayOrder   int       `json:"display_order"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名
func (ManufacturerPicture) TableName() string {
	return "manufacturer_pictures"
}

// VendorPicture 供应商图片关联
type VendorPicture struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	VendorID     uint      `gorm:"not null;index" json:"vendor_id"`
	PictureID    uint      `gorm:"not null;index" json:"picture_id"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (VendorPicture) TableName() string {
	return "vendor_pictures"
}

// Document 文档
type Document struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	MimeType    string         `gorm:"size:50;not null" json:"mime_type"`
	Path        string         `gorm:"size:500;not null" json:"path"`
	Size        int64          `json:"size"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// DownloadDownload 文件下载记录
type DownloadDownload struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DownloadID   uint      `gorm:"not null;index" json:"download_id"`
	CustomerID   uint      `gorm:"index" json:"customer_id"`
	OrderID      uint      `json:"order_id"`
	DownloadDate time.Time `json:"download_date"`
	IPAddress    string    `gorm:"size:50" json:"ip_address"`
}

// TableName 指定表名
func (DownloadDownload) TableName() string {
	return "download_downloads"
}

// UploadResult 上传结果
type UploadResult struct {
	ID       uint   `json:"id"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Size     int64  `json:"size"`
}

// PictureUploadRequest 图片上传请求
type PictureUploadRequest struct {
	AltAttribute   string `form:"alt_attribute"`
	TitleAttribute string `form:"title_attribute"`
	SeoFilename    string `form:"seo_filename"`
	EntityType     string `form:"entity_type"`   // product / category / manufacturer / vendor
	EntityID       uint   `form:"entity_id"`
	DisplayOrder   int    `form:"display_order"`
	IsMain         bool   `form:"is_main"`
}

// DocumentUploadRequest 文档上传请求
type DocumentUploadRequest struct {
	Name        string `form:"name" binding:"required"`
	Description string `form:"description"`
}