package response

import "nop-go/services/content-service/internal/service"

// ==================== 语言响应 ====================

// LanguageList 语言列表响应
type LanguageList struct {
	Items []*service.LanguageResponse `json:"items"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"size"`
}

// ==================== 本地化资源响应 ====================

// LocaleResourceList 本地化资源列表响应
type LocaleResourceList struct {
	Items []*service.LocaleResourceResponse `json:"items"`
	Total int64                             `json:"total"`
	Page  int                               `json:"page"`
	Size  int                               `json:"size"`
}