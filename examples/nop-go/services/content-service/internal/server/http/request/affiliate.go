package request

// ==================== 推广合作方请求 ====================

// CreateAffiliate 创建推广合作方请求
type CreateAffiliate struct {
	Name       string  `json:"name" binding:"required"`  // 合作方名称（必填）
	Code       string  `json:"code" binding:"required"`  // 合作方编码（必填）
	Contact    string  `json:"contact"`                  // 联系方式
	Website    string  `json:"website"`                  // 网站
	Commission float64 `json:"commission"`               // 佣金比例
	Status     string  `json:"status"`                   // 状态
}

// UpdateAffiliate 更新推广合作方请求
type UpdateAffiliate struct {
	Name       string  `json:"name" binding:"required"`
	Code       string  `json:"code" binding:"required"`
	Contact    string  `json:"contact"`
	Website    string  `json:"website"`
	Commission float64 `json:"commission"`
	Status     string  `json:"status"`
}