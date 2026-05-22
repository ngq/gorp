// Package handler 定义 HTTP 请求处理器。
// 每个 handler 方法对应一个路由端点，负责：
// 1. 解析和校验请求参数
// 2. 调用业务逻辑层（biz）执行核心逻辑
// 3. 使用 gorp.Success/gorp.Error/gorp.BadRequest 统一响应
package handler

import (
	"errors"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/user/internal/biz"
	"nop-go/services/user/internal/server/http/request"
	"nop-go/services/user/internal/server/http/response"
	"nop-go/services/user/internal/service"
)

// UserHandler 用户服务 HTTP 处理器。
// 聚合认证、用户信息、地址、头像等所有用户相关端点的处理方法。
type UserHandler struct {
	user *service.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(user *service.UserService) *UserHandler {
	return &UserHandler{user: user}
}

// ============================================================
// 认证相关 Handler
// ============================================================

// Login 登录。
// POST /api/v1/auth/login
// 接收用户名和密码，验证成功后返回 JWT 令牌。
func (h *UserHandler) Login(c gorp.Context) {
	var req request.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	result, err := h.user.Login(c, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidCredentials) || errors.Is(err, biz.ErrUserInactive) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.LoginResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		UserInfo:  toUserInfo(result.User),
	})
}

// Logout 登出。
// POST /api/v1/auth/logout
// 无状态 JWT 模式下，客户端删除 token 即可。
func (h *UserHandler) Logout(c gorp.Context) {
	// 从上下文中获取当前用户 ID（由认证中间件注入）
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	if err := h.user.Logout(c, userID); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.LogoutResponse{Message: "已成功登出"})
}

// Register 注册。
// POST /api/v1/auth/register
// 创建新用户并返回基本信息。
func (h *UserHandler) Register(c gorp.Context) {
	var req request.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	user, err := h.user.Register(c, req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrUsernameTaken) || errors.Is(err, biz.ErrEmailTaken) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// PasswordRecovery 密码恢复。
// POST /api/v1/auth/password-recovery
// 发送密码恢复链接到用户邮箱。
func (h *UserHandler) PasswordRecovery(c gorp.Context) {
	var req request.PasswordRecoveryRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if err := h.user.PasswordRecovery(c, req.Email); err != nil {
		gorp.Error(c, err)
		return
	}

	// 安全考虑：无论邮箱是否存在都返回相同提示
	gorp.Success(c, response.PasswordRecoveryResponse{
		Message: "如果该邮箱已注册，恢复链接已发送",
	})
}

// ConfirmPasswordRecovery 确认密码恢复。
// POST /api/v1/auth/password-recovery/confirm
// 验证恢复令牌并重置密码。
func (h *UserHandler) ConfirmPasswordRecovery(c gorp.Context) {
	var req request.ConfirmPasswordRecoveryRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if err := h.user.ConfirmPasswordRecovery(c, req.Token, req.NewPassword); err != nil {
		if errors.Is(err, biz.ErrInvalidRecoveryToken) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ConfirmPasswordRecoveryResponse{
		Message: "密码已重置",
	})
}

// MultiFactorVerification 多因素验证。
// GET /api/v1/auth/multi-factor-verification
// 验证用户提交的 MFA 验证码。
func (h *UserHandler) MultiFactorVerification(c gorp.Context) {
	var req request.MultiFactorVerificationRequest
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	if err := h.user.MultiFactorVerification(c, userID, req.Code); err != nil {
		if errors.Is(err, biz.ErrMFANotEnabled) || errors.Is(err, biz.ErrInvalidMFACode) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.MultiFactorVerificationResponse{Verified: true})
}

// ============================================================
// 用户信息相关 Handler
// ============================================================

// GetUserInfo 获取客户信息。
// GET /api/v1/users/info
// 返回当前登录用户的详细信息。
func (h *UserHandler) GetUserInfo(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	user, err := h.user.GetUserInfo(c, userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, toUserInfo(user))
}

// UpdateUserInfo 更新客户信息。
// PUT /api/v1/users/info
// 更新当前登录用户的邮箱、手机等信息。
func (h *UserHandler) UpdateUserInfo(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	var req request.UpdateUserInfoRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	user, err := h.user.UpdateUserInfo(c, userID, req.Email, req.Phone)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.UpdateUserInfoResponse{UserInfo: toUserInfo(user)})
}

// ChangePassword 修改密码。
// PUT /api/v1/users/password
// 验证旧密码后更新为新密码。
func (h *UserHandler) ChangePassword(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	var req request.ChangePasswordRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if err := h.user.ChangePassword(c, userID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, biz.ErrInvalidOldPassword) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ChangePasswordResponse{Message: "密码修改成功"})
}

// ============================================================
// 地址相关 Handler
// ============================================================

// ListAddresses 地址列表。
// GET /api/v1/users/addresses
// 返回当前用户的所有地址。
func (h *UserHandler) ListAddresses(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	addresses, err := h.user.ListAddresses(c, userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	items := make([]response.Address, len(addresses))
	for i, addr := range addresses {
		items[i] = toAddress(addr)
	}

	gorp.Success(c, response.AddressListResponse{Items: items})
}

// AddAddress 添加地址。
// POST /api/v1/users/addresses
// 为当前用户添加一个新的收货地址。
func (h *UserHandler) AddAddress(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	var req request.AddAddressRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 将请求结构体转换为 biz 层地址实体
	address := &biz.Address{
		UserID:          userID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		Fax:             req.Fax,
		Company:         req.Company,
		CountryID:       req.CountryID,
		StateProvinceID: req.StateProvinceID,
		City:            req.City,
		Address1:        req.Address1,
		Address2:        req.Address2,
		ZipPostalCode:   req.ZipPostalCode,
		IsDefault:       req.IsDefault,
	}

	id, err := h.user.AddAddress(c, address)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.AddAddressResponse{ID: id})
}

// UpdateAddress 编辑地址。
// PUT /api/v1/users/addresses/:id
// 更新指定 ID 的地址信息。
func (h *UserHandler) UpdateAddress(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	addressID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的地址 ID")
		return
	}

	var req request.UpdateAddressRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 将请求结构体转换为 biz 层地址实体
	address := &biz.Address{
		ID:              uint(addressID),
		UserID:          userID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		Fax:             req.Fax,
		Company:         req.Company,
		CountryID:       req.CountryID,
		StateProvinceID: req.StateProvinceID,
		City:            req.City,
		Address1:        req.Address1,
		Address2:        req.Address2,
		ZipPostalCode:   req.ZipPostalCode,
		IsDefault:       req.IsDefault,
	}

	if err := h.user.UpdateAddress(c, address); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "地址更新成功"})
}

// DeleteAddress 删除地址。
// DELETE /api/v1/users/addresses/:id
// 删除指定 ID 的地址。
func (h *UserHandler) DeleteAddress(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	addressID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的地址 ID")
		return
	}

	if err := h.user.DeleteAddress(c, userID, uint(addressID)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "地址删除成功"})
}

// ============================================================
// 头像相关 Handler
// ============================================================

// GetAvatar 头像信息。
// GET /api/v1/users/avatar
// 返回当前用户的头像 URL。
func (h *UserHandler) GetAvatar(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	user, err := h.user.GetAvatar(c, userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.AvatarResponse{AvatarURL: user.AvatarURL})
}

// UploadAvatar 上传头像。
// POST /api/v1/users/avatar/upload
// 接收 multipart/form-data 上传的图片文件，保存后更新用户头像 URL。
func (h *UserHandler) UploadAvatar(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	// 从 multipart form 中获取上传文件头
	// http.Request.FormFile 返回 (multipart.File, *multipart.FileHeader, error)
	file, fileHeader, err := c.Request().FormFile("file")
	if err != nil {
		gorp.BadRequest(c, "请上传头像文件")
		return
	}
	defer file.Close()

	avatarURL, err := h.user.UploadAvatar(c, userID, fileHeader)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.UploadAvatarResponse{AvatarURL: avatarURL})
}

// RemoveAvatar 移除头像。
// DELETE /api/v1/users/avatar
// 删除当前用户的头像。
func (h *UserHandler) RemoveAvatar(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	if err := h.user.RemoveAvatar(c, userID); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "头像已移除"})
}

// ============================================================
// 其他 Handler
// ============================================================

// CheckUsername 检查用户名可用性。
// POST /api/v1/users/check-username
// 检查指定用户名是否可被注册。
func (h *UserHandler) CheckUsername(c gorp.Context) {
	var req request.CheckUsernameRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	available, err := h.user.CheckUsernameAvailability(c, req.Username)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.CheckUsernameResponse{Available: available})
}

// GetDownloadableProducts 可下载产品。
// GET /api/v1/users/downloadable-products
// 返回当前用户已购买的可下载产品列表。
func (h *UserHandler) GetDownloadableProducts(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	products, err := h.user.GetDownloadableProducts(c, userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	items := make([]response.DownloadableProductResponse, len(products))
	for i, p := range products {
		items[i] = response.DownloadableProductResponse{
			ID:            p.ID,
			ProductID:     p.ProductID,
			DownloadCount: p.DownloadCount,
			MaxDownloads:  p.MaxDownloads,
			IsActivated:   p.IsActivated,
			ExpiresAt:     p.ExpiresAt,
		}
	}

	gorp.Success(c, response.DownloadableProductListResponse{Items: items})
}

// RemoveExternalAssociation 移除外部关联。
// POST /api/v1/users/external-association/remove
// 解除当前用户与指定第三方平台的绑定关系。
func (h *UserHandler) RemoveExternalAssociation(c gorp.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		gorp.BadRequest(c, "未登录")
		return
	}

	var req request.RemoveExternalAssociationRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if err := h.user.RemoveExternalAssociation(c, userID, req.Provider); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.RemoveExternalAssociationResponse{
		Message: "外部关联已移除",
	})
}

// ============================================================
// 辅助函数
// ============================================================

// getUserIDFromContext 从上下文中获取当前登录用户 ID。
// 实际项目中由认证中间件将用户 ID 注入上下文。
func getUserIDFromContext(c gorp.Context) uint {
	// 尝试从上下文中获取 "user_id"
	if v := c.Get("user_id"); v != nil {
		if id, ok := v.(uint); ok {
			return id
		}
		// 兼容 float64（JSON 解析默认类型）
		if id, ok := v.(float64); ok {
			return uint(id)
		}
	}
	return 0
}

// toUserInfo 将 biz.User 转换为 response.UserInfo。
// 集中转换逻辑，避免在多个 handler 中重复。
func toUserInfo(user *biz.User) response.UserInfo {
	return response.UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Phone:          user.Phone,
		Active:         user.Active,
		AvatarURL:      user.AvatarURL,
		MFAEnabled:     user.MFAEnabled,
		LastLoginAt:    user.LastLoginAt,
		LastActivityAt: user.LastActivityAt,
		CreatedAt:      user.CreatedAt,
	}
}

// toAddress 将 biz.Address 转换为 response.Address。
func toAddress(addr *biz.Address) response.Address {
	return response.Address{
		ID:              addr.ID,
		FirstName:       addr.FirstName,
		LastName:        addr.LastName,
		Email:           addr.Email,
		Phone:           addr.Phone,
		Fax:             addr.Fax,
		Company:         addr.Company,
		CountryID:       addr.CountryID,
		StateProvinceID: addr.StateProvinceID,
		City:            addr.City,
		Address1:        addr.Address1,
		Address2:        addr.Address2,
		ZipPostalCode:   addr.ZipPostalCode,
		IsDefault:       addr.IsDefault,
		CreatedAt:       addr.CreatedAt,
	}
}