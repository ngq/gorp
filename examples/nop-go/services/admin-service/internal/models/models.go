// Package models 管理后台服务数据模型
package models

import (
	"time"
)

// AdminUser 管理员用户
type AdminUser struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"size:128;uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"size:256;not null" json:"-"`
	DisplayName  string     `gorm:"size:64" json:"display_name"`
	AvatarURL    string     `gorm:"size:512" json:"avatar_url"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string     `gorm:"size:64" json:"last_login_ip"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Roles []AdminRole `gorm:"many2many:admin_user_roles;" json:"roles,omitempty"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

// AdminRole 管理员角色
type AdminRole struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	SystemName  string    `gorm:"size:64;uniqueIndex" json:"system_name"`
	Description string    `gorm:"type:text" json:"description"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	Permissions []AdminPermission `gorm:"many2many:admin_role_permissions;" json:"permissions,omitempty"`
}

func (AdminRole) TableName() string {
	return "admin_roles"
}

// AdminPermission 权限
type AdminPermission struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	SystemName  string    `gorm:"size:64;uniqueIndex;not null" json:"system_name"`
	Category    string    `gorm:"size:64;index" json:"category"` // catalog, order, customer, etc.
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AdminPermission) TableName() string {
	return "admin_permissions"
}

// Setting 系统设置
type Setting struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Value       string    `gorm:"type:text" json:"value"`
	Description string    `gorm:"size:256" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}

// ActivityLog 操作日志
type ActivityLog struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	AdminID    uint64    `gorm:"not null;index" json:"admin_id"`
	Action     string    `gorm:"size:64;not null;index" json:"action"`
	Entity     string    `gorm:"size:64;not null" json:"entity"`
	EntityID   uint64    `json:"entity_id"`
	OldData    string    `gorm:"type:json" json:"old_data"`
	NewData    string    `gorm:"type:json" json:"new_data"`
	IP         string    `gorm:"size:64" json:"ip"`
	UserAgent  string    `gorm:"size:256" json:"user_agent"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TotalOrders       int64   `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalCustomers    int64   `json:"total_customers"`
	TotalProducts     int64   `json:"total_products"`
	PendingOrders     int64   `json:"pending_orders"`
	LowStockProducts  int64   `json:"low_stock_products"`
	TodayOrders       int64   `json:"today_orders"`
	TodayRevenue      float64 `json:"today_revenue"`
}

// DTO
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token string          `json:"token"`
	User  AdminUserResponse `json:"user"`
}

type AdminUserResponse struct {
	ID          uint64   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	AvatarURL   string   `json:"avatar_url"`
	Roles       []string `json:"roles"`
}

func ToAdminUserResponse(u *AdminUser) AdminUserResponse {
	resp := AdminUserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}
	if len(u.Roles) > 0 {
		resp.Roles = make([]string, len(u.Roles))
		for i, role := range u.Roles {
			resp.Roles[i] = role.Name
		}
	}
	return resp
}

type UpdateSettingRequest struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}