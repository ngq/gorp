// Package models 网关服务数据模型
package models

import (
	"time"
)

// Route 路由配置
type Route struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Path        string    `gorm:"size:256;not null;uniqueIndex" json:"path"`
	ServiceName string    `gorm:"size:64;not null" json:"service_name"`
	StripPrefix bool      `gorm:"default:false" json:"strip_prefix"`
	RateLimit   int       `gorm:"default:0" json:"rate_limit"` // 请求/秒, 0=无限制
	Timeout     int       `gorm:"default:30" json:"timeout"`   // 秒
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Route) TableName() string {
	return "routes"
}

// RateLimitRecord 限流记录
type RateLimitRecord struct {
	Key       string    `json:"key"`
	Count     int       `json:"count"`
	ResetAt   time.Time `json:"reset_at"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Healthy   bool      `json:"healthy"`
	LastCheck time.Time `json:"last_check"`
}

// DTO
type RouteResponse struct {
	Path        string `json:"path"`
	ServiceName string `json:"service_name"`
	StripPrefix bool   `json:"strip_prefix"`
	RateLimit   int    `json:"rate_limit"`
	Timeout     int    `json:"timeout"`
}

type HealthResponse struct {
	Status   string                     `json:"status"`
	Services map[string]ServiceHealth   `json:"services"`
}