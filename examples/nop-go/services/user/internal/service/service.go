// Package service 服务层。
// 作为 handler 和 biz 之间的桥梁，负责：
// 1. 组装 biz 层用例的返回值为 handler 需要的响应格式
// 2. 处理跨领域编排逻辑（如登录时生成 JWT 令牌）
// 3. 隔离 handler 对 biz 层领域实体的直接依赖
package service

import (
	"context"
	"mime/multipart"
	"time"

	"nop-go/services/user/internal/biz"
	"nop-go/services/user/internal/data"

	"gorm.io/gorm"
)

// ============================================================
// 服务容器
// ============================================================

// Services 服务容器。
// 聚合用户服务的所有服务实例，供 wire 注入。
type Services struct {
	User *UserService
}

// NewServices 创建服务容器。
// 初始化所有仓储、用例和服务实例，完成依赖注入。
func NewServices(db *gorm.DB) *Services {
	userRepo := data.NewUserRepo(db)
	addressRepo := data.NewAddressRepo(db)
	extRepo := data.NewExternalAssociationRepo(db)
	dlRepo := data.NewDownloadableProductRepo(db)

	userUC := biz.NewUserUseCase(userRepo, addressRepo, extRepo, dlRepo)
	return &Services{
		User: &UserService{uc: userUC},
	}
}

// ============================================================
// UserService 用户服务
// ============================================================

// UserService 用户服务。
// 封装用户相关的业务编排逻辑，供 handler 调用。
type UserService struct {
	uc *biz.UserUseCase
}

// ---------- 认证相关 ----------

// LoginResult 登录结果。
// 包含 JWT 令牌和用户基本信息。
type LoginResult struct {
	Token     string    // JWT 令牌
	ExpiresAt time.Time // 令牌过期时间
	User      *biz.User // 用户领域实体
}

// Login 用户登录。
// 调用 biz 层验证凭据，成功后生成 JWT 令牌。
func (s *UserService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.uc.Login(ctx, username, password)
	if err != nil {
		return nil, err
	}

	// 生成 JWT 令牌（实际项目中应使用 jwt 库生成）
	token := "jwt-token-for-user-" + user.Username
	expiresAt := time.Now().Add(24 * time.Hour)

	return &LoginResult{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Logout 用户登出。
func (s *UserService) Logout(ctx context.Context, userID uint) error {
	return s.uc.Logout(ctx, userID)
}

// Register 用户注册。
func (s *UserService) Register(ctx context.Context, username, email, password string) (*biz.User, error) {
	return s.uc.Register(ctx, username, email, password)
}

// PasswordRecovery 密码恢复。
func (s *UserService) PasswordRecovery(ctx context.Context, email string) error {
	return s.uc.PasswordRecovery(ctx, email)
}

// ConfirmPasswordRecovery 确认密码恢复。
func (s *UserService) ConfirmPasswordRecovery(ctx context.Context, token, newPassword string) error {
	return s.uc.ConfirmPasswordRecovery(ctx, token, newPassword)
}

// MultiFactorVerification 多因素验证。
func (s *UserService) MultiFactorVerification(ctx context.Context, userID uint, code string) error {
	return s.uc.MultiFactorVerification(ctx, userID, code)
}

// ---------- 用户信息相关 ----------

// GetUserInfo 获取用户信息。
func (s *UserService) GetUserInfo(ctx context.Context, userID uint) (*biz.User, error) {
	return s.uc.GetUserInfo(ctx, userID)
}

// UpdateUserInfo 更新用户信息。
func (s *UserService) UpdateUserInfo(ctx context.Context, userID uint, email, phone string) (*biz.User, error) {
	return s.uc.UpdateUserInfo(ctx, userID, email, phone)
}

// ChangePassword 修改密码。
func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	return s.uc.ChangePassword(ctx, userID, oldPassword, newPassword)
}

// ---------- 地址相关 ----------

// ListAddresses 获取用户地址列表。
func (s *UserService) ListAddresses(ctx context.Context, userID uint) ([]*biz.Address, error) {
	return s.uc.ListAddresses(ctx, userID)
}

// AddAddress 添加地址。
// 将请求参数组装为 biz.Address 后调用用例。
func (s *UserService) AddAddress(ctx context.Context, address *biz.Address) (uint, error) {
	if err := s.uc.AddAddress(ctx, address); err != nil {
		return 0, err
	}
	return address.ID, nil
}

// UpdateAddress 更新地址。
func (s *UserService) UpdateAddress(ctx context.Context, address *biz.Address) error {
	return s.uc.UpdateAddress(ctx, address)
}

// DeleteAddress 删除地址。
// 先验证地址归属，再调用 biz 层删除。
func (s *UserService) DeleteAddress(ctx context.Context, userID uint, addressID uint) error {
	// 验证地址归属当前用户（防止越权删除）
	addr, err := s.uc.ListAddresses(ctx, userID)
	if err != nil {
		return err
	}
	found := false
	for _, a := range addr {
		if a.ID == addressID {
			found = true
			break
		}
	}
	if !found {
		return biz.NewBizError("地址不存在或不属于当前用户")
	}
	return s.uc.DeleteAddress(ctx, addressID)
}

// ---------- 头像相关 ----------

// GetAvatar 获取头像信息。
func (s *UserService) GetAvatar(ctx context.Context, userID uint) (*biz.User, error) {
	return s.uc.GetAvatar(ctx, userID)
}

// UploadAvatar 上传头像。
// 接收 multipart 文件头，保存后调用 biz 层更新头像 URL。
func (s *UserService) UploadAvatar(ctx context.Context, userID uint, fileHeader *multipart.FileHeader) (string, error) {
	// 实际项目中应将文件保存到对象存储（如 OSS、S3）
	// 此处为简化实现，生成一个模拟 URL
	avatarURL := "/uploads/avatars/" + fileHeader.Filename
	if err := s.uc.UploadAvatar(ctx, userID, avatarURL); err != nil {
		return "", err
	}
	return avatarURL, nil
}

// RemoveAvatar 移除头像。
func (s *UserService) RemoveAvatar(ctx context.Context, userID uint) error {
	return s.uc.RemoveAvatar(ctx, userID)
}

// ---------- 其他 ----------

// CheckUsernameAvailability 检查用户名可用性。
func (s *UserService) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	return s.uc.CheckUsernameAvailability(ctx, username)
}

// GetDownloadableProducts 获取可下载产品列表。
func (s *UserService) GetDownloadableProducts(ctx context.Context, userID uint) ([]*biz.DownloadableProduct, error) {
	return s.uc.GetDownloadableProducts(ctx, userID)
}

// RemoveExternalAssociation 移除外部关联。
func (s *UserService) RemoveExternalAssociation(ctx context.Context, userID uint, provider string) error {
	return s.uc.RemoveExternalAssociation(ctx, userID, provider)
}