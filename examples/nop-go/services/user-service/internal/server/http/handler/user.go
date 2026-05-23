// Package handler 定义用户服务的 HTTP 请求处理器。
// 本文件包含用户相关的接口处理器，使用 gorp.Context 抽象处理 HTTP 请求。
// 每个 handler 方法负责：解析请求参数、调用服务层、返回统一响应。
package handler

import (
	"errors"
	"strconv"

	"nop-go/services/user-service/internal/biz"
	"nop-go/services/user-service/internal/server/http/request"
	"nop-go/services/user-service/internal/service"

	gorp "github.com/ngq/gorp"
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
// 用户 CRUD Handler
// ============================================================

// Register 用户注册。
// POST /api/v1/auth/register
func (h *UserHandler) Register(c gorp.Context) {
	var req request.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	user := &biz.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Nickname: req.Nickname,
	}

	dto, err := h.user.CreateUser(c.Context(), user)
	if err != nil {
		if errors.Is(err, biz.ErrUserAlreadyExists) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// Login 用户登录。
// POST /api/v1/auth/login
func (h *UserHandler) Login(c gorp.Context) {
	var req request.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// TODO: 实现实际的登录逻辑（密码校验、JWT 签发等）
	dto, err := h.user.GetUser(c.Context(), 0) // 占位
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// CreateUser 创建用户（管理端）。
// POST /api/v1/users
func (h *UserHandler) CreateUser(c gorp.Context) {
	var req request.CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	user := &biz.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Status:   req.Status,
	}

	dto, err := h.user.CreateUser(c.Context(), user)
	if err != nil {
		if errors.Is(err, biz.ErrUserAlreadyExists) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// GetUser 获取单个用户。
// GET /api/v1/users/:id
func (h *UserHandler) GetUser(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	dto, err := h.user.GetUser(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// UpdateUser 更新用户信息。
// PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	var req request.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	user := &biz.User{
		ID:       uint(id),
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Status:   req.Status,
	}

	dto, err := h.user.UpdateUser(c.Context(), user)
	if err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// DeleteUser 删除用户。
// DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	if err := h.user.DeleteUser(c.Context(), uint(id)); err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "删除成功"})
}

// ListUsers 获取用户列表（分页）。
// GET /api/v1/users
func (h *UserHandler) ListUsers(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	items, total, err := h.user.ListUsers(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ============================================================
// 地址相关 Handler
// ============================================================

// ListAddresses 地址列表。
// GET /api/v1/users/:id/addresses
func (h *UserHandler) ListAddresses(c gorp.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	items, err := h.user.ListAddresses(c.Context(), uint(userID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, items)
}

// CreateAddress 添加地址。
// POST /api/v1/users/:id/addresses
func (h *UserHandler) CreateAddress(c gorp.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	var req request.CreateAddressRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	addr := &biz.Address{
		UserID:        uint(userID),
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		IsDefault:     req.IsDefault,
	}

	dto, err := h.user.CreateAddress(c.Context(), addr)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// UpdateAddress 编辑地址。
// PUT /api/v1/users/addresses/:addrId
func (h *UserHandler) UpdateAddress(c gorp.Context) {
	addrID, err := strconv.ParseUint(c.Param("addrId"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的地址 ID")
		return
	}

	var req request.UpdateAddressRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	addr := &biz.Address{
		ID:            uint(addrID),
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		IsDefault:     req.IsDefault,
	}

	dto, err := h.user.UpdateAddress(c.Context(), addr)
	if err != nil {
		if errors.Is(err, biz.ErrAddressNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// DeleteAddress 删除地址。
// DELETE /api/v1/users/addresses/:addrId
func (h *UserHandler) DeleteAddress(c gorp.Context) {
	addrID, err := strconv.ParseUint(c.Param("addrId"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的地址 ID")
		return
	}

	if err := h.user.DeleteAddress(c.Context(), uint(addrID)); err != nil {
		if errors.Is(err, biz.ErrAddressNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "删除成功"})
}

// ============================================================
// 外部关联 Handler
// ============================================================

// ListExternalAssociations 获取用户外部关联列表。
// GET /api/v1/users/:id/external-associations
func (h *UserHandler) ListExternalAssociations(c gorp.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	items, err := h.user.ListExternalAssociations(c.Context(), uint(userID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, items)
}

// CreateExternalAssociation 创建外部关联。
// POST /api/v1/users/:id/external-associations
func (h *UserHandler) CreateExternalAssociation(c gorp.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	var req request.CreateExternalAssociationRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	ea := &biz.ExternalAssociation{
		UserID:       uint(userID),
		Platform:     req.Platform,
		ExternalID:   req.ExternalID,
		ExternalData: req.ExternalData,
	}

	dto, err := h.user.CreateExternalAssociation(c.Context(), ea)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// DeleteExternalAssociation 删除外部关联。
// DELETE /api/v1/users/external-associations/:eaId
func (h *UserHandler) DeleteExternalAssociation(c gorp.Context) {
	eaID, err := strconv.ParseUint(c.Param("eaId"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的外部关联 ID")
		return
	}

	if err := h.user.DeleteExternalAssociation(c.Context(), uint(eaID)); err != nil {
		if errors.Is(err, biz.ErrExternalAssociationNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "删除成功"})
}

// ============================================================
// 可下载产品 Handler
// ============================================================

// ListDownloadableProducts 获取用户可下载产品列表。
// GET /api/v1/users/:id/downloadable-products
func (h *UserHandler) ListDownloadableProducts(c gorp.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	items, err := h.user.ListDownloadableProducts(c.Context(), uint(userID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, items)
}

// CreateDownloadableProduct 创建可下载产品。
// POST /api/v1/users/:id/downloadable-products
func (h *UserHandler) CreateDownloadableProduct(c gorp.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的用户 ID")
		return
	}

	var req request.CreateDownloadableProductRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	dp := &biz.DownloadableProduct{
		UserID:      uint(userID),
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		DownloadURL: req.DownloadURL,
		ExpireAt:    req.ExpireAt,
	}

	dto, err := h.user.CreateDownloadableProduct(c.Context(), dp)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, dto)
}

// DeleteDownloadableProduct 删除可下载产品。
// DELETE /api/v1/users/downloadable-products/:dpId
func (h *UserHandler) DeleteDownloadableProduct(c gorp.Context) {
	dpID, err := strconv.ParseUint(c.Param("dpId"), 10, 32)
	if err != nil {
		gorp.BadRequest(c, "无效的产品 ID")
		return
	}

	if err := h.user.DeleteDownloadableProduct(c.Request().Context(), uint(dpID)); err != nil {
		if errors.Is(err, biz.ErrDownloadableProductNotFound) {
			gorp.BadRequest(c, err.Error())
			return
		}
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"message": "删除成功"})
}