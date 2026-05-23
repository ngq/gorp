// Package handler 安全模块处理器 —— 处理权限与ACL的HTTP请求
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/server/http/request"
	"nop-go/services/admin-service/internal/service"
)

// ==================== 权限处理器 ====================

// PermissionHandler 权限处理器 —— 处理权限管理相关的HTTP请求
type PermissionHandler struct {
	svc *service.SecurityService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(svc *service.SecurityService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

// List 获取权限列表
func (h *PermissionHandler) List(c gorp.Context) {
	module := c.DefaultQuery("module", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	perms, total, err := h.svc.ListPermissions(c.Context(), module, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": perms,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取权限详情
func (h *PermissionHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	perm, err := h.svc.GetPermission(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, perm)
}

// Create 创建权限
func (h *PermissionHandler) Create(c gorp.Context) {
	var req request.CreatePermissionRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	p := &biz.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Module:      req.Module,
		Action:      req.Action,
	}
	if err := h.svc.CreatePermission(c.Context(), p); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, 201, p)
}

// Update 更新权限
func (h *PermissionHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdatePermissionRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	p := &biz.Permission{
		ID:          uint(id),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Module:      req.Module,
		Action:      req.Action,
	}
	if err := h.svc.UpdatePermission(c.Context(), p); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, p)
}

// Delete 删除权限
func (h *PermissionHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.svc.DeletePermission(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ==================== ACL处理器 ====================

// ACLHandler ACL处理器 —— 处理访问控制列表相关的HTTP请求
type ACLHandler struct {
	svc *service.SecurityService
}

// NewACLHandler 创建ACL处理器
func NewACLHandler(svc *service.SecurityService) *ACLHandler {
	return &ACLHandler{svc: svc}
}

// List 获取ACL规则列表
func (h *ACLHandler) List(c gorp.Context) {
	roleID, _ := strconv.ParseUint(c.DefaultQuery("role_id", "0"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	acls, total, err := h.svc.ListACLs(c.Context(), uint(roleID), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": acls,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取ACL规则详情
func (h *ACLHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	acl, err := h.svc.GetACL(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, acl)
}

// Create 创建ACL规则
func (h *ACLHandler) Create(c gorp.Context) {
	var req request.CreateACLRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	acl := &biz.ACL{
		RoleID:   req.RoleID,
		Resource: req.Resource,
		Action:   req.Action,
		Effect:   req.Effect,
	}
	if err := h.svc.CreateACL(c.Context(), acl); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, 201, acl)
}

// Update 更新ACL规则
func (h *ACLHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateACLRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	acl := &biz.ACL{
		ID:       uint(id),
		RoleID:   req.RoleID,
		Resource: req.Resource,
		Action:   req.Action,
		Effect:   req.Effect,
	}
	if err := h.svc.UpdateACL(c.Context(), acl); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, acl)
}

// Delete 删除ACL规则
func (h *ACLHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.svc.DeleteACL(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}