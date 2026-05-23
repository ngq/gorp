// Package data 安全模块数据层 —— 权限与ACL的持久化对象与仓储实现
package data

import (
	"context"

	"nop-go/services/admin-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== 持久化对象（PO） ====================

// PermissionPO 权限持久化对象 —— 映射数据库 permissions 表
type PermissionPO struct {
	ID          uint   `gorm:"primaryKey" db:"id"`
	Name        string `gorm:"column:name;type:varchar(100);not null" db:"name"`        // 权限名称
	Code        string `gorm:"column:code;type:varchar(100);uniqueIndex;not null" db:"code"` // 权限编码
	Description string `gorm:"column:description;type:varchar(500)" db:"description"`           // 权限描述
	Module      string `gorm:"column:module;type:varchar(50);index" db:"module"`           // 所属模块
	Action      string `gorm:"column:action;type:varchar(50)" db:"action"`                 // 动作类型
	CreatedAt   string `gorm:"column:created_at" db:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at" db:"updated_at"`
}

// TableName 指定权限表名
func (PermissionPO) TableName() string { return "permissions" }

// ACLPO 访问控制列表持久化对象 —— 映射数据库 acls 表
type ACLPO struct {
	ID        uint   `gorm:"primaryKey" db:"id"`
	RoleID    uint   `gorm:"column:role_id;type:bigint;index;not null" db:"role_id"` // 角色ID
	Resource  string `gorm:"column:resource;type:varchar(200);not null" db:"resource"` // 资源标识
	Action    string `gorm:"column:action;type:varchar(50);not null" db:"action"`    // 允许的动作
	Effect    string `gorm:"column:effect;type:varchar(20);not null" db:"effect"`    // 效果：allow/deny
	CreatedAt string `gorm:"column:created_at" db:"created_at"`
	UpdatedAt string `gorm:"column:updated_at" db:"updated_at"`
}

// TableName 指定ACL表名
func (ACLPO) TableName() string { return "acls" }

// ==================== PO ↔ Entity 转换 ====================

// toEntity 将 PermissionPO 转换为 biz.Permission 领域实体
func (p *PermissionPO) toEntity() *biz.Permission {
	return &biz.Permission{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Module:      p.Module,
		Action:      p.Action,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// toEntity 将 ACLPO 转换为 biz.ACL 领域实体
func (a *ACLPO) toEntity() *biz.ACL {
	return &biz.ACL{
		ID:        a.ID,
		RoleID:    a.RoleID,
		Resource:  a.Resource,
		Action:    a.Action,
		Effect:    a.Effect,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// permissionToPO 将 biz.Permission 领域实体转换为 PermissionPO
func permissionToPO(p *biz.Permission) *PermissionPO {
	return &PermissionPO{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Module:      p.Module,
		Action:      p.Action,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// aclToPO 将 biz.ACL 领域实体转换为 ACLPO
func aclToPO(a *biz.ACL) *ACLPO {
	return &ACLPO{
		ID:        a.ID,
		RoleID:    a.RoleID,
		Resource:  a.Resource,
		Action:    a.Action,
		Effect:    a.Effect,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// permissionRepo 权限仓储实现
type permissionRepo struct {
	db *gorm.DB
}

// NewPermissionRepo 创建权限仓储
func NewPermissionRepo(db *gorm.DB) biz.PermissionRepo {
	return &permissionRepo{db: db}
}

func (r *permissionRepo) Create(ctx context.Context, p *biz.Permission) error {
	po := permissionToPO(p)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *permissionRepo) GetByID(ctx context.Context, id uint) (*biz.Permission, error) {
	var po PermissionPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *permissionRepo) List(ctx context.Context, module string, page, pageSize int) ([]*biz.Permission, int64, error) {
	var pos []*PermissionPO
	var total int64
	q := r.db.WithContext(ctx).Model(&PermissionPO{})
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.Permission, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *permissionRepo) Update(ctx context.Context, p *biz.Permission) error {
	po := permissionToPO(p)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *permissionRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PermissionPO{}, id).Error
}

// aclRepo 访问控制列表仓储实现
type aclRepo struct {
	db *gorm.DB
}

// NewACLRepo 创建ACL仓储
func NewACLRepo(db *gorm.DB) biz.ACLRepo {
	return &aclRepo{db: db}
}

func (r *aclRepo) Create(ctx context.Context, acl *biz.ACL) error {
	po := aclToPO(acl)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *aclRepo) GetByID(ctx context.Context, id uint) (*biz.ACL, error) {
	var po ACLPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *aclRepo) List(ctx context.Context, roleID uint, page, pageSize int) ([]*biz.ACL, int64, error) {
	var pos []*ACLPO
	var total int64
	q := r.db.WithContext(ctx).Model(&ACLPO{})
	if roleID > 0 {
		q = q.Where("role_id = ?", roleID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.ACL, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *aclRepo) Update(ctx context.Context, acl *biz.ACL) error {
	po := aclToPO(acl)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *aclRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ACLPO{}, id).Error
}

func (r *aclRepo) GetByRoleID(ctx context.Context, roleID uint) ([]*biz.ACL, error) {
	var pos []*ACLPO
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.ACL, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, nil
}
