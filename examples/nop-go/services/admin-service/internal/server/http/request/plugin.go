// Package request 插件模块请求结构 —— 插件管理的HTTP请求DTO
package request

// CreatePluginRequest 创建插件请求
type CreatePluginRequest struct {
	Name        string `json:"name" binding:"required"`        // 插件名称
	Code        string `json:"code" binding:"required"`        // 插件编码
	Version     string `json:"version"`                         // 版本号
	Description string `json:"description"`                     // 描述
	Author      string `json:"author"`                          // 作者
	Config      string `json:"config"`                          // 插件配置（JSON）
	Status      int    `json:"status"`                          // 状态
	Sort        int    `json:"sort"`                            // 排序权重
}

// UpdatePluginRequest 更新插件请求
type UpdatePluginRequest struct {
	Name        string `json:"name"`                            // 插件名称
	Version     string `json:"version"`                         // 版本号
	Description string `json:"description"`                     // 描述
	Author      string `json:"author"`                          // 作者
	Config      string `json:"config"`                          // 插件配置（JSON）
	Status      int    `json:"status"`                          // 状态
	Sort        int    `json:"sort"`                            // 排序权重
}