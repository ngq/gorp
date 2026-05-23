// Package response 供应商模块响应结构 —— 供应商管理的HTTP响应DTO
package response

// VendorResponse 供应商响应
type VendorResponse struct {
	ID          uint   `json:"id"`            // 供应商ID
	Name        string `json:"name"`          // 供应商名称
	Code        string `json:"code"`          // 供应商编码
	Contact     string `json:"contact"`       // 联系人
	Phone       string `json:"phone"`        // 联系电话
	Email       string `json:"email"`        // 邮箱
	Address     string `json:"address"`      // 地址
	Category    string `json:"category"`     // 分类
	BankName    string `json:"bank_name"`    // 开户银行
	BankAccount string `json:"bank_account"` // 银行账号
	Status      int    `json:"status"`        // 状态
	CreatedAt   string `json:"created_at"`   // 创建时间
	UpdatedAt   string `json:"updated_at"`   // 更新时间
}