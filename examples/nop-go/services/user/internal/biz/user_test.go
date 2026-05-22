// Package biz_test 用户服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试 UserUseCase 的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/user/internal/biz"
)

// ============================================================
// Mock 仓储实现
// ============================================================

// MockUserRepository 用户仓储 mock 实现。
type MockUserRepository struct {
	Users         map[uint]*biz.User
	UsersByUsername map[string]*biz.User
	UsersByEmail  map[string]*biz.User
	NextID        uint
}

// NewMockUserRepository 创建 mock 用户仓储。
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:         make(map[uint]*biz.User),
		UsersByUsername: make(map[string]*biz.User),
		UsersByEmail:  make(map[string]*biz.User),
		NextID:        1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *biz.User) error {
	user.ID = m.NextID
	m.NextID++
	m.Users[user.ID] = user
	m.UsersByUsername[user.Username] = user
	m.UsersByEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*biz.User, error) {
	user, ok := m.Users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	user, ok := m.UsersByUsername[username]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	user, ok := m.UsersByEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *biz.User) error {
	m.Users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uint, passwordHash, passwordSalt string) error {
	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.PasswordHash = passwordHash
	user.PasswordSalt = passwordSalt
	return nil
}

func (m *MockUserRepository) SetPasswordRecoveryToken(ctx context.Context, id uint, token string) error {
	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.PasswordRecoveryToken = token
	return nil
}

func (m *MockUserRepository) GetByPasswordRecoveryToken(ctx context.Context, token string) (*biz.User, error) {
	for _, user := range m.Users {
		if user.PasswordRecoveryToken == token {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	now := time.Now()
	user.LastLoginAt = &now
	return nil
}

func (m *MockUserRepository) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	_, exists := m.UsersByUsername[username]
	return !exists, nil
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, id uint, avatarURL string) error {
	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.AvatarURL = avatarURL
	return nil
}

func (m *MockUserRepository) RemoveAvatar(ctx context.Context, id uint) error {
	user, ok := m.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.AvatarURL = ""
	return nil
}

// MockAddressRepository 地址仓储 mock 实现。
type MockAddressRepository struct {
	Addresses map[uint]*biz.Address
	NextID    uint
}

// NewMockAddressRepository 创建 mock 地址仓储。
func NewMockAddressRepository() *MockAddressRepository {
	return &MockAddressRepository{
		Addresses: make(map[uint]*biz.Address),
		NextID:    1,
	}
}

func (m *MockAddressRepository) ListByUserID(ctx context.Context, userID uint) ([]*biz.Address, error) {
	var result []*biz.Address
	for _, addr := range m.Addresses {
		if addr.UserID == userID {
			result = append(result, addr)
		}
	}
	return result, nil
}

func (m *MockAddressRepository) GetByID(ctx context.Context, id uint) (*biz.Address, error) {
	addr, ok := m.Addresses[id]
	if !ok {
		return nil, errors.New("address not found")
	}
	return addr, nil
}

func (m *MockAddressRepository) Create(ctx context.Context, address *biz.Address) error {
	address.ID = m.NextID
	m.NextID++
	m.Addresses[address.ID] = address
	return nil
}

func (m *MockAddressRepository) Update(ctx context.Context, address *biz.Address) error {
	m.Addresses[address.ID] = address
	return nil
}

func (m *MockAddressRepository) Delete(ctx context.Context, id uint) error {
	delete(m.Addresses, id)
	return nil
}

// MockExternalAssociationRepository 外部关联仓储 mock 实现。
type MockExternalAssociationRepository struct{}

func (m *MockExternalAssociationRepository) Remove(ctx context.Context, userID uint, provider string) error {
	return nil
}

// MockDownloadableProductRepository 可下载产品仓储 mock 实现。
type MockDownloadableProductRepository struct{}

func (m *MockDownloadableProductRepository) ListByUserID(ctx context.Context, userID uint) ([]*biz.DownloadableProduct, error) {
	return []*biz.DownloadableProduct{}, nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestUseCase 创建测试用的 UserUseCase。
func newTestUseCase() (*biz.UserUseCase, *MockUserRepository, *MockAddressRepository) {
	userRepo := NewMockUserRepository()
	addressRepo := NewMockAddressRepository()
	extRepo := &MockExternalAssociationRepository{}
	dlRepo := &MockDownloadableProductRepository{}

	uc := biz.NewUserUseCase(userRepo, addressRepo, extRepo, dlRepo)
	return uc, userRepo, addressRepo
}

// ============================================================
// 登录测试
// ============================================================

func TestLogin_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个用户
	testUser := &biz.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "correctpassword",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 使用用户名登录
	user, err := uc.Login(ctx, "testuser", "correctpassword")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
}

func TestLogin_WithEmail(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "correctpassword",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 使用邮箱登录
	user, err := uc.Login(ctx, "test@example.com", "correctpassword")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "correctpassword",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 错误密码
	user, err := uc.Login(ctx, "testuser", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrInvalidCredentials, err)
	assert.Nil(t, user)
}

func TestLogin_UserNotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	user, err := uc.Login(ctx, "nonexistent", "password")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrInvalidCredentials, err)
	assert.Nil(t, user)
}

func TestLogin_UserInactive(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "inactive",
		Email:        "inactive@example.com",
		PasswordHash: "password",
		Active:       false,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	user, err := uc.Login(ctx, "inactive", "password")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrUserInactive, err)
	assert.Nil(t, user)
}

// ============================================================
// 注册测试
// ============================================================

func TestRegister_Success(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	user, err := uc.Register(ctx, "newuser", "new@example.com", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "new@example.com", user.Email)
	assert.True(t, user.Active)
}

func TestRegister_UsernameTaken(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个用户
	testUser := &biz.User{
		Username:     "existinguser",
		Email:        "existing@example.com",
		PasswordHash: "password",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 尝试使用相同用户名注册
	user, err := uc.Register(ctx, "existinguser", "another@example.com", "password")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrUsernameTaken, err)
	assert.Nil(t, user)
}

func TestRegister_EmailTaken(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "existinguser",
		Email:        "existing@example.com",
		PasswordHash: "password",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 尝试使用相同邮箱注册
	user, err := uc.Register(ctx, "newuser", "existing@example.com", "password")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrEmailTaken, err)
	assert.Nil(t, user)
}

// ============================================================
// 用户信息测试
// ============================================================

func TestGetUserInfo_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "password",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	user, err := uc.GetUserInfo(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testUser.ID, user.ID)
}

func TestGetUserInfo_NotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	user, err := uc.GetUserInfo(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUpdateUserInfo_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "testuser",
		Email:        "test@example.com",
		Phone:        "1234567890",
		PasswordHash: "password",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	user, err := uc.UpdateUserInfo(ctx, testUser.ID, "newemail@example.com", "9876543210")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newemail@example.com", user.Email)
	assert.Equal(t, "9876543210", user.Phone)
}

// ============================================================
// 密码修改测试
// ============================================================

func TestChangePassword_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "testuser",
		PasswordHash: "oldpassword",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.ChangePassword(ctx, testUser.ID, "oldpassword", "newpassword")
	assert.NoError(t, err)

	// 验证密码已更新
	updated, _ := userRepo.GetByID(ctx, testUser.ID)
	assert.Equal(t, "newpassword", updated.PasswordHash)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{
		Username:     "testuser",
		PasswordHash: "oldpassword",
		Active:       true,
	}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.ChangePassword(ctx, testUser.ID, "wrongpassword", "newpassword")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrInvalidOldPassword, err)
}

// ============================================================
// 地址管理测试
// ============================================================

func TestListAddresses_Success(t *testing.T) {
	uc, userRepo, addressRepo := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 添加多个地址
	addr1 := &biz.Address{UserID: testUser.ID, City: "北京", Address1: "朝阳区"}
	addr2 := &biz.Address{UserID: testUser.ID, City: "上海", Address1: "浦东新区"}
	require.NoError(t, addressRepo.Create(ctx, addr1))
	require.NoError(t, addressRepo.Create(ctx, addr2))

	addresses, err := uc.ListAddresses(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, addresses, 2)
}

func TestAddAddress_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	address := &biz.Address{
		UserID:    testUser.ID,
		FirstName: "张",
		LastName:  "三",
		City:      "北京",
		Address1:  "朝阳区xxx",
	}

	err := uc.AddAddress(ctx, address)
	assert.NoError(t, err)
	assert.NotZero(t, address.ID)
}

func TestDeleteAddress_Success(t *testing.T) {
	uc, _, addressRepo := newTestUseCase()
	ctx := context.Background()

	address := &biz.Address{UserID: 1, City: "北京"}
	require.NoError(t, addressRepo.Create(ctx, address))

	err := uc.DeleteAddress(ctx, address.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = addressRepo.GetByID(ctx, address.ID)
	assert.Error(t, err)
}

// ============================================================
// 用户名可用性检查测试
// ============================================================

func TestCheckUsernameAvailability_Available(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	available, err := uc.CheckUsernameAvailability(ctx, "newuser")
	assert.NoError(t, err)
	assert.True(t, available)
}

func TestCheckUsernameAvailability_NotAvailable(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "existinguser", Email: "test@example.com", PasswordHash: "password", Active: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	available, err := uc.CheckUsernameAvailability(ctx, "existinguser")
	assert.NoError(t, err)
	assert.False(t, available)
}

// ============================================================
// 密码恢复测试
// ============================================================

func TestPasswordRecovery_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.PasswordRecovery(ctx, "test@example.com")
	assert.NoError(t, err)

	// 验证 token 已设置
	updated, _ := userRepo.GetByID(ctx, testUser.ID)
	assert.NotEmpty(t, updated.PasswordRecoveryToken)
}

func TestPasswordRecovery_EmailNotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	// 不暴露邮箱是否存在，返回 nil
	err := uc.PasswordRecovery(ctx, "nonexistent@example.com")
	assert.NoError(t, err)
}

func TestConfirmPasswordRecovery_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "oldpassword", Active: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	// 设置恢复令牌
	token := "recovery-token-test@example.com"
	require.NoError(t, userRepo.SetPasswordRecoveryToken(ctx, testUser.ID, token))

	// 确认恢复
	err := uc.ConfirmPasswordRecovery(ctx, token, "newpassword")
	assert.NoError(t, err)

	// 验证密码已更新
	updated, _ := userRepo.GetByID(ctx, testUser.ID)
	assert.Equal(t, "newpassword", updated.PasswordHash)
}

func TestConfirmPasswordRecovery_InvalidToken(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	err := uc.ConfirmPasswordRecovery(ctx, "invalid-token", "newpassword")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrInvalidRecoveryToken, err)
}

// ============================================================
// 头像管理测试
// ============================================================

func TestUploadAvatar_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.UploadAvatar(ctx, testUser.ID, "https://example.com/avatar.jpg")
	assert.NoError(t, err)

	updated, _ := userRepo.GetByID(ctx, testUser.ID)
	assert.Equal(t, "https://example.com/avatar.jpg", updated.AvatarURL)
}

func TestRemoveAvatar_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true, AvatarURL: "https://example.com/avatar.jpg"}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.RemoveAvatar(ctx, testUser.ID)
	assert.NoError(t, err)

	updated, _ := userRepo.GetByID(ctx, testUser.ID)
	assert.Empty(t, updated.AvatarURL)
}

// ============================================================
// 多因素认证测试
// ============================================================

func TestMultiFactorVerification_Success(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true, MFAEnabled: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.MultiFactorVerification(ctx, testUser.ID, "123456")
	assert.NoError(t, err)
}

func TestMultiFactorVerification_NotEnabled(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true, MFAEnabled: false}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.MultiFactorVerification(ctx, testUser.ID, "123456")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrMFANotEnabled, err)
}

func TestMultiFactorVerification_EmptyCode(t *testing.T) {
	uc, userRepo, _ := newTestUseCase()
	ctx := context.Background()

	testUser := &biz.User{Username: "testuser", Email: "test@example.com", PasswordHash: "password", Active: true, MFAEnabled: true}
	require.NoError(t, userRepo.Create(ctx, testUser))

	err := uc.MultiFactorVerification(ctx, testUser.ID, "")
	assert.Error(t, err)
	assert.Equal(t, biz.ErrInvalidMFACode, err)
}
