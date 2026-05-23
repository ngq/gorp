package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/content-service/internal/server/http/request"
	"nop-go/services/content-service/internal/server/http/response"
	"nop-go/services/content-service/internal/service"
)

// LocalizationHandler 本地化服务 HTTP 处理器
type LocalizationHandler struct {
	loc *service.LocalizationService
}

// NewLocalizationHandler 创建本地化服务处理器
func NewLocalizationHandler(loc *service.LocalizationService) *LocalizationHandler {
	return &LocalizationHandler{loc: loc}
}

// ==================== 语言 ====================

// ListLanguages 语言列表
func (h *LocalizationHandler) ListLanguages(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.loc.ListLanguages(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.LanguageList{Items: items, Total: total, Page: page, Size: size})
}

// CreateLanguage 创建语言
func (h *LocalizationHandler) CreateLanguage(c gorp.Context) {
	var req request.CreateLanguage
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	lang, err := h.loc.CreateLanguage(c.Context(), service.LanguageRequest{
		Code: req.Code, Name: req.Name, IsDefault: req.IsDefault,
		SortOrder: req.SortOrder, IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, lang)
}

// GetLanguage 获取语言详情
func (h *LocalizationHandler) GetLanguage(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	lang, err := h.loc.GetLanguage(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, lang)
}

// UpdateLanguage 更新语言
func (h *LocalizationHandler) UpdateLanguage(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.UpdateLanguage
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	lang, err := h.loc.UpdateLanguage(c.Context(), id, service.LanguageRequest{
		Code: req.Code, Name: req.Name, IsDefault: req.IsDefault,
		SortOrder: req.SortOrder, IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, lang)
}

// DeleteLanguage 删除语言
func (h *LocalizationHandler) DeleteLanguage(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.loc.DeleteLanguage(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 本地化资源 ====================

// ListLocaleResources 本地化资源列表
func (h *LocalizationHandler) ListLocaleResources(c gorp.Context) {
	langID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的语言ID参数")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.loc.ListLocaleResources(c.Context(), langID, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.LocaleResourceList{Items: items, Total: total, Page: page, Size: size})
}

// CreateLocaleResource 创建本地化资源
func (h *LocalizationHandler) CreateLocaleResource(c gorp.Context) {
	langID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的语言ID参数")
		return
	}
	var req request.CreateLocaleResource
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	res, err := h.loc.CreateLocaleResource(c.Context(), service.LocaleResourceRequest{
		LanguageID: langID, Key: req.Key, Value: req.Value, Module: req.Module,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, res)
}

// GetLocaleResource 获取本地化资源详情
func (h *LocalizationHandler) GetLocaleResource(c gorp.Context) {
	resID, err := strconv.ParseUint(c.Param("resId"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的资源ID参数")
		return
	}
	res, err := h.loc.GetLocaleResource(c.Context(), resID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, res)
}

// UpdateLocaleResource 更新本地化资源
func (h *LocalizationHandler) UpdateLocaleResource(c gorp.Context) {
	resID, err := strconv.ParseUint(c.Param("resId"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的资源ID参数")
		return
	}
	var req request.UpdateLocaleResource
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	res, err := h.loc.UpdateLocaleResource(c.Context(), resID, service.LocaleResourceRequest{
		Key: req.Key, Value: req.Value, Module: req.Module,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, res)
}

// DeleteLocaleResource 删除本地化资源
func (h *LocalizationHandler) DeleteLocaleResource(c gorp.Context) {
	resID, err := strconv.ParseUint(c.Param("resId"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的资源ID参数")
		return
	}
	if err := h.loc.DeleteLocaleResource(c.Context(), resID); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}