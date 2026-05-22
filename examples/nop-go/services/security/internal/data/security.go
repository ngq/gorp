// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/security/internal/biz"

	"gorm.io/gorm"
)

// PermissionPO 权限持久化对象。
type PermissionPO struct {
	ID           uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name         string    `gorm:"size:256;column:name" db:"name" json:"name"`
	SystemName   string    `gorm:"size:256;uniqueIndex;column:system_name" db:"system_name" json:"system_name"`
	Category     string    `gorm:"size:128;column:category" db:"category" json:"category"`
	DisplayOrder int       `gorm:"column:display_order" db:"display_order" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (PermissionPO) TableName() string {
	return "permissions"
}

// ToEntity 转换为领域实体。
func (po *PermissionPO) ToEntity() *biz.Permission {
	return &biz.Permission{
		ID:           po.ID,
		Name:         po.Name,
		SystemName:   po.SystemName,
		Category:     po.Category,
		DisplayOrder: po.DisplayOrder,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// ACLRecordPO ACL记录持久化对象。
type ACLRecordPO struct {
	ID             uint   `gorm:"primaryKey;column:id" db:"id" json:"id"`
	UserID         uint   `gorm:"index;column:user_id" db:"user_id" json:"user_id"`
	PermissionID   uint   `gorm:"index;column:permission_id" db:"permission_id" json:"permission_id"`
	PermissionName string `gorm:"size:256;column:permission_name" db:"permission_name" json:"permission_name"`
}

// TableName 表名。
func (ACLRecordPO) TableName() string {
	return "acl_records"
}

// ToEntity 转换为领域实体。
func (po *ACLRecordPO) ToEntity() *biz.ACLRecord {
	return &biz.ACLRecord{
		ID:             po.ID,
		UserID:         po.UserID,
		PermissionID:   po.PermissionID,
		PermissionName: po.PermissionName,
	}
}

// permissionRepo 权限仓储实现。
type permissionRepo struct {
	db *gorm.DB
}

// NewPermissionRepo 创建权限仓储。
func NewPermissionRepo(db *gorm.DB) biz.PermissionRepository {
	return &permissionRepo{db: db}
}

// Create 创建权限。
func (r *permissionRepo) Create(ctx context.Context, perm *biz.Permission) error {
	po := &PermissionPO{
		Name:         perm.Name,
		SystemName:   perm.SystemName,
		Category:     perm.Category,
		DisplayOrder: perm.DisplayOrder,
		CreatedAt:    perm.CreatedAt,
		UpdatedAt:    perm.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取权限。
func (r *permissionRepo) GetByID(ctx context.Context, id uint) (*biz.Permission, error) {
	var po PermissionPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取权限列表。
func (r *permissionRepo) List(ctx context.Context, page, size int) ([]*biz.Permission, int64, error) {
	var pos []PermissionPO
	var total int64

	r.db.WithContext(ctx).Model(&PermissionPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	perms := make([]*biz.Permission, len(pos))
	for i, po := range pos {
		perms[i] = po.ToEntity()
	}

	return perms, total, nil
}

// Update 更新权限。
func (r *permissionRepo) Update(ctx context.Context, perm *biz.Permission) error {
	return r.db.WithContext(ctx).Model(&PermissionPO{}).Where("id = ?", perm.ID).Updates(map[string]interface{}{
		"name":          perm.Name,
		"system_name":   perm.SystemName,
		"category":      perm.Category,
		"display_order": perm.DisplayOrder,
		"updated_at":    perm.UpdatedAt,
	}).Error
}

// Delete 删除权限。
func (r *permissionRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PermissionPO{}, id).Error
}

// aclRecordRepo ACL记录仓储实现。
type aclRecordRepo struct {
	db *gorm.DB
}

// NewACLRecordRepo 创建ACL记录仓储。
func NewACLRecordRepo(db *gorm.DB) biz.ACLRepository {
	return &aclRecordRepo{db: db}
}

// Create 创建ACL记录。
func (r *aclRecordRepo) Create(ctx context.Context, record *biz.ACLRecord) error {
	po := &ACLRecordPO{
		UserID:         record.UserID,
		PermissionID:   record.PermissionID,
		PermissionName: record.PermissionName,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// List 获取ACL记录列表。
func (r *aclRecordRepo) List(ctx context.Context, page, size int) ([]*biz.ACLRecord, int64, error) {
	var pos []ACLRecordPO
	var total int64

	r.db.WithContext(ctx).Model(&ACLRecordPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	records := make([]*biz.ACLRecord, len(pos))
	for i, po := range pos {
		records[i] = po.ToEntity()
	}

	return records, total, nil
}

// Delete 删除ACL记录。
func (r *aclRecordRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ACLRecordPO{}, id).Error
}