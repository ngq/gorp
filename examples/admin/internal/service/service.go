// Package service 负责业务编排。
package service

import (
	"context"

	"admin/internal/biz"
)

// Services 聚合应用服务。
type Services struct {
	Demo *DemoService
	Auth *AuthService
	User *UserService
	Role *RoleService
}

// NewServices 创建应用服务集合。
func NewServices(biz *biz.Biz) *Services {
	return &Services{
		Demo: &DemoService{uc: biz.Demo},
		Auth: &AuthService{uc: biz.Auth},
		User: &UserService{uc: biz.User},
		Role: &RoleService{uc: biz.Role},
	}
}

// DemoService 编排 Demo 用例。
type DemoService struct {
	uc *biz.DemoUseCase
}

func (s *DemoService) Create(ctx context.Context, name string) (*biz.Demo, error) {
	return s.uc.Create(ctx, name)
}

func (s *DemoService) GetByID(ctx context.Context, id uint) (*biz.Demo, error) {
	return s.uc.GetByID(ctx, id)
}

func (s *DemoService) List(ctx context.Context, page, pageSize int) ([]*biz.Demo, int64, error) {
	return s.uc.List(ctx, page, pageSize)
}

func (s *DemoService) Update(ctx context.Context, id uint, name string) (*biz.Demo, error) {
	return s.uc.Update(ctx, id, name)
}

func (s *DemoService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// AuthService 认证服务。
type AuthService struct {
	uc *biz.AuthUseCase
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	return s.uc.Login(ctx, username, password)
}

func (s *AuthService) ValidateToken(token string) (*biz.Claims, error) {
	return s.uc.ValidateToken(token)
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uint) (*biz.User, error) {
	return s.uc.GetCurrentUser(ctx, userID)
}

// UserService 用户服务。
type UserService struct {
	uc *biz.UserUseCase
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*biz.User, error) {
	return s.uc.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, page, pageSize int) ([]*biz.User, int64, error) {
	return s.uc.List(ctx, page, pageSize)
}

func (s *UserService) Create(ctx context.Context, username, password, nickname string) (*biz.User, error) {
	return s.uc.Create(ctx, username, password, nickname)
}

func (s *UserService) Update(ctx context.Context, id uint, nickname, email, phone string) error {
	return s.uc.Update(ctx, id, nickname, email, phone)
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

func (s *UserService) AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return s.uc.AssignRoles(ctx, userID, roleIDs)
}

// RoleService 角色服务。
type RoleService struct {
	uc *biz.RoleUseCase
}

func (s *RoleService) GetByID(ctx context.Context, id uint) (*biz.Role, error) {
	return s.uc.GetByID(ctx, id)
}

func (s *RoleService) List(ctx context.Context) ([]*biz.Role, error) {
	return s.uc.List(ctx)
}

func (s *RoleService) Create(ctx context.Context, name, code, desc string) (*biz.Role, error) {
	return s.uc.Create(ctx, name, code, desc)
}

func (s *RoleService) Update(ctx context.Context, id uint, name, desc string) error {
	return s.uc.Update(ctx, id, name, desc)
}

func (s *RoleService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
