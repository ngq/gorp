// Package data 管理后台服务数据访问层
package data

import (
	"context"

	"nop-go/services/admin-service/internal/models"

	"gorm.io/gorm"
)

type AdminUserRepository interface {
	Create(ctx context.Context, user *models.AdminUser) error
	GetByID(ctx context.Context, id uint64) (*models.AdminUser, error)
	GetByUsername(ctx context.Context, username string) (*models.AdminUser, error)
	GetByEmail(ctx context.Context, email string) (*models.AdminUser, error)
	Update(ctx context.Context, user *models.AdminUser) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, page, pageSize int) ([]*models.AdminUser, int64, error)
}

type AdminRoleRepository interface {
	Create(ctx context.Context, role *models.AdminRole) error
	GetByID(ctx context.Context, id uint64) (*models.AdminRole, error)
	GetBySystemName(ctx context.Context, systemName string) (*models.AdminRole, error)
	Update(ctx context.Context, role *models.AdminRole) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.AdminRole, error)
}

type AdminPermissionRepository interface {
	GetByID(ctx context.Context, id uint64) (*models.AdminPermission, error)
	GetBySystemName(ctx context.Context, systemName string) (*models.AdminPermission, error)
	List(ctx context.Context) ([]*models.AdminPermission, error)
	ListByRoleID(ctx context.Context, roleID uint64) ([]*models.AdminPermission, error)
}

type SettingRepository interface {
	GetByName(ctx context.Context, name string) (*models.Setting, error)
	GetAll(ctx context.Context) ([]*models.Setting, error)
	Set(ctx context.Context, name, value string) error
	Delete(ctx context.Context, name string) error
}

type ActivityLogRepository interface {
	Create(ctx context.Context, log *models.ActivityLog) error
	GetByAdminID(ctx context.Context, adminID uint64, limit int) ([]*models.ActivityLog, error)
	List(ctx context.Context, page, pageSize int) ([]*models.ActivityLog, int64, error)
}

type adminUserRepo struct{ db *gorm.DB }

func NewAdminUserRepository(db *gorm.DB) AdminUserRepository {
	return &adminUserRepo{db: db}
}

func (r *adminUserRepo) Create(ctx context.Context, u *models.AdminUser) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *adminUserRepo) GetByID(ctx context.Context, id uint64) (*models.AdminUser, error) {
	var u models.AdminUser
	err := r.db.WithContext(ctx).Preload("Roles").First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *adminUserRepo) GetByUsername(ctx context.Context, username string) (*models.AdminUser, error) {
	var u models.AdminUser
	err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *adminUserRepo) GetByEmail(ctx context.Context, email string) (*models.AdminUser, error) {
	var u models.AdminUser
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *adminUserRepo) Update(ctx context.Context, u *models.AdminUser) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *adminUserRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.AdminUser{}, id).Error
}

func (r *adminUserRepo) List(ctx context.Context, page, pageSize int) ([]*models.AdminUser, int64, error) {
	var list []*models.AdminUser
	var total int64
	db := r.db.WithContext(ctx).Model(&models.AdminUser{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type adminRoleRepo struct{ db *gorm.DB }

func NewAdminRoleRepository(db *gorm.DB) AdminRoleRepository {
	return &adminRoleRepo{db: db}
}

func (r *adminRoleRepo) Create(ctx context.Context, role *models.AdminRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *adminRoleRepo) GetByID(ctx context.Context, id uint64) (*models.AdminRole, error) {
	var role models.AdminRole
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *adminRoleRepo) GetBySystemName(ctx context.Context, systemName string) (*models.AdminRole, error) {
	var role models.AdminRole
	err := r.db.WithContext(ctx).Where("system_name = ?", systemName).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *adminRoleRepo) Update(ctx context.Context, role *models.AdminRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *adminRoleRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.AdminRole{}, id).Error
}

func (r *adminRoleRepo) List(ctx context.Context) ([]*models.AdminRole, error) {
	var list []*models.AdminRole
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

type adminPermissionRepo struct{ db *gorm.DB }

func NewAdminPermissionRepository(db *gorm.DB) AdminPermissionRepository {
	return &adminPermissionRepo{db: db}
}

func (r *adminPermissionRepo) GetByID(ctx context.Context, id uint64) (*models.AdminPermission, error) {
	var p models.AdminPermission
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *adminPermissionRepo) GetBySystemName(ctx context.Context, systemName string) (*models.AdminPermission, error) {
	var p models.AdminPermission
	err := r.db.WithContext(ctx).Where("system_name = ?", systemName).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *adminPermissionRepo) List(ctx context.Context) ([]*models.AdminPermission, error) {
	var list []*models.AdminPermission
	err := r.db.WithContext(ctx).Order("category, name").Find(&list).Error
	return list, err
}

func (r *adminPermissionRepo) ListByRoleID(ctx context.Context, roleID uint64) ([]*models.AdminPermission, error) {
	var list []*models.AdminPermission
	err := r.db.WithContext(ctx).
		Joins("JOIN admin_role_permissions ON admin_role_permissions.admin_permission_id = admin_permissions.id").
		Where("admin_role_permissions.admin_role_id = ?", roleID).
		Find(&list).Error
	return list, err
}

type settingRepo struct{ db *gorm.DB }

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepo{db: db}
}

func (r *settingRepo) GetByName(ctx context.Context, name string) (*models.Setting, error) {
	var s models.Setting
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settingRepo) GetAll(ctx context.Context) ([]*models.Setting, error) {
	var list []*models.Setting
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *settingRepo) Set(ctx context.Context, name, value string) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO settings (name, value, created_at, updated_at) VALUES (?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE value = ?, updated_at = NOW()",
		name, value, value,
	).Error
}

func (r *settingRepo) Delete(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Where("name = ?", name).Delete(&models.Setting{}).Error
}

type activityLogRepo struct{ db *gorm.DB }

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepo{db: db}
}

func (r *activityLogRepo) Create(ctx context.Context, log *models.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *activityLogRepo) GetByAdminID(ctx context.Context, adminID uint64, limit int) ([]*models.ActivityLog, error) {
	var list []*models.ActivityLog
	err := r.db.WithContext(ctx).Where("admin_id = ?", adminID).
		Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *activityLogRepo) List(ctx context.Context, page, pageSize int) ([]*models.ActivityLog, int64, error) {
	var list []*models.ActivityLog
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ActivityLog{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}