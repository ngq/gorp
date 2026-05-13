// Package data 实现数据访问。
package data

import "gorm.io/gorm"

// Data 聚合数据访问依赖。
type Data struct {
	db *gorm.DB
}

// NewData 创建数据访问层。
func NewData(db *gorm.DB) *Data {
	return &Data{db: db}
}

// DB 返回数据库连接。
func (d *Data) DB() *gorm.DB {
	return d.db
}
