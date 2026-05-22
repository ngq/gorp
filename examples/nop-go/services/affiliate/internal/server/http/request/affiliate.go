package request

// CreateAffiliate 创建联盟请求。
type CreateAffiliate struct {
	Name   string `json:"name" binding:"required"`     // 联盟名称
	Url    string `json:"url" binding:"required"`      // 联盟URL
	Active bool   `json:"active"`                      // 是否启用
}

// UpdateAffiliate 更新联盟请求。
type UpdateAffiliate struct {
	Name   string `json:"name" binding:"required"`     // 联盟名称
	Url    string `json:"url" binding:"required"`      // 联盟URL
	Active bool   `json:"active"`                      // 是否启用
}