// Package plugin 配送插件接口
package plugin

import (
	"context"
)

// ShippingMethod 配送方式插件接口
//
// 中文说明:
// - 所有配送插件(顺丰、FedEx、DHL等)都实现此接口;
// - 继承基础 Plugin 接口,增加配送特有能力;
// - CalculateShippingRate 计算运费,是核心方法;
// - CreateShipment/GetTrackingInfo 用于物流追踪。
type ShippingMethod interface {
	Plugin

	// CalculateShippingRate 计算运费
	//
	// 中文说明:
	// - 根据收货地址、包裹信息计算运费;
	// - 可能返回多个配送选项(次日达、经济型等);
	// - 用于订单结算时展示配送费用选项。
	CalculateShippingRate(ctx context.Context, req *ShippingRateRequest) (*ShippingRateResult, error)

	// GetTrackingInfo 获取物流追踪信息
	//
	// 中文说明:
	// - 根据运单号查询物流轨迹;
	// - 返回包裹当前状态和历史轨迹。
	GetTrackingInfo(ctx context.Context, trackingNumber string) (*TrackingInfo, error)

	// CreateShipment 创建运单
	//
	// 中文说明:
	// - 向物流公司下单,获取运单号;
	// - 用于订单发货时生成物流单。
	CreateShipment(ctx context.Context, req *CreateShipmentRequest) (*CreateShipmentResult, error)

	// CancelShipment 取消运单
	//
	// 中文说明:
	// - 取消未揽收的运单;
	// - 用于订单取消时。
	CancelShipment(ctx context.Context, shipmentID string) error

	// GetConfiguration 获取配送配置项
	GetConfiguration() []ShippingConfigItem

	// ValidateConfiguration 验证配置是否正确
	ValidateConfiguration(config map[string]string) error
}

// ShippingRateRequest 运费计算请求
//
// 中文说明:
// - 包含计算运费所需的所有信息;
// - 包裹信息和收发货地址。
type ShippingRateRequest struct {
	// OrderID 订单ID(可选)
	OrderID uint64

	// FromAddress 发货地址
	FromAddress ShippingAddress

	// ToAddress 收货地址
	ToAddress ShippingAddress

	// Packages 包裹列表
	Packages []PackageInfo

	// Currency 货币代码
	Currency string
}

// ShippingAddress 配送地址
type ShippingAddress struct {
	Country    string
	State      string
	City       string
	Address1   string
	Address2   string
	ZipCode    string
	Phone      string
}

// PackageInfo 包裹信息
type PackageInfo struct {
	// Weight 重量(kg)
	Weight float64

	// Length 长度(cm)
	Length float64

	// Width 宽度(cm)
	Width float64

	// Height 高度(cm)
	Height float64

	// Value 包裹价值
	Value float64

	// Quantity 数量
	Quantity int
}

// ShippingRateResult 运费计算结果
//
// 中文说明:
// - 可能包含多个配送选项;
// - 用户可选择不同的配送方式。
type ShippingRateResult struct {
	// Rates 配送选项列表
	Rates []ShippingRateOption

	// ErrorMessage 错误信息
	ErrorMessage string
}

// ShippingRateOption 配送选项
type ShippingRateOption struct {
	// Name 配送方式名称
	//
	// 中文说明:
	// - 例如: "顺丰次日达", "顺丰经济型"。
	Name string

	// Code 配送方式代码
	Code string

	// Amount 运费金额
	Amount float64

	// Currency 货币代码
	Currency string

	// EstimatedDays 预计天数
	EstimatedDays int

	// Description 描述
	Description string
}

// TrackingInfo 物流追踪信息
type TrackingInfo struct {
	// TrackingNumber 运单号
	TrackingNumber string

	// Status 当前状态
	//
	// 中文说明:
	// - 例如: "已揽收", "运输中", "已签收"。
	Status string

	// Events 物流轨迹事件
	Events []TrackingEvent

	// EstimatedDelivery 预计送达时间
	EstimatedDelivery string

	// SignedBy 签收人(已签收时)
	SignedBy string

	// SignedAt 签收时间
	SignedAt string
}

// TrackingEvent 物流轨迹事件
type TrackingEvent struct {
	// Time 事件时间
	Time string

	// Location 事件地点
	Location string

	// Description 事件描述
	Description string

	// Status 状态代码
	Status string
}

// CreateShipmentRequest 创建运单请求
type CreateShipmentRequest struct {
	// OrderID 订单ID
	OrderID uint64

	// FromAddress 发货地址
	FromAddress ShippingAddress

	// ToAddress 收货地址
	ToAddress ShippingAddress

	// Packages 包裹列表
	Packages []PackageInfo

	// ServiceCode 配送服务代码
	//
	// 中文说明:
	// - 从 ShippingRateOption.Code 选择。
	ServiceCode string

	// Reference 备注
	Reference string
}

// CreateShipmentResult 创建运单结果
type CreateShipmentResult struct {
	// Success 是否成功
	Success bool

	// ShipmentID 运单ID
	ShipmentID string

	// TrackingNumber 运单号
	TrackingNumber string

	// LabelURL 运单标签打印URL
	LabelURL string

	// ErrorMessage 错误信息
	ErrorMessage string
}

// ShippingConfigItem 配送配置项
type ShippingConfigItem struct {
	Name     string
	Label    string
	Type     string // text, textarea, boolean, select
	Required bool
	Default  string
	Options  []ShippingConfigOption
	HelpText string
}

// ShippingConfigOption 下拉选项
type ShippingConfigOption struct {
	Value string
	Label string
}