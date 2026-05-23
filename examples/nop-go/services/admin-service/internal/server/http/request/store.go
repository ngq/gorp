// Package request 门店模块请求结构 —— 门店管理的HTTP请求DTO
package request

// CreateStoreRequest 创建门店请求
type CreateStoreRequest struct {
	Name     string  `json:"name" binding:"required"`        // 门店名称
	Code     string  `json:"code" binding:"required"`        // 门店编码
	Address  string  `json:"address"`                         // 地址
	Phone    string  `json:"phone"`                           // 联系电话
	Manager  string  `json:"manager"`                         // 店长
	Region   string  `json:"region"`                          // 区域
	Business string  `json:"business"`                        // 营业时间
	Status   int     `json:"status"`                          // 状态
	Lng      float64 `json:"lng"`                             // 经度
	Lat      float64 `json:"lat"`                             // 纬度
}

// UpdateStoreRequest 更新门店请求
type UpdateStoreRequest struct {
	Name     string  `json:"name"`                            // 门店名称
	Address  string  `json:"address"`                         // 地址
	Phone    string  `json:"phone"`                           // 联系电话
	Manager  string  `json:"manager"`                         // 店长
	Region   string  `json:"region"`                          // 区域
	Business string  `json:"business"`                        // 营业时间
	Status   int     `json:"status"`                          // 状态
	Lng      float64 `json:"lng"`                             // 经度
	Lat      float64 `json:"lat"`                             // 纬度
}