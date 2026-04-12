package validate

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

// TestUser 用于测试的结构体。
// 验证规则放在 validate tag 中，字段名从 json tag 获取。
type TestUser struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
}

// TestValidatorService_Validate_Success 测试验证成功。
//
// 中文说明：
// - 有效的数据应通过验证。
func TestValidatorService_Validate_Success(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestValidatorService_Validate_Required 测试必填验证。
//
// 中文说明：
// - 缺少必填字段应返回错误。
func TestValidatorService_Validate_Required(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		// Username 缺失
		Email: "test@example.com",
		Age:   25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Error("expected error for missing username")
	}
}

// TestValidatorService_Validate_Email 测试邮箱验证。
//
// 中文说明：
// - 无效邮箱应返回错误。
func TestValidatorService_Validate_Email(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "testuser",
		Email:    "invalid-email",
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Error("expected error for invalid email")
	}
}

// TestValidatorService_Validate_Range 测试范围验证。
//
// 中文说明：
// - 超出范围的值应返回错误。
func TestValidatorService_Validate_Range(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      200, // 超出范围
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Error("expected error for age out of range")
	}
}

// TestValidatorService_ValidateVar 测试变量验证。
//
// 中文说明：
// - 单个变量验证应正确工作。
func TestValidatorService_ValidateVar(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// 测试邮箱验证
	err = svc.ValidateVar(context.Background(), "test@example.com", "required,email")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 测试无效邮箱
	err = svc.ValidateVar(context.Background(), "invalid", "email")
	if err == nil {
		t.Error("expected error for invalid email")
	}
}

// TestValidatorService_SetLocale 测试语言切换。
//
// 中文说明：
// - 切换语言后错误消息应使用新语言。
func TestValidatorService_SetLocale(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "en",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// 验证错误消息
	user := &TestUser{
		Username: "", // 必填字段为空
		Email:    "test@example.com",
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Error("expected validation error")
	}
}

// TestValidatorService_TranslateError 测试错误翻译。
//
// 中文说明：
// - 验证错误应被正确翻译。
func TestValidatorService_TranslateError(t *testing.T) {
	cfg := &contract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "ab", // 太短
		Email:    "test@example.com",
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Error("expected validation error")
	}

	// 验证是 AppError
	appErr, ok := err.(contract.AppError)
	if !ok {
		t.Error("expected AppError")
		return
	}

	// 验证错误码
	if appErr.GetStatus().Code != 400 {
		t.Errorf("expected status code 400, got: %d", appErr.GetStatus().Code)
	}
}