// Package biz 管理后台服务业务逻辑层
package biz

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/ngq/gorp/framework/contract"
	"nop-go/services/admin-service/internal/data"
	"nop-go/services/admin-service/internal/models"
)

type AdminUserUseCase struct {
	userRepo data.AdminUserRepository
	roleRepo data.AdminRoleRepository
	logRepo  data.ActivityLogRepository
	jwtSvc   contract.JWTService
}

// NewAdminUserUseCase 创建管理后台用户 UseCase。
//
// 中文说明：
// - 使用 framework 级 JWTService，替代项目层 jwtSecret/jwtExpire；
// - JWTService 统一处理签发/验证，配置从 auth.jwt.* 读取。
func NewAdminUserUseCase(userRepo data.AdminUserRepository, roleRepo data.AdminRoleRepository, logRepo data.ActivityLogRepository, jwtSvc contract.JWTService) *AdminUserUseCase {
	return &AdminUserUseCase{userRepo: userRepo, roleRepo: roleRepo, logRepo: logRepo, jwtSvc: jwtSvc}
}

func (uc *AdminUserUseCase) Login(ctx context.Context, username, password string) (*models.AdminUser, string, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, "", errors.New("account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.SystemName)
	}

	// 使用 framework JWTService 签发 token
	claims := uc.jwtSvc.NewClaims(int64(user.ID), "admin", user.Username, roles, 86400)
	token, err := uc.jwtSvc.Sign(claims)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (uc *AdminUserUseCase) CreateAdmin(ctx context.Context, user *models.AdminUser, password string) error {
	existing, _ := uc.userRepo.GetByUsername(ctx, user.Username)
	if existing != nil {
		return errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return uc.userRepo.Create(ctx, user)
}

func (uc *AdminUserUseCase) GetAdmin(ctx context.Context, id uint64) (*models.AdminUser, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *AdminUserUseCase) ListAdmins(ctx context.Context, page, pageSize int) ([]*models.AdminUser, int64, error) {
	return uc.userRepo.List(ctx, page, pageSize)
}

func (uc *AdminUserUseCase) UpdateAdmin(ctx context.Context, user *models.AdminUser) error {
	return uc.userRepo.Update(ctx, user)
}

func (uc *AdminUserUseCase) ChangePassword(ctx context.Context, id uint64, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return uc.userRepo.Update(ctx, user)
}

func (uc *AdminUserUseCase) DeleteAdmin(ctx context.Context, id uint64) error {
	return uc.userRepo.Delete(ctx, id)
}

type AdminRoleUseCase struct {
	roleRepo       data.AdminRoleRepository
	permissionRepo data.AdminPermissionRepository
}

func NewAdminRoleUseCase(roleRepo data.AdminRoleRepository, permissionRepo data.AdminPermissionRepository) *AdminRoleUseCase {
	return &AdminRoleUseCase{roleRepo: roleRepo, permissionRepo: permissionRepo}
}

func (uc *AdminRoleUseCase) CreateRole(ctx context.Context, role *models.AdminRole) error {
	return uc.roleRepo.Create(ctx, role)
}

func (uc *AdminRoleUseCase) GetRole(ctx context.Context, id uint64) (*models.AdminRole, error) {
	return uc.roleRepo.GetByID(ctx, id)
}

func (uc *AdminRoleUseCase) ListRoles(ctx context.Context) ([]*models.AdminRole, error) {
	return uc.roleRepo.List(ctx)
}

func (uc *AdminRoleUseCase) UpdateRole(ctx context.Context, role *models.AdminRole) error {
	return uc.roleRepo.Update(ctx, role)
}

func (uc *AdminRoleUseCase) DeleteRole(ctx context.Context, id uint64) error {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	return uc.roleRepo.Delete(ctx, id)
}

func (uc *AdminRoleUseCase) ListPermissions(ctx context.Context) ([]*models.AdminPermission, error) {
	return uc.permissionRepo.List(ctx)
}

type SettingUseCase struct {
	settingRepo data.SettingRepository
}

func NewSettingUseCase(settingRepo data.SettingRepository) *SettingUseCase {
	return &SettingUseCase{settingRepo: settingRepo}
}

func (uc *SettingUseCase) GetSetting(ctx context.Context, name string) (string, error) {
	setting, err := uc.settingRepo.GetByName(ctx, name)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (uc *SettingUseCase) GetAllSettings(ctx context.Context) (map[string]string, error) {
	settings, err := uc.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Name] = s.Value
	}
	return result, nil
}

func (uc *SettingUseCase) SetSetting(ctx context.Context, name, value string) error {
	return uc.settingRepo.Set(ctx, name, value)
}

func (uc *SettingUseCase) DeleteSetting(ctx context.Context, name string) error {
	return uc.settingRepo.Delete(ctx, name)
}

type ActivityLogUseCase struct {
	logRepo data.ActivityLogRepository
}

func NewActivityLogUseCase(logRepo data.ActivityLogRepository) *ActivityLogUseCase {
	return &ActivityLogUseCase{logRepo: logRepo}
}

func (uc *ActivityLogUseCase) Log(ctx context.Context, adminID uint64, action, entity string, entityID uint64, oldData, newData, ip, userAgent string) error {
	log := &models.ActivityLog{
		AdminID:   adminID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		OldData:   oldData,
		NewData:   newData,
		IP:        ip,
		UserAgent: userAgent,
	}
	return uc.logRepo.Create(ctx, log)
}

func (uc *ActivityLogUseCase) GetAdminLogs(ctx context.Context, adminID uint64, limit int) ([]*models.ActivityLog, error) {
	return uc.logRepo.GetByAdminID(ctx, adminID, limit)
}

func (uc *ActivityLogUseCase) ListLogs(ctx context.Context, page, pageSize int) ([]*models.ActivityLog, int64, error) {
	return uc.logRepo.List(ctx, page, pageSize)
}