// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// User 用户领域实体。
type User struct {
	ID        uint
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository 用户仓储接口。
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context, page, size int) ([]*User, int64, error)
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

// Create 创建用户。
func (uc *UserUseCase) Create(ctx context.Context, username, email string) (*User, error) {
	user := &User{
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID 根据ID获取用户。
func (uc *UserUseCase) GetByID(ctx context.Context, id uint) (*User, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取用户列表。
func (uc *UserUseCase) List(ctx context.Context, page, size int) ([]*User, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除用户。
func (uc *UserUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}