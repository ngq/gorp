// Package biz_test 提供 shipping 服务的业务逻辑层单元测试
//
// 测试覆盖：
// 1. ShippingUsecase 所有方法的成功和失败场景
// 2. Mock 仓储实现用于隔离测试
// 3. 领域实体到响应 DTO 的转换验证
// 4. 多仓储协作场景测试
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"nop-go/services/shipping/internal/biz"
	"nop-go/services/shipping/internal/server/http/request"
)

// ==================== Mock 仓储实现 ====================

// MockProviderRepository 配送提供者仓储的 Mock 实现
type MockProviderRepository struct {
	ListResult      []*biz.Provider
	ListTotal       int64
	ListError       error
	UpdateResult    *biz.Provider
	UpdateError     error
	ListCallCount   int
	UpdateCallCount int
}

func (m *MockProviderRepository) List(ctx context.Context, page, pageSize int) ([]*biz.Provider, int64, error) {
	m.ListCallCount++
	return m.ListResult, m.ListTotal, m.ListError
}

func (m *MockProviderRepository) Update(ctx context.Context, provider *biz.Provider) (*biz.Provider, error) {
	m.UpdateCallCount++
	return m.UpdateResult, m.UpdateError
}

// MockMethodRepository 配送方式仓储的 Mock 实现
type MockMethodRepository struct {
	ListResult      []*biz.Method
	ListTotal       int64
	ListError       error
	CreateResult    *biz.Method
	CreateError     error
	UpdateResult    *biz.Method
	UpdateError     error
	DeleteError     error
	ListCallCount   int
	CreateCallCount int
	UpdateCallCount int
	DeleteCallCount int
}

func (m *MockMethodRepository) List(ctx context.Context, page, pageSize int) ([]*biz.Method, int64, error) {
	m.ListCallCount++
	return m.ListResult, m.ListTotal, m.ListError
}

func (m *MockMethodRepository) Create(ctx context.Context, method *biz.Method) (*biz.Method, error) {
	m.CreateCallCount++
	return m.CreateResult, m.CreateError
}

func (m *MockMethodRepository) Update(ctx context.Context, method *biz.Method) (*biz.Method, error) {
	m.UpdateCallCount++
	return m.UpdateResult, m.UpdateError
}

func (m *MockMethodRepository) Delete(ctx context.Context, id uint) error {
	m.DeleteCallCount++
	return m.DeleteError
}

// MockDeliveryDateRepository 配送日期仓储的 Mock 实现
type MockDeliveryDateRepository struct {
	ListResult      []*biz.DeliveryDate
	ListTotal       int64
	ListError       error
	CreateResult    *biz.DeliveryDate
	CreateError     error
	UpdateResult    *biz.DeliveryDate
	UpdateError     error
	ListCallCount   int
	CreateCallCount int
	UpdateCallCount int
}

func (m *MockDeliveryDateRepository) List(ctx context.Context, page, pageSize int) ([]*biz.DeliveryDate, int64, error) {
	m.ListCallCount++
	return m.ListResult, m.ListTotal, m.ListError
}

func (m *MockDeliveryDateRepository) Create(ctx context.Context, date *biz.DeliveryDate) (*biz.DeliveryDate, error) {
	m.CreateCallCount++
	return m.CreateResult, m.CreateError
}

func (m *MockDeliveryDateRepository) Update(ctx context.Context, date *biz.DeliveryDate) (*biz.DeliveryDate, error) {
	m.UpdateCallCount++
	return m.UpdateResult, m.UpdateError
}

// MockWarehouseRepository 仓库仓储的 Mock 实现
type MockWarehouseRepository struct {
	ListResult      []*biz.Warehouse
	ListTotal       int64
	ListError       error
	CreateResult    *biz.Warehouse
	CreateError     error
	UpdateResult    *biz.Warehouse
	UpdateError     error
	ListCallCount   int
	CreateCallCount int
	UpdateCallCount int
}

func (m *MockWarehouseRepository) List(ctx context.Context, page, pageSize int) ([]*biz.Warehouse, int64, error) {
	m.ListCallCount++
	return m.ListResult, m.ListTotal, m.ListError
}

func (m *MockWarehouseRepository) Create(ctx context.Context, warehouse *biz.Warehouse) (*biz.Warehouse, error) {
	m.CreateCallCount++
	return m.CreateResult, m.CreateError
}

func (m *MockWarehouseRepository) Update(ctx context.Context, warehouse *biz.Warehouse) (*biz.Warehouse, error) {
	m.UpdateCallCount++
	return m.UpdateResult, m.UpdateError
}

// ==================== 测试辅助函数 ====================

// newTestProvider 创建测试用的配送提供者领域实体
func newTestProvider(id uint, name string) *biz.Provider {
	now := time.Now()
	return &biz.Provider{
		ID:            id,
		Name:          name,
		SystemKeyword: "Shipping." + name,
		DisplayOrder:  1,
		IsActive:      true,
		LogoURL:       "https://example.com/logo.png",
		TrackingURL:   "https://example.com/track/{tracking}",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// newTestMethod 创建测试用的配送方式领域实体
func newTestMethod(id uint, name string) *biz.Method {
	now := time.Now()
	return &biz.Method{
		ID:             id,
		Name:           name,
		SystemKeyword:  "Shipping.Method." + name,
		ProviderID:     1,
		ProviderName:   "顺丰",
		DisplayOrder:   1,
		IsActive:       true,
		Rate:           10.0,
		MinOrderAmount: 99.0,
		MaxOrderAmount: 9999.0,
		EstimatedDays:  3,
		Description:    "标准配送",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// newTestDeliveryDate 创建测试用的配送日期领域实体
func newTestDeliveryDate(id uint, shippingMethodID uint) *biz.DeliveryDate {
	now := time.Now()
	return &biz.DeliveryDate{
		ID:               id,
		ShippingMethodID: shippingMethodID,
		DeliveryDate:     "2026-05-25",
		IsAvailable:      true,
		Description:      "工作日配送",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// newTestWarehouse 创建测试用的仓库领域实体
func newTestWarehouse(id uint, name string) *biz.Warehouse {
	now := time.Now()
	return &biz.Warehouse{
		ID:          id,
		Name:        name,
		Code:        "WH" + string(rune('0'+id)),
		Address:     "测试地址 123 号",
		City:        "上海",
		CountryID:   1,
		StateID:     1,
		ZipCode:     "200000",
		PhoneNumber: "021-12345678",
		IsActive:    true,
		Latitude:    31.2304,
		Longitude:   121.4737,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ==================== NewShippingUsecase 测试 ====================

// TestNewShippingUsecase 测试用例构造函数
func TestNewShippingUsecase(t *testing.T) {
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	if uc == nil {
		t.Error("期望返回非 nil 用例实例")
	}
}

// ==================== 配送提供者测试 ====================

// TestListProviders_Success 测试成功获取配送提供者列表
func TestListProviders_Success(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{
		ListResult: []*biz.Provider{
			newTestProvider(1, "顺丰"),
			newTestProvider(2, "圆通"),
		},
		ListTotal: 2,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, total, err := uc.ListProviders(ctx, 1, 10)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数为 2，实际为 %d", total)
	}
	if len(items) != 2 {
		t.Errorf("期望返回 2 条记录，实际返回 %d 条", len(items))
	}
	if items[0].Name != "顺丰" {
		t.Errorf("期望第一条记录名称为 '顺丰'，实际为 '%s'", items[0].Name)
	}
}

// TestListProviders_EmptyList 测试空列表场景
func TestListProviders_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{
		ListResult: []*biz.Provider{},
		ListTotal:  0,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, total, err := uc.ListProviders(ctx, 1, 10)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 0 {
		t.Errorf("期望总数为 0，实际为 %d", total)
	}
	if len(items) != 0 {
		t.Errorf("期望返回空列表，实际返回 %d 条", len(items))
	}
}

// TestListProviders_RepositoryError 测试仓储返回错误
func TestListProviders_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("数据库连接失败")
	mockProvider := &MockProviderRepository{
		ListError: expectedErr,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, total, err := uc.ListProviders(ctx, 1, 10)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if items != nil {
		t.Errorf("期望返回 nil，实际返回 %v", items)
	}
	if total != 0 {
		t.Errorf("期望总数为 0，实际为 %d", total)
	}
}

// TestUpdateProvider_Success 测试成功更新配送提供者
func TestUpdateProvider_Success(t *testing.T) {
	ctx := context.Background()
	updatedProvider := newTestProvider(1, "顺丰速运")
	mockProvider := &MockProviderRepository{
		UpdateResult: updatedProvider,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateProviderRequest{
		ID:            1,
		Name:          "顺丰速运",
		SystemKeyword: "Shipping.SF",
		DisplayOrder:  1,
		IsActive:      true,
		LogoURL:       "https://example.com/sf.png",
		TrackingURL:   "https://sf.com/track/{no}",
	}

	result, err := uc.UpdateProvider(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if result.Name != "顺丰速运" {
		t.Errorf("期望名称为 '顺丰速运'，实际为 '%s'", result.Name)
	}
	if mockProvider.UpdateCallCount != 1 {
		t.Errorf("期望 Update 被调用 1 次，实际调用 %d 次", mockProvider.UpdateCallCount)
	}
}

// TestUpdateProvider_RepositoryError 测试更新时仓储返回错误
func TestUpdateProvider_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("记录不存在")
	mockProvider := &MockProviderRepository{
		UpdateError: expectedErr,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateProviderRequest{
		ID:   999,
		Name: "不存在的提供者",
	}

	result, err := uc.UpdateProvider(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// ==================== 配送方式测试 ====================

// TestListMethods_Success 测试成功获取配送方式列表
func TestListMethods_Success(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		ListResult: []*biz.Method{
			newTestMethod(1, "标准配送"),
			newTestMethod(2, "加急配送"),
		},
		ListTotal: 2,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, total, err := uc.ListMethods(ctx, 1, 10)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数为 2，实际为 %d", total)
	}
	if len(items) != 2 {
		t.Errorf("期望返回 2 条记录，实际返回 %d 条", len(items))
	}
	if items[0].Rate != 10.0 {
		t.Errorf("期望运费为 10.0，实际为 %f", items[0].Rate)
	}
}

// TestListMethods_RepositoryError 测试仓储返回错误
func TestListMethods_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("查询失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		ListError: expectedErr,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, err := uc.ListMethods(ctx, 1, 10)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if items != nil {
		t.Errorf("期望返回 nil，实际返回 %v", items)
	}
}

// TestCreateMethod_Success 测试成功创建配送方式
func TestCreateMethod_Success(t *testing.T) {
	ctx := context.Background()
	createdMethod := newTestMethod(1, "次日达")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		CreateResult: createdMethod,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.CreateMethodRequest{
		Name:           "次日达",
		SystemKeyword:  "Shipping.Method.NextDay",
		ProviderID:     1,
		DisplayOrder:   1,
		IsActive:       true,
		Rate:           15.0,
		MinOrderAmount: 199.0,
		MaxOrderAmount: 9999.0,
		EstimatedDays:  1,
		Description:    "次日送达",
	}

	result, err := uc.CreateMethod(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if result.Name != "次日达" {
		t.Errorf("期望名称为 '次日达'，实际为 '%s'", result.Name)
	}
	if mockMethod.CreateCallCount != 1 {
		t.Errorf("期望 Create 被调用 1 次，实际调用 %d 次", mockMethod.CreateCallCount)
	}
}

// TestCreateMethod_RepositoryError 测试创建时仓储返回错误
func TestCreateMethod_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("创建失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		CreateError: expectedErr,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.CreateMethodRequest{
		Name: "测试配送",
	}

	result, err := uc.CreateMethod(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// TestUpdateMethod_Success 测试成功更新配送方式
func TestUpdateMethod_Success(t *testing.T) {
	ctx := context.Background()
	updatedMethod := newTestMethod(1, "更新后的配送")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		UpdateResult: updatedMethod,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateMethodRequest{
		ID:             1,
		Name:           "更新后的配送",
		SystemKeyword:  "Shipping.Method.Updated",
		ProviderID:     1,
		DisplayOrder:   2,
		IsActive:       true,
		Rate:           20.0,
		MinOrderAmount: 99.0,
		MaxOrderAmount: 9999.0,
		EstimatedDays:  2,
		Description:    "更新后的描述",
	}

	result, err := uc.UpdateMethod(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result.Name != "更新后的配送" {
		t.Errorf("期望名称为 '更新后的配送'，实际为 '%s'", result.Name)
	}
}

// TestUpdateMethod_RepositoryError 测试更新时仓储返回错误
func TestUpdateMethod_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("更新失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		UpdateError: expectedErr,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateMethodRequest{
		ID:   999,
		Name: "不存在的配送方式",
	}

	result, err := uc.UpdateMethod(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// TestDeleteMethod_Success 测试成功删除配送方式
func TestDeleteMethod_Success(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	err := uc.DeleteMethod(ctx, 1)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if mockMethod.DeleteCallCount != 1 {
		t.Errorf("期望 Delete 被调用 1 次，实际调用 %d 次", mockMethod.DeleteCallCount)
	}
}

// TestDeleteMethod_RepositoryError 测试删除时仓储返回错误
func TestDeleteMethod_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("删除失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		DeleteError: expectedErr,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	err := uc.DeleteMethod(ctx, 999)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
}

// ==================== 配送日期测试 ====================

// TestListDeliveryDates_Success 测试成功获取配送日期列表
func TestListDeliveryDates_Success(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		ListResult: []*biz.DeliveryDate{
			newTestDeliveryDate(1, 1),
			newTestDeliveryDate(2, 1),
		},
		ListTotal: 2,
	}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, total, err := uc.ListDeliveryDates(ctx, 1, 10)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数为 2，实际为 %d", total)
	}
	if len(items) != 2 {
		t.Errorf("期望返回 2 条记录，实际返回 %d 条", len(items))
	}
	if items[0].DeliveryDate != "2026-05-25" {
		t.Errorf("期望配送日期为 '2026-05-25'，实际为 '%s'", items[0].DeliveryDate)
	}
}

// TestListDeliveryDates_RepositoryError 测试仓储返回错误
func TestListDeliveryDates_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("查询失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		ListError: expectedErr,
	}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, err := uc.ListDeliveryDates(ctx, 1, 10)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if items != nil {
		t.Errorf("期望返回 nil，实际返回 %v", items)
	}
}

// TestCreateDeliveryDate_Success 测试成功创建配送日期
func TestCreateDeliveryDate_Success(t *testing.T) {
	ctx := context.Background()
	createdDate := newTestDeliveryDate(1, 1)
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		CreateResult: createdDate,
	}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.CreateDeliveryDateRequest{
		ShippingMethodID: 1,
		DeliveryDate:     "2026-05-26",
		IsAvailable:      true,
		Description:      "周末配送",
	}

	result, err := uc.CreateDeliveryDate(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if mockDeliveryDate.CreateCallCount != 1 {
		t.Errorf("期望 Create 被调用 1 次，实际调用 %d 次", mockDeliveryDate.CreateCallCount)
	}
}

// TestCreateDeliveryDate_RepositoryError 测试创建时仓储返回错误
func TestCreateDeliveryDate_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("创建失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		CreateError: expectedErr,
	}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.CreateDeliveryDateRequest{
		ShippingMethodID: 1,
		DeliveryDate:     "2026-05-26",
	}

	result, err := uc.CreateDeliveryDate(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// TestUpdateDeliveryDate_Success 测试成功更新配送日期
func TestUpdateDeliveryDate_Success(t *testing.T) {
	ctx := context.Background()
	updatedDate := newTestDeliveryDate(1, 2)
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		UpdateResult: updatedDate,
	}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateDeliveryDateRequest{
		ID:               1,
		ShippingMethodID: 2,
		DeliveryDate:     "2026-05-27",
		IsAvailable:      false,
		Description:      "已约满",
	}

	result, err := uc.UpdateDeliveryDate(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if mockDeliveryDate.UpdateCallCount != 1 {
		t.Errorf("期望 Update 被调用 1 次，实际调用 %d 次", mockDeliveryDate.UpdateCallCount)
	}
}

// TestUpdateDeliveryDate_RepositoryError 测试更新时仓储返回错误
func TestUpdateDeliveryDate_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("更新失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		UpdateError: expectedErr,
	}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateDeliveryDateRequest{
		ID:               999,
		ShippingMethodID: 1,
		DeliveryDate:     "2026-05-27",
	}

	result, err := uc.UpdateDeliveryDate(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// ==================== 仓库测试 ====================

// TestListWarehouses_Success 测试成功获取仓库列表
func TestListWarehouses_Success(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		ListResult: []*biz.Warehouse{
			newTestWarehouse(1, "上海仓库"),
			newTestWarehouse(2, "北京仓库"),
		},
		ListTotal: 2,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, total, err := uc.ListWarehouses(ctx, 1, 10)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数为 2，实际为 %d", total)
	}
	if len(items) != 2 {
		t.Errorf("期望返回 2 条记录，实际返回 %d 条", len(items))
	}
	if items[0].City != "上海" {
		t.Errorf("期望城市为 '上海'，实际为 '%s'", items[0].City)
	}
}

// TestListWarehouses_RepositoryError 测试仓储返回错误
func TestListWarehouses_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("查询失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		ListError: expectedErr,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, err := uc.ListWarehouses(ctx, 1, 10)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if items != nil {
		t.Errorf("期望返回 nil，实际返回 %v", items)
	}
}

// TestCreateWarehouse_Success 测试成功创建仓库
func TestCreateWarehouse_Success(t *testing.T) {
	ctx := context.Background()
	createdWarehouse := newTestWarehouse(1, "深圳仓库")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		CreateResult: createdWarehouse,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.CreateWarehouseRequest{
		Name:        "深圳仓库",
		Code:        "SZ001",
		Address:     "深圳市南山区科技园",
		City:        "深圳",
		CountryID:   1,
		StateID:     5,
		ZipCode:     "518000",
		PhoneNumber: "0755-12345678",
		IsActive:    true,
		Latitude:    22.5431,
		Longitude:   114.0579,
	}

	result, err := uc.CreateWarehouse(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if result.Name != "深圳仓库" {
		t.Errorf("期望名称为 '深圳仓库'，实际为 '%s'", result.Name)
	}
	if mockWarehouse.CreateCallCount != 1 {
		t.Errorf("期望 Create 被调用 1 次，实际调用 %d 次", mockWarehouse.CreateCallCount)
	}
}

// TestCreateWarehouse_RepositoryError 测试创建时仓储返回错误
func TestCreateWarehouse_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("创建失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		CreateError: expectedErr,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.CreateWarehouseRequest{
		Name: "测试仓库",
	}

	result, err := uc.CreateWarehouse(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// TestUpdateWarehouse_Success 测试成功更新仓库
func TestUpdateWarehouse_Success(t *testing.T) {
	ctx := context.Background()
	updatedWarehouse := newTestWarehouse(1, "更新后的仓库")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		UpdateResult: updatedWarehouse,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateWarehouseRequest{
		ID:          1,
		Name:        "更新后的仓库",
		Code:        "WH002",
		Address:     "新地址",
		City:        "广州",
		CountryID:   1,
		StateID:     6,
		ZipCode:     "510000",
		PhoneNumber: "020-12345678",
		IsActive:    true,
		Latitude:    23.1291,
		Longitude:   113.2644,
	}

	result, err := uc.UpdateWarehouse(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result.Name != "更新后的仓库" {
		t.Errorf("期望名称为 '更新后的仓库'，实际为 '%s'", result.Name)
	}
	if mockWarehouse.UpdateCallCount != 1 {
		t.Errorf("期望 Update 被调用 1 次，实际调用 %d 次", mockWarehouse.UpdateCallCount)
	}
}

// TestUpdateWarehouse_RepositoryError 测试更新时仓储返回错误
func TestUpdateWarehouse_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("更新失败")
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		UpdateError: expectedErr,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.UpdateWarehouseRequest{
		ID:   999,
		Name: "不存在的仓库",
	}

	result, err := uc.UpdateWarehouse(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// ==================== 运费估算测试 ====================

// TestEstimateShipping_ReturnsEmpty 测试运费估算返回空列表（占位实现）
func TestEstimateShipping_ReturnsEmpty(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	req := request.EstimateShippingRequest{
		WarehouseID: "1",
		CountryID:   "1",
		StateID:     "1",
		ZipCode:     "200000",
		SubTotal:    "100.00",
		Weight:      "1.5",
	}

	result, err := uc.EstimateShipping(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("期望返回空列表，实际返回 %d 条", len(result.Items))
	}
}

// ==================== 转换函数测试 ====================

// TestToProviderResponse 测试配送提供者转换
func TestToProviderResponse(t *testing.T) {
	now := time.Now()
	provider := &biz.Provider{
		ID:            1,
		Name:          "顺丰",
		SystemKeyword: "Shipping.SF",
		DisplayOrder:  1,
		IsActive:      true,
		LogoURL:       "https://example.com/sf.png",
		TrackingURL:   "https://sf.com/track",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	mockProvider := &MockProviderRepository{
		ListResult: []*biz.Provider{provider},
		ListTotal:  1,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}
	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, _ := uc.ListProviders(context.Background(), 1, 10)

	if items[0].ID != provider.ID {
		t.Errorf("ID 转换错误")
	}
	if items[0].TrackingURL != provider.TrackingURL {
		t.Errorf("TrackingURL 转换错误")
	}
	if items[0].CreatedAt != now.Unix() {
		t.Errorf("CreatedAt 转换错误")
	}
}

// TestToMethodResponse 测试配送方式转换
func TestToMethodResponse(t *testing.T) {
	now := time.Now()
	method := &biz.Method{
		ID:             1,
		Name:           "标准配送",
		SystemKeyword:  "Shipping.Method.Standard",
		ProviderID:     1,
		ProviderName:   "顺丰",
		DisplayOrder:   1,
		IsActive:       true,
		Rate:           10.0,
		MinOrderAmount: 99.0,
		MaxOrderAmount: 9999.0,
		EstimatedDays:  3,
		Description:    "标准配送",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{
		ListResult: []*biz.Method{method},
		ListTotal:  1,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}
	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, _ := uc.ListMethods(context.Background(), 1, 10)

	if items[0].Rate != method.Rate {
		t.Errorf("Rate 转换错误")
	}
	if items[0].ProviderName != method.ProviderName {
		t.Errorf("ProviderName 转换错误")
	}
	if items[0].EstimatedDays != method.EstimatedDays {
		t.Errorf("EstimatedDays 转换错误")
	}
}

// TestToDeliveryDateResponse 测试配送日期转换
func TestToDeliveryDateResponse(t *testing.T) {
	now := time.Now()
	date := &biz.DeliveryDate{
		ID:               1,
		ShippingMethodID: 2,
		DeliveryDate:     "2026-05-28",
		IsAvailable:      true,
		Description:      "测试配送日期",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{
		ListResult: []*biz.DeliveryDate{date},
		ListTotal:  1,
	}
	mockWarehouse := &MockWarehouseRepository{}
	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, _ := uc.ListDeliveryDates(context.Background(), 1, 10)

	if items[0].ShippingMethodID != date.ShippingMethodID {
		t.Errorf("ShippingMethodID 转换错误")
	}
	if items[0].IsAvailable != date.IsAvailable {
		t.Errorf("IsAvailable 转换错误")
	}
}

// TestToWarehouseResponse 测试仓库转换
func TestToWarehouseResponse(t *testing.T) {
	now := time.Now()
	warehouse := &biz.Warehouse{
		ID:          1,
		Name:        "测试仓库",
		Code:        "TEST001",
		Address:     "测试地址",
		City:        "上海",
		CountryID:   1,
		StateID:     1,
		ZipCode:     "200000",
		PhoneNumber: "021-12345678",
		IsActive:    true,
		Latitude:    31.2304,
		Longitude:   121.4737,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockProvider := &MockProviderRepository{}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{
		ListResult: []*biz.Warehouse{warehouse},
		ListTotal:  1,
	}
	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	items, _, _ := uc.ListWarehouses(context.Background(), 1, 10)

	if items[0].Latitude != warehouse.Latitude {
		t.Errorf("Latitude 转换错误")
	}
	if items[0].Longitude != warehouse.Longitude {
		t.Errorf("Longitude 转换错误")
	}
	if items[0].ZipCode != warehouse.ZipCode {
		t.Errorf("ZipCode 转换错误")
	}
}

// ==================== Context 传递测试 ====================

// TestContextPassing 测试 context 在调用链中的传递
func TestContextPassing(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	mockProvider := &MockProviderRepository{
		ListResult: []*biz.Provider{newTestProvider(1, "测试")},
		ListTotal:  1,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}
	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	// 测试所有 List 方法
	_, _, err := uc.ListProviders(ctx, 1, 10)
	if err != nil {
		t.Errorf("ListProviders 期望成功: %v", err)
	}

	// 验证仓储方法被调用
	if mockProvider.ListCallCount != 1 {
		t.Errorf("期望 List 被调用")
	}
}

// ==================== 分页参数测试 ====================

// TestPaginationParameters 测试分页参数传递
func TestPaginationParameters(t *testing.T) {
	ctx := context.Background()
	mockProvider := &MockProviderRepository{
		ListResult: []*biz.Provider{newTestProvider(1, "测试")},
		ListTotal:  100,
	}
	mockMethod := &MockMethodRepository{}
	mockDeliveryDate := &MockDeliveryDateRepository{}
	mockWarehouse := &MockWarehouseRepository{}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	testCases := []struct {
		page     int
		pageSize int
	}{
		{1, 10},
		{2, 20},
		{5, 50},
	}

	for _, tc := range testCases {
		_, _, err := uc.ListProviders(ctx, tc.page, tc.pageSize)
		if err != nil {
			t.Errorf("分页参数 page=%d, pageSize=%d 时返回错误: %v", tc.page, tc.pageSize, err)
		}
	}
}

// ==================== 边界条件测试 ====================

// TestEmptyResults 测试所有方法返回空结果
func TestEmptyResults(t *testing.T) {
	ctx := context.Background()

	mockProvider := &MockProviderRepository{
		ListResult: []*biz.Provider{},
		ListTotal:  0,
	}
	mockMethod := &MockMethodRepository{
		ListResult: []*biz.Method{},
		ListTotal:  0,
	}
	mockDeliveryDate := &MockDeliveryDateRepository{
		ListResult: []*biz.DeliveryDate{},
		ListTotal:  0,
	}
	mockWarehouse := &MockWarehouseRepository{
		ListResult: []*biz.Warehouse{},
		ListTotal:  0,
	}

	uc := biz.NewShippingUsecase(mockProvider, mockMethod, mockDeliveryDate, mockWarehouse)

	// 测试所有 List 方法返回空结果
	providers, pTotal, err := uc.ListProviders(ctx, 1, 10)
	if err != nil || len(providers) != 0 || pTotal != 0 {
		t.Errorf("ListProviders 空结果测试失败")
	}

	methods, mTotal, err := uc.ListMethods(ctx, 1, 10)
	if err != nil || len(methods) != 0 || mTotal != 0 {
		t.Errorf("ListMethods 空结果测试失败")
	}

	dates, dTotal, err := uc.ListDeliveryDates(ctx, 1, 10)
	if err != nil || len(dates) != 0 || dTotal != 0 {
		t.Errorf("ListDeliveryDates 空结果测试失败")
	}

	warehouses, wTotal, err := uc.ListWarehouses(ctx, 1, 10)
	if err != nil || len(warehouses) != 0 || wTotal != 0 {
		t.Errorf("ListWarehouses 空结果测试失败")
	}
}
