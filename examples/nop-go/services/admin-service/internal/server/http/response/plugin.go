// Package response 插件模块响应结构 —— 插件管理的HTTP响应DTO
package response

// PluginResponse 插件响应
type PluginResponse struct {
	ID          uint   `json:"id"`           // 插件ID
	Name        string `json:"name"`         // 插件名称
	Code        string `json:"code"`         // 插件编码
	Version     string `json:"version"`      // 版本号
	Description string `json:"description"`  // 描述
	Author      string `json:"author"`       // 作者
	Config      string `json:"config"`       // 插件配置
	Status      int    `json:"status"`       // 状态
	Sort        int    `json:"sort"`         // 排序
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
}