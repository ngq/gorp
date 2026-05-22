package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/security/internal/server/http/request"
	"nop-go/services/security/internal/server/http/response"
	"nop-go/services/security/internal/service"
)

// PermissionHandler 权限管理 HTTP 处理器。
type PermissionHandler struct {
	permission *service.PermissionService
}

// NewPermissionHandler 创建权限管理处理器。
func NewPermissionHandler(permission *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permission: permission}
}

// List 权限列表。
func (h *PermissionHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.permission.List(c, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Permission, len(items))
	for i, item := range items {
		respItems[i] = response.Permission{
			ID:            item.ID,
			Name:          item.Name,
			SystemName:    item.SystemName,
			Category:      item.Category,
			DisplayOrder:  item.DisplayOrder,
		}
	}

	gorp.Success(c, response.PermissionList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建权限。
func (h *PermissionHandler) Create(c gorp.Context) {
	var req request.CreatePermission
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	perm, err := h.permission.Create(c, service.CreatePermissionRequest{
		Name:         req.Name,
		SystemName:   req.SystemName,
		Category:     req.Category,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, perm)
}

// Update 更新权限。
func (h *PermissionHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdatePermission
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	perm, err := h.permission.Update(c, uint(id), service.UpdatePermissionRequest{
		Name:         req.Name,
		SystemName:   req.SystemName,
		Category:     req.Category,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, perm)
}

// Delete 删除权限。
func (h *PermissionHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.permission.Delete(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ACLHandler ACL管理 HTTP 处理器。
type ACLHandler struct {
	acl *service.ACLService
}

// NewACLHandler 创建ACL管理处理器。
func NewACLHandler(acl *service.ACLService) *ACLHandler {
	return &ACLHandler{acl: acl}
}

// List ACL记录列表。
func (h *ACLHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.acl.List(c, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.ACLRecord, len(items))
	for i, item := range items {
		respItems[i] = response.ACLRecord{
			ID:             item.ID,
			UserID:         item.UserID,
			PermissionID:   item.PermissionID,
			PermissionName: item.PermissionName,
		}
	}

	gorp.Success(c, response.ACLRecordList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建ACL记录。
func (h *ACLHandler) Create(c gorp.Context) {
	var req request.CreateACLRecord
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	record, err := h.acl.Create(c, service.CreateACLRequest{
		UserID:       req.UserID,
		PermissionID: req.PermissionID,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, record)
}

// Delete 删除ACL记录。
func (h *ACLHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.acl.Delete(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}