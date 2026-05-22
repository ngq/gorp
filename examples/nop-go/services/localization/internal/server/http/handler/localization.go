package handler

import (
	"net/http"
	"strconv"
	"time"

	gorp "github.com/ngq/gorp"
	"nop-go/services/localization/internal/server/http/request"
	"nop-go/services/localization/internal/server/http/response"
	"nop-go/services/localization/internal/service"
)

// toLanguageResponse 将 service.LanguageResponse 转换为 response.Language。
//
// service 层返回的时间字段为格式化字符串，response 层需要 time.Time，
// 因此需要解析字符串还原为 time.Time。
func toLanguageResponse(src *service.LanguageResponse) response.Language {
	return response.Language{
		ID:                src.ID,
		Name:              src.Name,
		LanguageCulture:   src.LanguageCulture,
		UniqueSeoCode:     src.UniqueSeoCode,
		FlagImageFileName: src.FlagImageFileName,
		Rtl:               src.Rtl,
		IsActive:          src.IsActive,
		DisplayOrder:      src.DisplayOrder,
		CreatedAt:         parseTime(src.CreatedAt),
		UpdatedAt:         parseTime(src.UpdatedAt),
	}
}

// toLocaleResourceResponse 将 service.LocaleResourceResponse 转换为 response.LocaleResource。
func toLocaleResourceResponse(src *service.LocaleResourceResponse) response.LocaleResource {
	return response.LocaleResource{
		ID:            src.ID,
		LanguageID:    src.LanguageID,
		ResourceName:  src.ResourceName,
		ResourceValue: src.ResourceValue,
		CreatedAt:     parseTime(src.CreatedAt),
		UpdatedAt:     parseTime(src.UpdatedAt),
	}
}

// parseTime 解析 service 层返回的时间字符串为 time.Time。
//
// service 层统一使用 "2006-01-02 15:04:05" 格式输出时间，
// 解析失败时返回零值。
func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// LocalizationHandler 本地化服务 HTTP 处理器。
type LocalizationHandler struct {
	loc *service.LocalizationService
}

// NewLocalizationHandler 创建本地化服务处理器。
func NewLocalizationHandler(loc *service.LocalizationService) *LocalizationHandler {
	return &LocalizationHandler{loc: loc}
}

// ListLanguages 语言列表。
func (h *LocalizationHandler) ListLanguages(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.loc.ListLanguages(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 将 service 层返回的 LanguageResponse 转换为 response.Language
	respItems := make([]response.Language, len(items))
	for i, item := range items {
		respItems[i] = toLanguageResponse(&item)
	}

	gorp.Success(c, response.LanguageList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// CreateLanguage 创建语言。
func (h *LocalizationHandler) CreateLanguage(c gorp.Context) {
	var req request.CreateLanguage
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	lang, err := h.loc.CreateLanguage(c.Context(), service.CreateLanguageRequest{
		Name:            req.Name,
		LanguageCulture: req.LanguageCulture,
		UniqueSeoCode:   req.UniqueSeoCode,
		FlagImageFileName: req.FlagImageFileName,
		Rtl:             req.Rtl,
		IsActive:        req.IsActive,
		DisplayOrder:    req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, lang)
}

// UpdateLanguage 更新语言。
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

	lang, err := h.loc.UpdateLanguage(c.Context(), uint(id), service.UpdateLanguageRequest{
		Name:            req.Name,
		LanguageCulture: req.LanguageCulture,
		UniqueSeoCode:   req.UniqueSeoCode,
		FlagImageFileName: req.FlagImageFileName,
		Rtl:             req.Rtl,
		IsActive:        req.IsActive,
		DisplayOrder:    req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, lang)
}

// DeleteLanguage 删除语言。
func (h *LocalizationHandler) DeleteLanguage(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.loc.DeleteLanguage(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ListResources 本地化资源列表。
func (h *LocalizationHandler) ListResources(c gorp.Context) {
	langID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的语言ID参数")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.loc.ListResources(c.Context(), uint(langID), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 将 service 层返回的 LocaleResourceResponse 转换为 response.LocaleResource
	respItems := make([]response.LocaleResource, len(items))
	for i, item := range items {
		respItems[i] = toLocaleResourceResponse(&item)
	}

	gorp.Success(c, response.LocaleResourceList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// AddResource 添加本地化资源。
func (h *LocalizationHandler) AddResource(c gorp.Context) {
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

	res, err := h.loc.AddResource(c.Context(), uint(langID), service.CreateLocaleResourceRequest{
		ResourceName: req.ResourceName,
		ResourceValue: req.ResourceValue,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, res)
}

// UpdateResource 更新本地化资源。
func (h *LocalizationHandler) UpdateResource(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的资源ID参数")
		return
	}

	var req request.UpdateLocaleResource
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	res, err := h.loc.UpdateResource(c.Context(), uint(id), service.UpdateLocaleResourceRequest{
		ResourceName:  req.ResourceName,
		ResourceValue: req.ResourceValue,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, res)
}

// DeleteResource 删除本地化资源。
func (h *LocalizationHandler) DeleteResource(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的资源ID参数")
		return
	}

	if err := h.loc.DeleteResource(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ExportResources 导出语言资源。
func (h *LocalizationHandler) ExportResources(c gorp.Context) {
	langID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的语言ID参数")
		return
	}

	data, err := h.loc.ExportResources(c.Context(), uint(langID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, data)
}

// ImportResources 导入语言资源。
func (h *LocalizationHandler) ImportResources(c gorp.Context) {
	langID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的语言ID参数")
		return
	}

	var req request.ImportResources
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 将 request.CreateLocaleResource 转换为 service.CreateLocaleResourceRequest
	svcResources := make([]service.CreateLocaleResourceRequest, len(req.Resources))
	for i, r := range req.Resources {
		svcResources[i] = service.CreateLocaleResourceRequest{
			ResourceName:  r.ResourceName,
			ResourceValue: r.ResourceValue,
		}
	}

	if err := h.loc.ImportResources(c.Context(), uint(langID), svcResources); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]string{"message": "资源导入成功"})
}