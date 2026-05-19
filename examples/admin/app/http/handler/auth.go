// Package handler HTTP 处理器。
//
// 使用 gorp.Context 抽象接口实现，与模板生成代码风格一致。
// 中间件使用 Context 的 Get/Set/Abort/AbortWithJSON/Next 方法。
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"admin/internal/service"

	gorp "github.com/ngq/gorp"
)

// AuthHandler 认证处理器。
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login 登录。
func (h *AuthHandler) Login(c gorp.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	token, err := h.svc.Login(c, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"token": token})
}

// Logout 登出。
func (h *AuthHandler) Logout(c gorp.Context) {
	c.JSON(http.StatusOK, map[string]any{"message": "logout success"})
}

// GetCurrentUser 获取当前用户。
func (h *AuthHandler) GetCurrentUser(c gorp.Context) {
	userID := c.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	user, err := h.svc.GetCurrentUser(c, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UserHandler 用户处理器。
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// List 用户列表。
func (h *UserHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.svc.List(c, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"data":  users,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Create 创建用户。
func (h *UserHandler) Create(c gorp.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	user, err := h.svc.Create(c, req.Username, req.Password, req.Nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Update 更新用户。
func (h *UserHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	if err := h.svc.Update(c, uint(id), req.Nickname, req.Email, req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"message": "update success"})
}

// Delete 删除用户。
func (h *UserHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignRoles 分配角色。
func (h *UserHandler) AssignRoles(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	var req struct {
		RoleIDs []uint `json:"role_ids"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	if err := h.svc.AssignRoles(c, uint(id), req.RoleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]any{"message": "assign success"})
}

// RoleHandler 角色处理器。
type RoleHandler struct {
	svc *service.RoleService
}

// NewRoleHandler 创建角色处理器。
func NewRoleHandler(svc *service.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// List 角色列表。
func (h *RoleHandler) List(c gorp.Context) {
	roles, err := h.svc.List(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// Create 创建角色。
func (h *RoleHandler) Create(c gorp.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
		Desc string `json:"desc"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	role, err := h.svc.Create(c, req.Name, req.Code, req.Desc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// Delete 删除角色。
func (h *RoleHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AuthMiddleware JWT 认证中间件。
//
// 使用 gorp.Context 接口的 Get/Set/AbortWithJSON/Next 方法，
// 与模板生成的中间件风格一致。
func AuthMiddleware(svc *service.AuthService) gorp.Middleware {
	return func(next gorp.Handler) gorp.Handler {
		return func(c gorp.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.AbortWithJSON(http.StatusUnauthorized, map[string]any{"error": "missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithJSON(http.StatusUnauthorized, map[string]any{"error": "invalid authorization format"})
				return
			}

			claims, err := svc.ValidateToken(parts[1])
			if err != nil {
				c.AbortWithJSON(http.StatusUnauthorized, map[string]any{"error": "invalid token"})
				return
			}

			// 使用 Set 存储用户信息，供后续 handler 使用
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)

			// 继续执行下一个 handler
			c.Next()
		}
	}
}
