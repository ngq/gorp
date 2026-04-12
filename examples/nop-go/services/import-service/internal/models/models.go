// Package models 导入导出服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// ImportProfile 导入配置
type ImportProfile struct {
	ID                     uint           `gorm:"primaryKey" json:"id"`
	Name                   string         `gorm:"size:256;not null" json:"name"`
	EntityType             string         `gorm:"size:50;not null;index" json:"entity_type"`           // 实体类型(product/customer/order等)
	FilePath               string         `gorm:"size:512" json:"file_path"`                           // 文件路径
	FileType               string         `gorm:"size:20" json:"file_type"`                            // 文件类型(csv/xlsx/xml)
	Separator              string         `gorm:"size:10" json:"separator"`                            // 分隔符
	SkipAttributeValidation bool           `gorm:"default:false" json:"skip_attribute_validation"`
	ColumnMapping          string         `gorm:"type:json" json:"column_mapping"`                     // 列映射JSON
	IsActive               bool           `gorm:"default:true" json:"is_active"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ImportProfile) TableName() string {
	return "import_profiles"
}

// ExportProfile 导出配置
type ExportProfile struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:256;not null" json:"name"`
	EntityType       string         `gorm:"size:50;not null;index" json:"entity_type"`          // 实体类型
	FilePath         string         `gorm:"size:512" json:"file_path"`                          // 导出文件路径
	FileType         string         `gorm:"size:20" json:"file_type"`                           // 文件类型
	ExportWithIds    bool           `gorm:"default:true" json:"export_with_ids"`                // 是否导出ID
	ColumnSelection  string         `gorm:"type:json" json:"column_selection"`                  // 列选择JSON
	FilterCriteria   string         `gorm:"type:json" json:"filter_criteria"`                   // 过滤条件JSON
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ExportProfile) TableName() string {
	return "export_profiles"
}

// ImportHistory 导入历史记录
type ImportHistory struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	ProfileID     uint           `gorm:"not null;index" json:"profile_id"`
	FileName      string         `gorm:"size:256;not null" json:"file_name"`
	TotalRecords  int            `gorm:"default:0" json:"total_records"`              // 总记录数
	SuccessCount  int            `gorm:"default:0" json:"success_count"`              // 成功数
	ErrorCount    int            `gorm:"default:0" json:"error_count"`                // 失败数
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`              // 错误信息
	Status        string         `gorm:"size:20;not null" json:"status"`              // pending/processing/completed/failed
	StartedAt     time.Time      `json:"started_at"`
	CompletedAt   time.Time      `json:"completed_at"`
	CreatedAt     time.Time      `json:"created_at"`
}

// TableName 指定表名
func (ImportHistory) TableName() string {
	return "import_histories"
}

// ExportHistory 导出历史记录
type ExportHistory struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	ProfileID    uint      `gorm:"not null;index" json:"profile_id"`
	FileName     string    `gorm:"size:256;not null" json:"file_name"`
	RecordCount  int       `gorm:"default:0" json:"record_count"`
	FileSize     int64     `gorm:"default:0" json:"file_size"`             // 文件大小(字节)
	FilePath     string    `gorm:"size:512" json:"file_path"`             // 导出文件路径
	Status       string    `gorm:"size:20;not null" json:"status"`         // pending/processing/completed/failed
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (ExportHistory) TableName() string {
	return "export_histories"
}

// ImportError 导入错误记录
type ImportError struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	HistoryID   uint64    `gorm:"not null;index" json:"history_id"`
	RowNumber   int       `gorm:"not null" json:"row_number"`            // 行号
	ColumnName  string    `gorm:"size:100" json:"column_name"`           // 列名
	FieldValue  string    `gorm:"size:500" json:"field_value"`           // 字段值
	ErrorMsg    string    `gorm:"type:text" json:"error_msg"`            // 错误信息
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (ImportError) TableName() string {
	return "import_errors"
}

// ========== DTO ==========

// ImportProfileCreateRequest 导入配置创建请求
type ImportProfileCreateRequest struct {
	Name                   string                 `json:"name" binding:"required"`
	EntityType             string                 `json:"entity_type" binding:"required"`
	FilePath               string                 `json:"file_path"`
	FileType               string                 `json:"file_type"`
	Separator              string                 `json:"separator"`
	SkipAttributeValidation bool                   `json:"skip_attribute_validation"`
	ColumnMapping          map[string]string      `json:"column_mapping"`
}

// ImportProfileUpdateRequest 导入配置更新请求
type ImportProfileUpdateRequest struct {
	Name                   string                 `json:"name"`
	FilePath               string                 `json:"file_path"`
	FileType               string                 `json:"file_type"`
	Separator              string                 `json:"separator"`
	SkipAttributeValidation bool                   `json:"skip_attribute_validation"`
	ColumnMapping          map[string]string      `json:"column_mapping"`
	IsActive               bool                   `json:"is_active"`
}

// ExportProfileCreateRequest 导出配置创建请求
type ExportProfileCreateRequest struct {
	Name            string                 `json:"name" binding:"required"`
	EntityType      string                 `json:"entity_type" binding:"required"`
	FilePath        string                 `json:"file_path"`
	FileType        string                 `json:"file_type"`
	ExportWithIds   bool                   `json:"export_with_ids"`
	ColumnSelection []string               `json:"column_selection"`
	FilterCriteria  map[string]interface{} `json:"filter_criteria"`
}

// ExportProfileUpdateRequest 导出配置更新请求
type ExportProfileUpdateRequest struct {
	Name            string                 `json:"name"`
	FilePath        string                 `json:"file_path"`
	FileType        string                 `json:"file_type"`
	ExportWithIds   bool                   `json:"export_with_ids"`
	ColumnSelection []string               `json:"column_selection"`
	FilterCriteria  map[string]interface{} `json:"filter_criteria"`
	IsActive        bool                   `json:"is_active"`
}

// ImportExecuteRequest 执行导入请求
type ImportExecuteRequest struct {
	ProfileID uint   `json:"profile_id" binding:"required"`
	FileName  string `json:"file_name" binding:"required"`
}

// ExportExecuteRequest 执行导出请求
type ExportExecuteRequest struct {
	ProfileID uint `json:"profile_id" binding:"required"`
}

// ImportResult 导入结果
type ImportResult struct {
	HistoryID    uint64 `json:"history_id"`
	TotalRecords int    `json:"total_records"`
	SuccessCount int    `json:"success_count"`
	ErrorCount   int    `json:"error_count"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// ExportResult 导出结果
type ExportResult struct {
	HistoryID   uint64 `json:"history_id"`
	RecordCount int    `json:"record_count"`
	FileSize    int64  `json:"file_size"`
	FilePath    string `json:"file_path"`
	DownloadURL string `json:"download_url"`
	Status      string `json:"status"`
}

// EntityType 实体类型定义
type EntityType struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Columns     []string `json:"columns"`
}

// 预定义实体类型
var EntityTypes = []EntityType{
	{Name: "product", DisplayName: "商品", Columns: []string{"name", "sku", "price", "stock", "category", "brand"}},
	{Name: "customer", DisplayName: "客户", Columns: []string{"username", "email", "phone", "first_name", "last_name"}},
	{Name: "order", DisplayName: "订单", Columns: []string{"order_no", "customer_id", "total", "status", "created_at"}},
	{Name: "category", DisplayName: "分类", Columns: []string{"name", "parent_id", "display_order", "is_active"}},
	{Name: "brand", DisplayName: "品牌", Columns: []string{"name", "description", "is_active"}},
}