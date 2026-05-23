// Package biz 安全模块业务层 —— 权限与访问控制的核心领域逻辑
package biz

import "context"

// ==================== 领域实体 ====================

// Permission 权限实体 —— 描述系统中的操作权限
type Permission struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`         // 权限名称，如 "user:create"
	Code        string `json:"code"`         // 权限编码，用于程序判断，如 "USER_CREATE"
	Description string `json:"description"`  // 权限描述
	Module      string `json:"module"`       // 所属模块，如 "user"
	Action      string `json:"action"`       // 动作类型，如 "create"
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ACL 访问控制列表实体 —— 定义角色对资源的访问规则
type ACL struct {
	ID         uint   `json:"id"`
	RoleID     uint   `json:"role_id"`     // 角色ID
	Resource   string `json:"resource"`    // 资源标识，如 "/api/users"
	Action     string `json:"action"`      // 允许的动作，如 "GET", "POST", "*"
	Effect     string `json:"effect"`      // 效果：allow 或 deny
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// PermissionRepo 权限仓储接口 —— 定义权限数据访问的契约
type PermissionRepo interface {
	// Create 创建权限
	Create(ctx context.Context, p *Permission) error
	// GetByID 根据ID获取权限
	GetByID(ctx context.Context, id uint) (*Permission, error)
	// List 获取权限列表，支持模块过滤
	List(ctx context.Context, module string, page, pageSize int) ([]*Permission, int64, error)
	// Update 更新权限
	Update(ctx context.Context, p *Permission) error
	// Delete 删除权限
	Delete(ctx context.Context, id uint) error
}

// ACLRepo 访问控制列表仓储接口 —— 定义ACL数据访问的契约
type ACLRepo interface {
	// Create 创建ACL规则
	Create(ctx context.Context, acl *ACL) error
	// GetByID 根据ID获取ACL规则
	GetByID(ctx context.Context, id uint) (*ACL, error)
	// List 获取ACL规则列表，支持角色ID过滤
	List(ctx context.Context, roleID uint, page, pageSize int) ([]*ACL, int64, error)
	// Update 更新ACL规则
	Update(ctx context.Context, acl *ACL) error
	// Delete 删除ACL规则
	Delete(ctx context.Context, id uint) error
	// GetByRoleID 根据角色ID获取所有ACL规则
	GetByRoleID(ctx context.Context, roleID uint) ([]*ACL, error)
}

// ==================== 用例 ====================

// SecurityUseCase 安全模块用例 —— 封装权限与ACL的业务逻辑
type SecurityUseCase struct {
	permRepo PermissionRepo // 权限仓储
	aclRepo  ACLRepo        // 访问控制列表仓储
}

// NewSecurityUseCase 创建安全模块用例
func NewSecurityUseCase(permRepo PermissionRepo, aclRepo ACLRepo) *SecurityUseCase {
	return &SecurityUseCase{
		permRepo: permRepo,
		aclRepo:  aclRepo,
	}
}

// ---------- 权限用例方法 ----------

// CreatePermission 创建权限
func (uc *SecurityUseCase) CreatePermission(ctx context.Context, p *Permission) error {
	return uc.permRepo.Create(ctx, p)
}

// GetPermission 根据ID获取权限
func (uc *SecurityUseCase) GetPermission(ctx context.Context, id uint) (*Permission, error) {
	return uc.permRepo.GetByID(ctx, id)
}

// ListPermissions 获取权限列表
func (uc *SecurityUseCase) ListPermissions(ctx context.Context, module string, page, pageSize int) ([]*Permission, int64, error) {
	return uc.permRepo.List(ctx, module, page, pageSize)
}

// UpdatePermission 更新权限
func (uc *SecurityUseCase) UpdatePermission(ctx context.Context, p *Permission) error {
	return uc.permRepo.Update(ctx, p)
}

// DeletePermission 删除权限
func (uc *SecurityUseCase) DeletePermission(ctx context.Context, id uint) error {
	return uc.permRepo.Delete(ctx, id)
}

// ---------- ACL用例方法 ----------

// CreateACL 创建ACL规则
func (uc *SecurityUseCase) CreateACL(ctx context.Context, acl *ACL) error {
	return uc.aclRepo.Create(ctx, acl)
}

// GetACL 根据ID获取ACL规则
func (uc *SecurityUseCase) GetACL(ctx context.Context, id uint) (*ACL, error) {
	return uc.aclRepo.GetByID(ctx, id)
}

// ListACLs 获取ACL规则列表
func (uc *SecurityUseCase) ListACLs(ctx context.Context, roleID uint, page, pageSize int) ([]*ACL, int64, error) {
	return uc.aclRepo.List(ctx, roleID, page, pageSize)
}

// UpdateACL 更新ACL规则
func (uc *SecurityUseCase) UpdateACL(ctx context.Context, acl *ACL) error {
	return uc.aclRepo.Update(ctx, acl)
}

// DeleteACL 删除ACL规则
func (uc *SecurityUseCase) DeleteACL(ctx context.Context, id uint) error {
	return uc.aclRepo.Delete(ctx, id)
}

// GetACLsByRoleID 根据角色ID获取所有ACL规则
func (uc *SecurityUseCase) GetACLsByRoleID(ctx context.Context, roleID uint) ([]*ACL, error) {
	return uc.aclRepo.GetByRoleID(ctx, roleID)
}
