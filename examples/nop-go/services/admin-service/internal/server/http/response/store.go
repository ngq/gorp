// Package response 门店模块响应结构 —— 门店管理的HTTP响应DTO
package response

// StoreResponse 门店响应
type StoreResponse struct {
	ID        uint    `json:"id"`          // 门店ID
	Name      string  `json:"name"`        // 门店名称
	Code      string  `json:"code"`        // 门店编码
	Address   string  `json:"address"`     // 地址
	Phone     string  `json:"phone"`       // 联系电话
	Manager   string  `json:"manager"`    // 店长
	Region    string  `json:"region"`     // 区域
	Business  string  `json:"business"`   // 营业时间
	Status    int     `json:"status"`     // 状态
	Lng       float64 `json:"lng"`         // 经度
	Lat       float64 `json:"lat"`         // 纬度
	CreatedAt string  `json:"created_at"` // 创建时间
	UpdatedAt string  `json:"updated_at"` // 更新时间
}