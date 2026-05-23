// Package request 供应商模块请求结构 —— 供应商管理的HTTP请求DTO
package request

// CreateVendorRequest 创建供应商请求
type CreateVendorRequest struct {
	Name        string `json:"name" binding:"required"`        // 供应商名称
	Code        string `json:"code" binding:"required"`        // 供应商编码
	Contact     string `json:"contact"`                         // 联系人
	Phone       string `json:"phone"`                           // 联系电话
	Email       string `json:"email"`                           // 邮箱
	Address     string `json:"address"`                         // 地址
	Category    string `json:"category"`                        // 分类
	BankName    string `json:"bank_name"`                       // 开户银行
	BankAccount string `json:"bank_account"`                    // 银行账号
	Status      int    `json:"status"`                          // 状态
}

// UpdateVendorRequest 更新供应商请求
type UpdateVendorRequest struct {
	Name        string `json:"name"`                            // 供应商名称
	Contact     string `json:"contact"`                         // 联系人
	Phone       string `json:"phone"`                           // 联系电话
	Email       string `json:"email"`                           // 邮箱
	Address     string `json:"address"`                         // 地址
	Category    string `json:"category"`                        // 分类
	BankName    string `json:"bank_name"`                       // 开户银行
	BankAccount string `json:"bank_account"`                    // 银行账号
	Status      int    `json:"status"`                          // 状态
}