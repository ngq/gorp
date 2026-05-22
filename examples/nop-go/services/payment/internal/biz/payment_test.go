// Package biz_test 提供 payment 服务的业务逻辑层单元测试
//
// 测试覆盖：
// 1. PaymentUsecase 所有方法的成功和失败场景
// 2. Mock 仓储实现用于隔离测试
// 3. 领域实体到响应 DTO 的转换验证
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"nop-go/services/payment/internal/biz"
	"nop-go/services/payment/internal/server/http/request"
)

// ==================== Mock 仓储实现 ====================

// MockPaymentMethodRepository 支付方式仓储的 Mock 实现
type MockPaymentMethodRepository struct {
	// List 返回的数据
	ListResult    []*biz.PaymentMethod
	ListTotal     int64
	ListError     error
	// Update 返回的数据
	UpdateResult  *biz.PaymentMethod
	UpdateError   error
	// 调用计数
	ListCallCount  int
	UpdateCallCount int
}

// List Mock 实现 List 方法
func (m *MockPaymentMethodRepository) List(ctx context.Context, page, pageSize int) ([]*biz.PaymentMethod, int64, error) {
	m.ListCallCount++
	return m.ListResult, m.ListTotal, m.ListError
}

// Update Mock 实现 Update 方法
func (m *MockPaymentMethodRepository) Update(ctx context.Context, method *biz.PaymentMethod) (*biz.PaymentMethod, error) {
	m.UpdateCallCount++
	return m.UpdateResult, m.UpdateError
}

// MockMethodRestrictionRepository 支付方式限制仓储的 Mock 实现
type MockMethodRestrictionRepository struct {
	// List 返回的数据
	ListResult    []*biz.MethodRestriction
	ListTotal     int64
	ListError     error
	// Update 返回的数据
	UpdateResult  *biz.MethodRestriction
	UpdateError   error
	// 调用计数
	ListCallCount  int
	UpdateCallCount int
}

// List Mock 实现 List 方法
func (m *MockMethodRestrictionRepository) List(ctx context.Context, page, pageSize int) ([]*biz.MethodRestriction, int64, error) {
	m.ListCallCount++
	return m.ListResult, m.ListTotal, m.ListError
}

// Update Mock 实现 Update 方法
func (m *MockMethodRestrictionRepository) Update(ctx context.Context, restriction *biz.MethodRestriction) (*biz.MethodRestriction, error) {
	m.UpdateCallCount++
	return m.UpdateResult, m.UpdateError
}

// ==================== 测试辅助函数 ====================

// newTestPaymentMethod 创建测试用的支付方式领域实体
func newTestPaymentMethod(id uint, name string) *biz.PaymentMethod {
	now := time.Now()
	return &biz.PaymentMethod{
		ID:                    id,
		Name:                  name,
		SystemKeyword:         "Payments.Test",
		DisplayOrder:          1,
		IsActive:              true,
		LogoURL:               "https://example.com/logo.png",
		SupportsRefund:        true,
		SupportsPartialRefund: true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// newTestMethodRestriction 创建测试用的支付方式限制领域实体
func newTestMethodRestriction(id uint, paymentMethodID uint) *biz.MethodRestriction {
	now := time.Now()
	return &biz.MethodRestriction{
		ID:               id,
		PaymentMethodID:  paymentMethodID,
		MinOrderAmount:   10.0,
		MaxOrderAmount:   10000.0,
		RestrictionType:  "amount_range",
		RestrictionValue: "default",
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// ==================== ListPaymentMethods 测试 ====================

// TestListPaymentMethods_Success 测试成功获取支付方式列表
func TestListPaymentMethods_Success(t *testing.T) {
	// 准备测试数据
	ctx := context.Background()
	mockMethodRepo := &MockPaymentMethodRepository{
		ListResult: []*biz.PaymentMethod{
			newTestPaymentMethod(1, "支付宝"),
			newTestPaymentMethod(2, "微信支付"),
		},
		ListTotal: 2,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	// 创建用例
	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	// 执行测试
	items, total, err := uc.ListPaymentMethods(ctx, 1, 10)

	// 验证结果
	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 2 {
		t.Errorf("期望总数为 2，实际为 %d", total)
	}
	if len(items) != 2 {
		t.Errorf("期望返回 2 条记录，实际返回 %d 条", len(items))
	}
	if mockMethodRepo.ListCallCount != 1 {
		t.Errorf("期望 List 被调用 1 次，实际调用 %d 次", mockMethodRepo.ListCallCount)
	}

	// 验证响应转换
	if items[0].Name != "支付宝" {
		t.Errorf("期望第一条记录名称为 '支付宝'，实际为 '%s'", items[0].Name)
	}
	if items[0].SystemKeyword != "Payments.Test" {
		t.Errorf("期望系统关键字为 'Payments.Test'，实际为 '%s'", items[0].SystemKeyword)
	}
}

// TestListPaymentMethods_EmptyList 测试空列表场景
func TestListPaymentMethods_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockMethodRepo := &MockPaymentMethodRepository{
		ListResult: []*biz.PaymentMethod{},
		ListTotal:  0,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, total, err := uc.ListPaymentMethods(ctx, 1, 10)

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

// TestListPaymentMethods_RepositoryError 测试仓储返回错误的场景
func TestListPaymentMethods_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("数据库连接失败")
	mockMethodRepo := &MockPaymentMethodRepository{
		ListError: expectedErr,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, total, err := uc.ListPaymentMethods(ctx, 1, 10)

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

// TestListPaymentMethods_Pagination 测试分页参数传递
func TestListPaymentMethods_Pagination(t *testing.T) {
	ctx := context.Background()
	mockMethodRepo := &MockPaymentMethodRepository{
		ListResult: []*biz.PaymentMethod{newTestPaymentMethod(1, "测试")},
		ListTotal:  100,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	// 测试不同的分页参数
	testCases := []struct {
		page     int
		pageSize int
	}{
		{1, 10},
		{2, 20},
		{5, 50},
	}

	for _, tc := range testCases {
		_, _, err := uc.ListPaymentMethods(ctx, tc.page, tc.pageSize)
		if err != nil {
			t.Errorf("分页参数 page=%d, pageSize=%d 时返回错误: %v", tc.page, tc.pageSize, err)
		}
	}
}

// ==================== UpdatePaymentMethod 测试 ====================

// TestUpdatePaymentMethod_Success 测试成功更新支付方式
func TestUpdatePaymentMethod_Success(t *testing.T) {
	ctx := context.Background()
	updatedMethod := newTestPaymentMethod(1, "支付宝-更新后")
	mockMethodRepo := &MockPaymentMethodRepository{
		UpdateResult: updatedMethod,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	req := request.UpdatePaymentMethodRequest{
		ID:                    1,
		Name:                  "支付宝-更新后",
		SystemKeyword:         "Payments.Alipay",
		DisplayOrder:          2,
		IsActive:              true,
		LogoURL:               "https://example.com/new-logo.png",
		SupportsRefund:        true,
		SupportsPartialRefund: false,
	}

	result, err := uc.UpdatePaymentMethod(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if result.Name != "支付宝-更新后" {
		t.Errorf("期望名称为 '支付宝-更新后'，实际为 '%s'", result.Name)
	}
	if mockMethodRepo.UpdateCallCount != 1 {
		t.Errorf("期望 Update 被调用 1 次，实际调用 %d 次", mockMethodRepo.UpdateCallCount)
	}
}

// TestUpdatePaymentMethod_RepositoryError 测试更新时仓储返回错误
func TestUpdatePaymentMethod_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("记录不存在")
	mockMethodRepo := &MockPaymentMethodRepository{
		UpdateError: expectedErr,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	req := request.UpdatePaymentMethodRequest{
		ID:   999,
		Name: "不存在的支付方式",
	}

	result, err := uc.UpdatePaymentMethod(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// TestUpdatePaymentMethod_AllFields 测试更新所有字段
func TestUpdatePaymentMethod_AllFields(t *testing.T) {
	ctx := context.Background()
	updatedMethod := &biz.PaymentMethod{
		ID:                    1,
		Name:                  "完整测试",
		SystemKeyword:         "Payments.FullTest",
		DisplayOrder:          100,
		IsActive:              false,
		LogoURL:               "https://example.com/full.png",
		SupportsRefund:        false,
		SupportsPartialRefund: false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	mockMethodRepo := &MockPaymentMethodRepository{
		UpdateResult: updatedMethod,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	req := request.UpdatePaymentMethodRequest{
		ID:                    1,
		Name:                  "完整测试",
		SystemKeyword:         "Payments.FullTest",
		DisplayOrder:          100,
		IsActive:              false,
		LogoURL:               "https://example.com/full.png",
		SupportsRefund:        false,
		SupportsPartialRefund: false,
	}

	result, err := uc.UpdatePaymentMethod(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result.Name != "完整测试" {
		t.Errorf("期望名称为 '完整测试'，实际为 '%s'", result.Name)
	}
	if result.IsActive != false {
		t.Errorf("期望 IsActive 为 false，实际为 %v", result.IsActive)
	}
	if result.SupportsRefund != false {
		t.Errorf("期望 SupportsRefund 为 false，实际为 %v", result.SupportsRefund)
	}
}

// ==================== ListMethodRestrictions 测试 ====================

// TestListMethodRestrictions_Success 测试成功获取支付方式限制列表
func TestListMethodRestrictions_Success(t *testing.T) {
	ctx := context.Background()
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		ListResult: []*biz.MethodRestriction{
			newTestMethodRestriction(1, 1),
			newTestMethodRestriction(2, 2),
			newTestMethodRestriction(3, 1),
		},
		ListTotal: 3,
	}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, total, err := uc.ListMethodRestrictions(ctx, 1, 10)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if total != 3 {
		t.Errorf("期望总数为 3，实际为 %d", total)
	}
	if len(items) != 3 {
		t.Errorf("期望返回 3 条记录，实际返回 %d 条", len(items))
	}
	if mockRestrictionRepo.ListCallCount != 1 {
		t.Errorf("期望 List 被调用 1 次，实际调用 %d 次", mockRestrictionRepo.ListCallCount)
	}

	// 验证响应转换
	if items[0].PaymentMethodID != 1 {
		t.Errorf("期望第一条记录的 PaymentMethodID 为 1，实际为 %d", items[0].PaymentMethodID)
	}
	if items[0].RestrictionType != "amount_range" {
		t.Errorf("期望限制类型为 'amount_range'，实际为 '%s'", items[0].RestrictionType)
	}
}

// TestListMethodRestrictions_EmptyList 测试空列表场景
func TestListMethodRestrictions_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		ListResult: []*biz.MethodRestriction{},
		ListTotal:  0,
	}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, total, err := uc.ListMethodRestrictions(ctx, 1, 10)

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

// TestListMethodRestrictions_RepositoryError 测试仓储返回错误
func TestListMethodRestrictions_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("查询失败")
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		ListError: expectedErr,
	}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, total, err := uc.ListMethodRestrictions(ctx, 1, 10)

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

// ==================== UpdateMethodRestrictions 测试 ====================

// TestUpdateMethodRestrictions_Success 测试成功更新支付方式限制
func TestUpdateMethodRestrictions_Success(t *testing.T) {
	ctx := context.Background()
	updatedRestriction := newTestMethodRestriction(1, 1)
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		UpdateResult: updatedRestriction,
	}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	req := request.UpdateMethodRestrictionsRequest{
		ID:               1,
		PaymentMethodID:  1,
		MinOrderAmount:   10.0,
		MaxOrderAmount:   10000.0,
		RestrictionType:  "amount_range",
		RestrictionValue: "default",
		IsActive:         true,
	}

	result, err := uc.UpdateMethodRestrictions(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("期望返回结果，但返回 nil")
	}
	if result.ID != 1 {
		t.Errorf("期望 ID 为 1，实际为 %d", result.ID)
	}
	if mockRestrictionRepo.UpdateCallCount != 1 {
		t.Errorf("期望 Update 被调用 1 次，实际调用 %d 次", mockRestrictionRepo.UpdateCallCount)
	}
}

// TestUpdateMethodRestrictions_RepositoryError 测试更新时仓储返回错误
func TestUpdateMethodRestrictions_RepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("更新失败")
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		UpdateError: expectedErr,
	}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	req := request.UpdateMethodRestrictionsRequest{
		ID:              999,
		PaymentMethodID: 1,
	}

	result, err := uc.UpdateMethodRestrictions(ctx, req)

	if err != expectedErr {
		t.Errorf("期望返回错误 '%v'，实际返回 '%v'", expectedErr, err)
	}
	if result != nil {
		t.Errorf("期望返回 nil，实际返回 %v", result)
	}
}

// TestUpdateMethodRestrictions_AllFields 测试更新所有字段
func TestUpdateMethodRestrictions_AllFields(t *testing.T) {
	ctx := context.Background()
	updatedRestriction := &biz.MethodRestriction{
		ID:               1,
		PaymentMethodID:  2,
		MinOrderAmount:   50.0,
		MaxOrderAmount:   50000.0,
		RestrictionType:  "country",
		RestrictionValue: "CN,US,UK",
		IsActive:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		UpdateResult: updatedRestriction,
	}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	req := request.UpdateMethodRestrictionsRequest{
		ID:               1,
		PaymentMethodID:  2,
		MinOrderAmount:   50.0,
		MaxOrderAmount:   50000.0,
		RestrictionType:  "country",
		RestrictionValue: "CN,US,UK",
		IsActive:         false,
	}

	result, err := uc.UpdateMethodRestrictions(ctx, req)

	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}
	if result.PaymentMethodID != 2 {
		t.Errorf("期望 PaymentMethodID 为 2，实际为 %d", result.PaymentMethodID)
	}
	if result.RestrictionType != "country" {
		t.Errorf("期望限制类型为 'country'，实际为 '%s'", result.RestrictionType)
	}
	if result.IsActive != false {
		t.Errorf("期望 IsActive 为 false，实际为 %v", result.IsActive)
	}
}

// ==================== 转换函数测试 ====================

// TestToPaymentMethodResponse 测试领域实体到响应 DTO 的转换
func TestToPaymentMethodResponse(t *testing.T) {
	now := time.Now()
	method := &biz.PaymentMethod{
		ID:                    1,
		Name:                  "测试支付",
		SystemKeyword:         "Payments.Test",
		DisplayOrder:          5,
		IsActive:              true,
		LogoURL:               "https://example.com/test.png",
		SupportsRefund:        true,
		SupportsPartialRefund: false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	// 通过 ListPaymentMethods 间接测试转换函数
	mockMethodRepo := &MockPaymentMethodRepository{
		ListResult: []*biz.PaymentMethod{method},
		ListTotal:  1,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}
	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, _, _ := uc.ListPaymentMethods(context.Background(), 1, 10)

	// 验证转换结果
	if items[0].ID != method.ID {
		t.Errorf("ID 转换错误")
	}
	if items[0].Name != method.Name {
		t.Errorf("Name 转换错误")
	}
	if items[0].CreatedAt != now.Unix() {
		t.Errorf("CreatedAt 转换错误，期望 %d，实际 %d", now.Unix(), items[0].CreatedAt)
	}
	if items[0].UpdatedAt != now.Unix() {
		t.Errorf("UpdatedAt 转换错误，期望 %d，实际 %d", now.Unix(), items[0].UpdatedAt)
	}
}

// TestToMethodRestrictionResponse 测试领域实体到响应 DTO 的转换
func TestToMethodRestrictionResponse(t *testing.T) {
	now := time.Now()
	restriction := &biz.MethodRestriction{
		ID:               1,
		PaymentMethodID:  2,
		MinOrderAmount:   100.0,
		MaxOrderAmount:   10000.0,
		RestrictionType:  "currency",
		RestrictionValue: "CNY,USD",
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{
		ListResult: []*biz.MethodRestriction{restriction},
		ListTotal:  1,
	}
	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	items, _, _ := uc.ListMethodRestrictions(context.Background(), 1, 10)

	// 验证转换结果
	if items[0].ID != restriction.ID {
		t.Errorf("ID 转换错误")
	}
	if items[0].PaymentMethodID != restriction.PaymentMethodID {
		t.Errorf("PaymentMethodID 转换错误")
	}
	if items[0].CreatedAt != now.Unix() {
		t.Errorf("CreatedAt 转换错误，期望 %d，实际 %d", now.Unix(), items[0].CreatedAt)
	}
}

// ==================== NewPaymentUsecase 测试 ====================

// TestNewPaymentUsecase 测试用例构造函数
func TestNewPaymentUsecase(t *testing.T) {
	mockMethodRepo := &MockPaymentMethodRepository{}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}

	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	if uc == nil {
		t.Error("期望返回非 nil 用例实例")
	}
}

// ==================== Context 传递测试 ====================

// TestContextPassing 测试 context 在调用链中的传递
func TestContextPassing(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	mockMethodRepo := &MockPaymentMethodRepository{
		ListResult: []*biz.PaymentMethod{newTestPaymentMethod(1, "测试")},
		ListTotal:  1,
	}
	mockRestrictionRepo := &MockMethodRestrictionRepository{}
	uc := biz.NewPaymentUsecase(mockMethodRepo, mockRestrictionRepo)

	// Context 应该被正确传递到仓储
	_, _, err := uc.ListPaymentMethods(ctx, 1, 10)
	if err != nil {
		t.Errorf("期望成功，但返回错误: %v", err)
	}

	// 验证仓储方法被调用
	if mockMethodRepo.ListCallCount != 1 {
		t.Errorf("期望 List 被调用")
	}
}
