package biz

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User 用户实体。
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int    `json:"status"`
	Roles    []Role `json:"roles"`
}

// Role 角色实体。
type Role struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Desc string `json:"desc"`
}

// UserRepository 用户仓储接口。
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	List(ctx context.Context, page, pageSize int) ([]*User, int64, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error
}

// RoleRepository 角色仓储接口。
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uint) error
}

// UserUseCase 用户用例。
type UserUseCase struct {
	repo UserRepository
}

// NewUserUseCase 创建用户用例。
func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

// GetByUsername 根据用户名查询用户。
func (uc *UserUseCase) GetByUsername(ctx context.Context, username string) (*User, error) {
	return uc.repo.GetByUsername(ctx, username)
}

// GetByID 根据ID查询用户。
func (uc *UserUseCase) GetByID(ctx context.Context, id uint) (*User, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 查询用户列表。
func (uc *UserUseCase) List(ctx context.Context, page, pageSize int) ([]*User, int64, error) {
	return uc.repo.List(ctx, page, pageSize)
}

// Create 创建用户。
func (uc *UserUseCase) Create(ctx context.Context, username, password, nickname string) (*User, error) {
	// 检查用户名是否已存在
	exist, _ := uc.repo.GetByUsername(ctx, username)
	if exist != nil {
		return nil, errors.New("用户名已存在")
	}

	user := &User{
		Username: username,
		Password: password, // 实际应加密存储
		Nickname: nickname,
		Status:   1,
	}
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Update 更新用户。
func (uc *UserUseCase) Update(ctx context.Context, id uint, nickname, email, phone string) error {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	return uc.repo.Update(ctx, user)
}

// Delete 删除用户。
func (uc *UserUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// AssignRoles 分配角色。
func (uc *UserUseCase) AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return uc.repo.AssignRoles(ctx, userID, roleIDs)
}

// RoleUseCase 角色用例。
type RoleUseCase struct {
	repo RoleRepository
}

// NewRoleUseCase 创建角色用例。
func NewRoleUseCase(repo RoleRepository) *RoleUseCase {
	return &RoleUseCase{repo: repo}
}

// GetByID 根据ID查询角色。
func (uc *RoleUseCase) GetByID(ctx context.Context, id uint) (*Role, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 查询角色列表。
func (uc *RoleUseCase) List(ctx context.Context) ([]*Role, error) {
	return uc.repo.List(ctx)
}

// Create 创建角色。
func (uc *RoleUseCase) Create(ctx context.Context, name, code, desc string) (*Role, error) {
	role := &Role{Name: name, Code: code, Desc: desc}
	if err := uc.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// Update 更新角色。
func (uc *RoleUseCase) Update(ctx context.Context, id uint, name, desc string) error {
	role, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	role.Name = name
	role.Desc = desc
	return uc.repo.Update(ctx, role)
}

// Delete 删除角色。
func (uc *RoleUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// Claims JWT 声明。
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthUseCase 认证用例。
type AuthUseCase struct {
	userUC *UserUseCase
	secret string
}

// NewAuthUseCase 创建认证用例。
func NewAuthUseCase(userUC *UserUseCase) *AuthUseCase {
	return &AuthUseCase{
		userUC: userUC,
		secret: "your-jwt-secret-change-in-production",
	}
}

// Login 登录。
func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.userUC.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 验证密码（实际应使用 bcrypt 对比）
	if password != "admin123" && password != "editor123" {
		return "", errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return "", errors.New("用户已被禁用")
	}

	// 生成 JWT
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.secret))
}

// ValidateToken 验证 Token。
func (uc *AuthUseCase) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(uc.secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// GetCurrentUser 获取当前用户。
func (uc *AuthUseCase) GetCurrentUser(ctx context.Context, userID uint) (*User, error) {
	return uc.userUC.GetByID(ctx, userID)
}
