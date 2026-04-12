// Package plugin 支付插件接口
package plugin

import (
	"context"
)

// PaymentMethod 支付方式插件接口
//
// 中文说明:
// - 所有支付插件(支付宝、微信、PayPal等)都实现此接口;
// - 继承基础 Plugin 接口,增加支付特有能力;
// - ProcessPayment 是核心方法,创建支付订单;
// - Refund/Capture/Void 是可选方法,根据支付方式特性实现。
type PaymentMethod interface {
	Plugin

	// ProcessPayment 处理支付
	//
	// 中文说明:
	// - 创建支付订单,返回支付链接或跳转URL;
	// - 不同支付方式的实现差异很大;
	// - 例如支付宝返回跳转URL,微信返回二维码。
	ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResult, error)

	// Refund 退款
	//
	// 中文说明:
	// - 全额或部分退款;
	// - 返回退款流水号。
	Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error)

	// Capture 捕获预授权
	//
	// 中文说明:
	// - 用于信用卡预授权模式;
	// - 先授权后扣款;
	// - 不支持预授权的支付方式应返回 error。
	Capture(ctx context.Context, req *CaptureRequest) (*CaptureResult, error)

	// Void 取消预授权
	//
	// 中文说明:
	// - 取消未扣款的预授权;
	// - 不支持预授权的支付方式应返回 error。
	Void(ctx context.Context, req *VoidRequest) (*VoidResult, error)

	// GetConfiguration 获取支付配置项
	//
	// 中文说明:
	// - 返回需要在管理后台配置的字段列表;
	// - 例如支付宝需要 app_id、private_key 等;
	// - 用于生成动态配置表单。
	GetConfiguration() []PaymentConfigItem

	// ValidateConfiguration 验证配置是否正确
	//
	// 中文说明:
	// - 保存配置前验证;
	// - 检查必填项、格式等;
	// - 返回具体的错误信息。
	ValidateConfiguration(config map[string]string) error
}

// ProcessPaymentRequest 支付请求
//
// 中文说明:
// - 包含创建支付订单所需的所有信息;
// - OrderID 用于关联订单;
// - ReturnURL/NotifyURL 用于支付结果回调。
type ProcessPaymentRequest struct {
	// OrderID 订单ID
	OrderID uint64

	// Amount 支付金额
	Amount float64

	// Currency 货币代码
	Currency string

	// CustomerID 客户ID
	CustomerID uint64

	// ReturnURL 前端同步回调URL
	//
	// 中文说明:
	// - 支付完成后浏览器跳转到此URL;
	// - 用于显示支付结果页面。
	ReturnURL string

	// NotifyURL 后端异步回调URL
	//
	// 中文说明:
	// - 支付平台异步通知此URL;
	// - 用于更新订单状态。
	NotifyURL string

	// CustomFields 自定义字段
	//
	// 中文说明:
	// - 各支付方式特有的参数;
	// - 例如微信的 openid、支付宝的 buyer_id。
	CustomFields map[string]string
}

// ProcessPaymentResult 支付结果
//
// 中文说明:
// - 包含支付处理后的返回信息;
// - RedirectURL 用于跳转支付页面;
// - QRCodeURL 用于扫码支付。
type ProcessPaymentResult struct {
	// Success 是否成功创建支付订单
	Success bool

	// TransactionID 交易流水号
	//
	// 中文说明:
	// - 支付平台返回的流水号;
	// - 用于后续查询和退款。
	TransactionID string

	// RedirectURL 跳转支付页面URL
	//
	// 中文说明:
	// - 用于网页支付,浏览器跳转到此URL;
	// - 支付宝网页支付、微信H5支付等。
	RedirectURL string

	// QRCodeURL 二维码内容URL
	//
	// 中文说明:
	// - 用于扫码支付;
	// - 前端生成二维码供用户扫描。
	QRCodeURL string

	// ErrorMessage 错误信息
	ErrorMessage string

	// RawData 支付平台原始返回
	//
	// 中文说明:
	// - 用于调试和日志记录;
	// - 不同支付方式格式不同。
	RawData map[string]interface{}
}

// RefundRequest 退款请求
type RefundRequest struct {
	// PaymentID 支付记录ID
	PaymentID uint64

	// TransactionID 原交易流水号
	TransactionID string

	// Amount 退款金额
	Amount float64

	// Reason 退款原因
	Reason string
}

// RefundResult 退款结果
type RefundResult struct {
	// Success 是否成功
	Success bool

	// RefundTransactionID 退款流水号
	RefundTransactionID string

	// ErrorMessage 错误信息
	ErrorMessage string
}

// CaptureRequest 捕获预授权请求
type CaptureRequest struct {
	// TransactionID 预授权流水号
	TransactionID string

	// Amount 捕获金额
	Amount float64
}

// CaptureResult 捕获结果
type CaptureResult struct {
	Success       bool
	TransactionID string
	ErrorMessage  string
}

// VoidRequest 取消预授权请求
type VoidRequest struct {
	TransactionID string
}

// VoidResult 取消结果
type VoidResult struct {
	Success      bool
	ErrorMessage string
}

// PaymentConfigItem 支付配置项
//
// 中文说明:
// - 描述一个配置字段的元信息;
// - 用于在管理后台生成配置表单;
// - Type 决定输入控件类型。
type PaymentConfigItem struct {
	// Name 配置字段名
	Name string

	// Label 显示标签
	Label string

	// Type 输入类型
	//
	// 中文说明:
	// - text: 单行文本;
	// - textarea: 多行文本;
	// - boolean: 开关;
	// - select: 下拉选择;
	// - password: 密码输入。
	Type string

	// Required 是否必填
	Required bool

	// Default 默认值
	Default string

	// Options 下拉选项(仅 select 类型)
	Options []PaymentConfigOption

	// HelpText 帮助文本
	HelpText string
}

// PaymentConfigOption 下拉选项
type PaymentConfigOption struct {
	Value string
	Label string
}