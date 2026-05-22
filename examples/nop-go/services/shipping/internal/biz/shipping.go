// Package biz 提供 shipping 服务的业务逻辑层
//
// 配送服务包含五个子领域：
// 1. 配送提供者（Provider）— 物流公司/承运商
// 2. 配送方式（Method）— 具体配送选项（如标准/加急）
// 3. 配送日期（DeliveryDate）— 可选配送日期
// 4. 仓库（Warehouse）— 发货仓库
// 5. 运费估算（Estimate）— 根据条件估算运费
package biz

import (
	"context"
	"time"

	"nop-go/services/shipping/internal/server/http/request"
	"nop-go/services/shipping/internal/server/http/response"
)

// ==================== 领域实体定义 ====================

// Provider 配送提供者领域实体
type Provider struct {
	ID            uint      // 提供者 ID
	Name          string    // 提供者名称
	SystemKeyword string    // 系统关键字标识
	DisplayOrder  int       // 显示排序
	IsActive      bool      // 是否启用
	LogoURL       string    // Logo 地址
	TrackingURL   string    // 物流追踪 URL 模板
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// Method 配送方式领域实体
type Method struct {
	ID             uint      // 配送方式 ID
	Name           string    // 配送方式名称
	SystemKeyword  string    // 系统关键字标识
	ProviderID     uint      // 关联的配送提供者 ID
	ProviderName   string    // 配送提供者名称（冗余展示）
	DisplayOrder   int       // 显示排序
	IsActive       bool      // 是否启用
	Rate           float64   // 基础运费
	MinOrderAmount float64   // 免运费最低订单金额
	MaxOrderAmount float64   // 运费适用最高订单金额
	EstimatedDays  int       // 预计配送天数
	Description    string    // 配送方式描述
	CreatedAt      time.Time // 创建时间
	UpdatedAt      time.Time // 更新时间
}

// DeliveryDate 配送日期领域实体
type DeliveryDate struct {
	ID               uint      // 配送日期 ID
	ShippingMethodID uint      // 关联的配送方式 ID
	DeliveryDate     string    // 可选配送日期
	IsAvailable      bool      // 该日期是否可选
	Description      string    // 日期说明
	CreatedAt        time.Time // 创建时间
	UpdatedAt        time.Time // 更新时间
}

// Warehouse 仓库领域实体
type Warehouse struct {
	ID          uint      // 仓库 ID
	Name        string    // 仓库名称
	Code        string    // 仓库编码
	Address     string    // 仓库地址
	City        string    // 城市
	CountryID   uint      // 国家 ID
	StateID     uint      // 省/州 ID
	ZipCode     string    // 邮编
	PhoneNumber string    // 联系电话
	IsActive    bool      // 是否启用
	Latitude    float64   // 纬度
	Longitude   float64   // 经度
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// ==================== 仓储接口定义 ====================

// ProviderRepository 配送提供者仓储接口
type ProviderRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*Provider, int64, error)
	Update(ctx context.Context, provider *Provider) (*Provider, error)
}

// MethodRepository 配送方式仓储接口
type MethodRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*Method, int64, error)
	Create(ctx context.Context, method *Method) (*Method, error)
	Update(ctx context.Context, method *Method) (*Method, error)
	Delete(ctx context.Context, id uint) error
}

// DeliveryDateRepository 配送日期仓储接口
type DeliveryDateRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*DeliveryDate, int64, error)
	Create(ctx context.Context, date *DeliveryDate) (*DeliveryDate, error)
	Update(ctx context.Context, date *DeliveryDate) (*DeliveryDate, error)
}

// WarehouseRepository 仓库仓储接口
type WarehouseRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*Warehouse, int64, error)
	Create(ctx context.Context, warehouse *Warehouse) (*Warehouse, error)
	Update(ctx context.Context, warehouse *Warehouse) (*Warehouse, error)
}

// ==================== 用例实现 ====================

// ShippingUsecase 配送服务业务用例
type ShippingUsecase struct {
	providerRepo    ProviderRepository
	methodRepo      MethodRepository
	deliveryDateRepo DeliveryDateRepository
	warehouseRepo   WarehouseRepository
}

// NewShippingUsecase 创建配送服务业务用例
func NewShippingUsecase(
	providerRepo ProviderRepository,
	methodRepo MethodRepository,
	deliveryDateRepo DeliveryDateRepository,
	warehouseRepo WarehouseRepository,
) *ShippingUsecase {
	return &ShippingUsecase{
		providerRepo:    providerRepo,
		methodRepo:      methodRepo,
		deliveryDateRepo: deliveryDateRepo,
		warehouseRepo:   warehouseRepo,
	}
}

// ==================== 配送提供者用例 ====================

// ListProviders 获取配送提供者列表
func (uc *ShippingUsecase) ListProviders(ctx context.Context, page, pageSize int) ([]*response.ProviderResponse, int64, error) {
	providers, total, err := uc.providerRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.ProviderResponse, len(providers))
	for i, p := range providers {
		items[i] = toProviderResponse(p)
	}
	return items, total, nil
}

// UpdateProvider 更新配送提供者
func (uc *ShippingUsecase) UpdateProvider(ctx context.Context, req request.UpdateProviderRequest) (*response.ProviderResponse, error) {
	provider := &Provider{
		ID:            req.ID,
		Name:          req.Name,
		SystemKeyword: req.SystemKeyword,
		DisplayOrder:  req.DisplayOrder,
		IsActive:      req.IsActive,
		LogoURL:       req.LogoURL,
		TrackingURL:   req.TrackingURL,
	}

	updated, err := uc.providerRepo.Update(ctx, provider)
	if err != nil {
		return nil, err
	}
	return toProviderResponse(updated), nil
}

// ==================== 配送方式用例 ====================

// ListMethods 获取配送方式列表
func (uc *ShippingUsecase) ListMethods(ctx context.Context, page, pageSize int) ([]*response.MethodResponse, int64, error) {
	methods, total, err := uc.methodRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.MethodResponse, len(methods))
	for i, m := range methods {
		items[i] = toMethodResponse(m)
	}
	return items, total, nil
}

// CreateMethod 创建配送方式
func (uc *ShippingUsecase) CreateMethod(ctx context.Context, req request.CreateMethodRequest) (*response.MethodResponse, error) {
	method := &Method{
		Name:           req.Name,
		SystemKeyword:  req.SystemKeyword,
		ProviderID:     req.ProviderID,
		DisplayOrder:   req.DisplayOrder,
		IsActive:       req.IsActive,
		Rate:           req.Rate,
		MinOrderAmount: req.MinOrderAmount,
		MaxOrderAmount: req.MaxOrderAmount,
		EstimatedDays:  req.EstimatedDays,
		Description:    req.Description,
	}

	created, err := uc.methodRepo.Create(ctx, method)
	if err != nil {
		return nil, err
	}
	return toMethodResponse(created), nil
}

// UpdateMethod 更新配送方式
func (uc *ShippingUsecase) UpdateMethod(ctx context.Context, req request.UpdateMethodRequest) (*response.MethodResponse, error) {
	method := &Method{
		ID:             req.ID,
		Name:           req.Name,
		SystemKeyword:  req.SystemKeyword,
		ProviderID:     req.ProviderID,
		DisplayOrder:   req.DisplayOrder,
		IsActive:       req.IsActive,
		Rate:           req.Rate,
		MinOrderAmount: req.MinOrderAmount,
		MaxOrderAmount: req.MaxOrderAmount,
		EstimatedDays:  req.EstimatedDays,
		Description:    req.Description,
	}

	updated, err := uc.methodRepo.Update(ctx, method)
	if err != nil {
		return nil, err
	}
	return toMethodResponse(updated), nil
}

// DeleteMethod 删除配送方式
func (uc *ShippingUsecase) DeleteMethod(ctx context.Context, id uint) error {
	return uc.methodRepo.Delete(ctx, id)
}

// ==================== 配送日期用例 ====================

// ListDeliveryDates 获取配送日期列表
func (uc *ShippingUsecase) ListDeliveryDates(ctx context.Context, page, pageSize int) ([]*response.DeliveryDateResponse, int64, error) {
	dates, total, err := uc.deliveryDateRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.DeliveryDateResponse, len(dates))
	for i, d := range dates {
		items[i] = toDeliveryDateResponse(d)
	}
	return items, total, nil
}

// CreateDeliveryDate 创建配送日期
func (uc *ShippingUsecase) CreateDeliveryDate(ctx context.Context, req request.CreateDeliveryDateRequest) (*response.DeliveryDateResponse, error) {
	date := &DeliveryDate{
		ShippingMethodID: req.ShippingMethodID,
		DeliveryDate:     req.DeliveryDate,
		IsAvailable:      req.IsAvailable,
		Description:      req.Description,
	}

	created, err := uc.deliveryDateRepo.Create(ctx, date)
	if err != nil {
		return nil, err
	}
	return toDeliveryDateResponse(created), nil
}

// UpdateDeliveryDate 更新配送日期
func (uc *ShippingUsecase) UpdateDeliveryDate(ctx context.Context, req request.UpdateDeliveryDateRequest) (*response.DeliveryDateResponse, error) {
	date := &DeliveryDate{
		ID:               req.ID,
		ShippingMethodID: req.ShippingMethodID,
		DeliveryDate:     req.DeliveryDate,
		IsAvailable:      req.IsAvailable,
		Description:      req.Description,
	}

	updated, err := uc.deliveryDateRepo.Update(ctx, date)
	if err != nil {
		return nil, err
	}
	return toDeliveryDateResponse(updated), nil
}

// ==================== 仓库用例 ====================

// ListWarehouses 获取仓库列表
func (uc *ShippingUsecase) ListWarehouses(ctx context.Context, page, pageSize int) ([]*response.WarehouseResponse, int64, error) {
	warehouses, total, err := uc.warehouseRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.WarehouseResponse, len(warehouses))
	for i, w := range warehouses {
		items[i] = toWarehouseResponse(w)
	}
	return items, total, nil
}

// CreateWarehouse 创建仓库
func (uc *ShippingUsecase) CreateWarehouse(ctx context.Context, req request.CreateWarehouseRequest) (*response.WarehouseResponse, error) {
	warehouse := &Warehouse{
		Name:        req.Name,
		Code:        req.Code,
		Address:     req.Address,
		City:        req.City,
		CountryID:   req.CountryID,
		StateID:     req.StateID,
		ZipCode:     req.ZipCode,
		PhoneNumber: req.PhoneNumber,
		IsActive:    req.IsActive,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
	}

	created, err := uc.warehouseRepo.Create(ctx, warehouse)
	if err != nil {
		return nil, err
	}
	return toWarehouseResponse(created), nil
}

// UpdateWarehouse 更新仓库
func (uc *ShippingUsecase) UpdateWarehouse(ctx context.Context, req request.UpdateWarehouseRequest) (*response.WarehouseResponse, error) {
	warehouse := &Warehouse{
		ID:          req.ID,
		Name:        req.Name,
		Code:        req.Code,
		Address:     req.Address,
		City:        req.City,
		CountryID:   req.CountryID,
		StateID:     req.StateID,
		ZipCode:     req.ZipCode,
		PhoneNumber: req.PhoneNumber,
		IsActive:    req.IsActive,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
	}

	updated, err := uc.warehouseRepo.Update(ctx, warehouse)
	if err != nil {
		return nil, err
	}
	return toWarehouseResponse(updated), nil
}

// ==================== 运费估算用例 ====================

// EstimateShipping 估算运费
// 根据发货仓库、收货地址、订单金额和重量计算各可用配送方式的运费
func (uc *ShippingUsecase) EstimateShipping(ctx context.Context, req request.EstimateShippingRequest) (*response.ListShippingEstimatesResponse, error) {
	// TODO: 实现运费估算逻辑
	// 1. 根据仓库 ID 获取仓库信息
	// 2. 根据收货地址筛选可用的配送方式
	// 3. 根据订单金额判断是否免运费
	// 4. 根据重量计算运费
	// 当前返回空列表占位
	return &response.ListShippingEstimatesResponse{
		Items: []*response.ShippingEstimateResponse{},
	}, nil
}

// ==================== 内部转换函数 ====================

// toProviderResponse 领域实体转换为配送提供者响应 DTO
func toProviderResponse(p *Provider) *response.ProviderResponse {
	return &response.ProviderResponse{
		ID:            p.ID,
		Name:          p.Name,
		SystemKeyword: p.SystemKeyword,
		DisplayOrder:  p.DisplayOrder,
		IsActive:      p.IsActive,
		LogoURL:       p.LogoURL,
		TrackingURL:   p.TrackingURL,
		CreatedAt:     p.CreatedAt.Unix(),
		UpdatedAt:     p.UpdatedAt.Unix(),
	}
}

// toMethodResponse 领域实体转换为配送方式响应 DTO
func toMethodResponse(m *Method) *response.MethodResponse {
	return &response.MethodResponse{
		ID:             m.ID,
		Name:           m.Name,
		SystemKeyword:  m.SystemKeyword,
		ProviderID:     m.ProviderID,
		ProviderName:   m.ProviderName,
		DisplayOrder:   m.DisplayOrder,
		IsActive:       m.IsActive,
		Rate:           m.Rate,
		MinOrderAmount: m.MinOrderAmount,
		MaxOrderAmount: m.MaxOrderAmount,
		EstimatedDays:  m.EstimatedDays,
		Description:    m.Description,
		CreatedAt:      m.CreatedAt.Unix(),
		UpdatedAt:      m.UpdatedAt.Unix(),
	}
}

// toDeliveryDateResponse 领域实体转换为配送日期响应 DTO
func toDeliveryDateResponse(d *DeliveryDate) *response.DeliveryDateResponse {
	return &response.DeliveryDateResponse{
		ID:               d.ID,
		ShippingMethodID: d.ShippingMethodID,
		DeliveryDate:     d.DeliveryDate,
		IsAvailable:      d.IsAvailable,
		Description:      d.Description,
		CreatedAt:        d.CreatedAt.Unix(),
		UpdatedAt:        d.UpdatedAt.Unix(),
	}
}

// toWarehouseResponse 领域实体转换为仓库响应 DTO
func toWarehouseResponse(w *Warehouse) *response.WarehouseResponse {
	return &response.WarehouseResponse{
		ID:          w.ID,
		Name:        w.Name,
		Code:        w.Code,
		Address:     w.Address,
		City:        w.City,
		CountryID:   w.CountryID,
		StateID:     w.StateID,
		ZipCode:     w.ZipCode,
		PhoneNumber: w.PhoneNumber,
		IsActive:    w.IsActive,
		Latitude:    w.Latitude,
		Longitude:   w.Longitude,
		CreatedAt:   w.CreatedAt.Unix(),
		UpdatedAt:   w.UpdatedAt.Unix(),
	}
}
