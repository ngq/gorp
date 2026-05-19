package data

import (
	"context"
	"time"

	"admin/internal/biz"

	"gorm.io/gorm"
)

// UserPO 用户持久化对象。
type UserPO struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string         `gorm:"size:256;not null" json:"-"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	Email     string         `gorm:"size:128" json:"email"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Status    int            `gorm:"default:1" json:"status"` // 1: 正常, 0: 禁用
	Roles     []RolePO       `gorm:"many2many:user_roles;" json:"roles"`
}

// TableName 表名。
func (UserPO) TableName() string {
	return "sys_user"
}

// ToEntity 转换为领域实体。
func (po *UserPO) ToEntity() *biz.User {
	user := &biz.User{
		ID:       po.ID,
		Username: po.Username,
		Nickname: po.Nickname,
		Email:    po.Email,
		Phone:    po.Phone,
		Status:   po.Status,
	}
	user.Roles = make([]biz.Role, len(po.Roles))
	for i, role := range po.Roles {
		user.Roles[i] = biz.Role{
			ID:   role.ID,
			Name: role.Name,
			Code: role.Code,
			Desc: role.Desc,
		}
	}
	return user
}

// ToUserPO 领域实体转换为 PO。
func ToUserPO(user *biz.User) *UserPO {
	return &UserPO{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Status:   user.Status,
	}
}

// RolePO 角色持久化对象。
type RolePO struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Desc      string         `gorm:"size:256" json:"desc"`
	Status    int            `gorm:"default:1" json:"status"`
}

// TableName 表名。
func (RolePO) TableName() string {
	return "sys_role"
}

// ToEntity 转换为领域实体。
func (po *RolePO) ToEntity() *biz.Role {
	return &biz.Role{
		ID:   po.ID,
		Name: po.Name,
		Code: po.Code,
		Desc: po.Desc,
	}
}

// ToRolePO 领域实体转换为 PO。
func ToRolePO(role *biz.Role) *RolePO {
	return &RolePO{
		ID:   role.ID,
		Name: role.Name,
		Code: role.Code,
		Desc: role.Desc,
	}
}

// UserRepo 用户仓储，实现 biz.UserRepository 接口。
type UserRepo struct {
	data *Data
}

// NewUserRepo 创建用户仓储。
func NewUserRepo(data *Data) *UserRepo {
	return &UserRepo{data: data}
}

// GetByUsername 根据用户名查询用户。
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	var po UserPO
	err := r.data.DB().WithContext(ctx).Where("username = ?", username).First(&po).Error
	if err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByID 根据ID查询用户（含角色）。
func (r *UserRepo) GetByID(ctx context.Context, id uint) (*biz.User, error) {
	var po UserPO
	err := r.data.DB().WithContext(ctx).Preload("Roles").First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 查询用户列表。
func (r *UserRepo) List(ctx context.Context, page, pageSize int) ([]*biz.User, int64, error) {
	var users []UserPO
	var total int64

	db := r.data.DB().WithContext(ctx).Model(&UserPO{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.User, len(users))
	for i, po := range users {
		result[i] = po.ToEntity()
	}
	return result, total, nil
}

// Create 创建用户。
func (r *UserRepo) Create(ctx context.Context, user *biz.User) error {
	po := ToUserPO(user)
	return r.data.DB().WithContext(ctx).Create(po).Error
}

// Update 更新用户。
func (r *UserRepo) Update(ctx context.Context, user *biz.User) error {
	po := ToUserPO(user)
	return r.data.DB().WithContext(ctx).Save(po).Error
}

// Delete 删除用户。
func (r *UserRepo) Delete(ctx context.Context, id uint) error {
	return r.data.DB().WithContext(ctx).Delete(&UserPO{}, id).Error
}

// AssignRoles 分配角色。
func (r *UserRepo) AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	var roles []RolePO
	if len(roleIDs) > 0 {
		r.data.DB().WithContext(ctx).Find(&roles, roleIDs)
	}
	return r.data.DB().WithContext(ctx).Model(&UserPO{ID: userID}).Association("Roles").Replace(roles)
}

// RoleRepo 角色仓储，实现 biz.RoleRepository 接口。
type RoleRepo struct {
	data *Data
}

// NewRoleRepo 创建角色仓储。
func NewRoleRepo(data *Data) *RoleRepo {
	return &RoleRepo{data: data}
}

// GetByID 根据ID查询角色。
func (r *RoleRepo) GetByID(ctx context.Context, id uint) (*biz.Role, error) {
	var po RolePO
	err := r.data.DB().WithContext(ctx).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 查询角色列表。
func (r *RoleRepo) List(ctx context.Context) ([]*biz.Role, error) {
	var roles []RolePO
	err := r.data.DB().WithContext(ctx).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	result := make([]*biz.Role, len(roles))
	for i, po := range roles {
		result[i] = po.ToEntity()
	}
	return result, nil
}

// Create 创建角色。
func (r *RoleRepo) Create(ctx context.Context, role *biz.Role) error {
	po := ToRolePO(role)
	return r.data.DB().WithContext(ctx).Create(po).Error
}

// Update 更新角色。
func (r *RoleRepo) Update(ctx context.Context, role *biz.Role) error {
	po := ToRolePO(role)
	return r.data.DB().WithContext(ctx).Save(po).Error
}

// Delete 删除角色。
func (r *RoleRepo) Delete(ctx context.Context, id uint) error {
	return r.data.DB().WithContext(ctx).Delete(&RolePO{}, id).Error
}
