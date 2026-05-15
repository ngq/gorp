// Package validate_test provides unit tests for validation service and tag-based rules.
//
// 适用场景：
// - 验证 validate service 对 struct tag 规则（required、email、min、max）的解析和执行。
// - 确保错误消息和字段名返回正确。
// - 确保 JSON tag 字段名在验证错误中正确返回。
// - 确保国际化切换正常工作。
// - 确保自定义校验器可注册并正确执行。
// - 确保错误输出格式一致性。
package validate

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// TestUser 用于测试的结构体。
// 验证规则放在 validate tag 中，当前错误字段名默认使用结构体字段名。
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
	cfg := &datacontract.ValidatorConfig{
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
	cfg := &datacontract.ValidatorConfig{
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
	cfg := &datacontract.ValidatorConfig{
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
	cfg := &datacontract.ValidatorConfig{
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
	cfg := &datacontract.ValidatorConfig{
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
	cfg := &datacontract.ValidatorConfig{
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
	cfg := &datacontract.ValidatorConfig{
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
	var appErr resiliencecontract.AppError
	if !errors.As(err, &appErr) {
		t.Error("expected AppError")
		return
	}

	// 验证错误码
	if appErr.GetStatus().Code != 400 {
		t.Errorf("expected status code 400, got: %d", appErr.GetStatus().Code)
	}
}

// ===================== 新增测试：字段名一致性 =====================

// TestValidatorService_JSONTagNameConsistency 测试 JSON tag 字段名一致性。
//
// 中文说明：
// - 验证错误中应返回 JSON tag 名（如 username）而非 Go 结构体字段名（如 Username）。
// - 这是常见坑：如果不用 RegisterTagNameFunc，错误消息中会显示 Go 字段名，与前端 JSON 不一致。
func TestValidatorService_JSONTagNameConsistency(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "", // 必填字段为空
		Email:    "invalid-email",
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr resiliencecontract.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}

	status := appErr.GetStatus()

	// 错误消息中应包含 JSON tag 名 "username" 和 "email"，而非 "Username" 和 "Email"
	msg := status.Message
	if !strings.Contains(msg, "username") {
		t.Errorf("expected error message to contain JSON tag name 'username', got: %s", msg)
	}
	if strings.Contains(msg, "Username") {
		t.Errorf("error message should use JSON tag 'username' not Go field 'Username', got: %s", msg)
	}

	// metadata 中应有 validation_errors 详情
	if status.Metadata == nil {
		t.Fatal("expected metadata with validation_errors")
	}
	detailsJSON, ok := status.Metadata["validation_errors"]
	if !ok {
		t.Fatal("expected validation_errors in metadata")
	}

	var details []datacontract.ValidationError
	if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
		t.Fatalf("failed to unmarshal validation errors: %v", err)
	}

	// 每个 ValidationError 的 Field 应该是 JSON tag 名
	for _, d := range details {
		switch d.Field {
		case "username", "email":
			// 正确：使用了 JSON tag 名
		case "Username", "Email":
			t.Errorf("ValidationError.Field should use JSON tag name, got Go field name: %s", d.Field)
		default:
			t.Errorf("unexpected field name in validation error: %s", d.Field)
		}
	}
}

// TestValidatorService_NoJSONTagFallsBackToGoName 测试没有 json tag 时回退到 Go 字段名。
//
// 中文说明：
// - 如果字段没有 json tag 或 json tag 为 "-"，应回退到 Go 结构体字段名。
func TestValidatorService_NoJSONTagFallsBackToGoName(t *testing.T) {
	type NoJSONTagStruct struct {
		Name  string `validate:"required"`          // 无 json tag
		Inner string `json:"-" validate:"required"` // json tag 为 "-"
	}

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	obj := &NoJSONTagStruct{} // 两个字段都为空
	err = svc.Validate(context.Background(), obj)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr resiliencecontract.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}

	status := appErr.GetStatus()
	detailsJSON, ok := status.Metadata["validation_errors"]
	if !ok {
		t.Fatal("expected validation_errors in metadata")
	}

	var details []datacontract.ValidationError
	if err := json.Unmarshal([]byte(detailsJSON), &details); err != nil {
		t.Fatalf("failed to unmarshal validation errors: %v", err)
	}

	fieldSet := make(map[string]bool)
	for _, d := range details {
		fieldSet[d.Field] = true
	}

	// 无 json tag 应回退到 Go 字段名 "Name"
	if !fieldSet["Name"] {
		t.Errorf("expected field 'Name' (no json tag fallback), got fields: %v", fieldSet)
	}
	// json:"-" 应回退到 Go 字段名 "Inner"
	if !fieldSet["Inner"] {
		t.Errorf("expected field 'Inner' (json:'-' fallback), got fields: %v", fieldSet)
	}
}

// ===================== 新增测试：错误格式一致性 =====================

// TestValidatorService_ErrorFormatConsistency 测试错误输出格式一致性。
//
// 中文说明：
// - 验证失败时必须返回 AppError，status code 为 400，reason 为 BAD_REQUEST。
// - metadata 中必须包含 validation_errors（JSON 数组）和 error_count。
func TestValidatorService_ErrorFormatConsistency(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "",         // 缺失
		Email:    "not-email", // 无效
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr resiliencecontract.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError type")
	}

	status := appErr.GetStatus()

	// 1. 状态码必须是 400
	if status.Code != 400 {
		t.Errorf("expected code 400, got: %d", status.Code)
	}

	// 2. reason 必须是 BAD_REQUEST
	if status.Reason != resiliencecontract.ErrorReasonBadRequest {
		t.Errorf("expected reason BAD_REQUEST, got: %s", status.Reason)
	}

	// 3. 必须有 validation_errors metadata
	if status.Metadata == nil {
		t.Fatal("expected non-nil metadata")
	}
	if _, ok := status.Metadata["validation_errors"]; !ok {
		t.Error("expected 'validation_errors' key in metadata")
	}

	// 4. 必须有 error_count metadata
	if _, ok := status.Metadata["error_count"]; !ok {
		t.Error("expected 'error_count' key in metadata")
	}

	// 5. error_count 应该是 2（Username 和 Email 都校验失败）
	if status.Metadata["error_count"] != "2" {
		t.Errorf("expected error_count=2, got: %s", status.Metadata["error_count"])
	}
}

// TestValidatorService_TranslateErrorsFalse 测试 TranslateErrors=false 时返回原始英文错误。
//
// 中文说明：
// - 当 TranslateErrors 为 false 时，应返回原始英文错误不做翻译。
// - 此时 metadata 中不包含 validation_errors 详情，性能更好。
func TestValidatorService_TranslateErrorsFalse(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: false,
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{
		Username: "", // 缺失
		Email:    "test@example.com",
		Age:      25,
	}

	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr resiliencecontract.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError type")
	}

	status := appErr.GetStatus()

	// TranslateErrors=false 时，消息是原始英文格式
	if !strings.Contains(status.Message, "required") {
		t.Errorf("expected raw English error message with 'required', got: %s", status.Message)
	}

	// TranslateErrors=false 时，metadata 不含 validation_errors
	if status.Metadata != nil {
		if _, ok := status.Metadata["validation_errors"]; ok {
			t.Error("expected no validation_errors in metadata when TranslateErrors=false")
		}
	}
}

// ===================== 新增测试：国际化切换 =====================

// TestValidatorService_LocaleSwitch_ZhToEn 测试从中文切换到英文。
//
// 中文说明：
// - 初始使用中文，切换到英文后错误消息应变为英文。
func TestValidatorService_LocaleSwitch_ZhToEn(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// 用中文验证，错误消息包含中文
	user := &TestUser{Email: "test@example.com", Age: 25}
	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var zhAppErr resiliencecontract.AppError
	if !errors.As(err, &zhAppErr) {
		t.Fatal("expected AppError type")
	}
	zhMsg := zhAppErr.GetStatus().Message

	// 切换到英文
	if err := svc.SetLocale("en"); err != nil {
		t.Fatalf("failed to set locale: %v", err)
	}

	// 再次验证，错误消息应变为英文
	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var enAppErr resiliencecontract.AppError
	if !errors.As(err, &enAppErr) {
		t.Fatal("expected AppError type")
	}
	enMsg := enAppErr.GetStatus().Message

	// 两条消息不应该相同（中文和英文翻译不同）
	if zhMsg == enMsg {
		t.Errorf("expected different messages for zh/en, both: %s", zhMsg)
	}
}

// TestValidatorService_LocaleSwitch_UnsupportedLocale 测试切换到不支持的语言。
//
// 中文说明：
// - 切换到不支持的语言应返回错误。
func TestValidatorService_LocaleSwitch_UnsupportedLocale(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	err = svc.SetLocale("xx")
	if err == nil {
		t.Error("expected error for unsupported locale")
	}
}

// TestValidatorService_LocaleSwitch_EnDefault 测试英文为默认翻译。
//
// 中文说明：
// - 使用 en locale 初始化时，错误消息应为英文。
func TestValidatorService_LocaleSwitch_EnDefault(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: true,
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	user := &TestUser{Email: "test@example.com", Age: 25}
	err = svc.Validate(context.Background(), user)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr resiliencecontract.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError type")
	}
	msg := appErr.GetStatus().Message

	// 英文翻译应包含 "required" 关键字
	if !strings.Contains(strings.ToLower(msg), "required") {
		t.Errorf("expected English error message containing 'required', got: %s", msg)
	}
}

// ===================== 新增测试：自定义校验器 =====================

// TestValidatorService_CustomValidator 测试自定义校验器注册与使用。
//
// 中文说明：
// - 通过 CustomRules 配置注册自定义校验规则（如手机号）。
// - 自定义规则应在 Validate 中正确触发。
func TestValidatorService_CustomValidator(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale:  "zh",
		Enabled: true,
		CustomRules: map[string]datacontract.CustomRuleConfig{
			"mobile": {
				Name: "mobile",
				Fn: func(ctx context.Context, field any) bool {
					s, ok := field.(string)
					if !ok {
						return false
					}
					// 简化手机号校验：1 开头 + 10 位数字
					return len(s) == 11 && s[0] == '1'
				},
			},
		},
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type MobileReq struct {
		Phone string `json:"phone" validate:"required,mobile"`
	}

	// 有效手机号
	validReq := &MobileReq{Phone: "13800138000"}
	err = svc.Validate(context.Background(), validReq)
	if err != nil {
		t.Errorf("expected valid mobile to pass, got error: %v", err)
	}

	// 无效手机号
	invalidReq := &MobileReq{Phone: "abc"}
	err = svc.Validate(context.Background(), invalidReq)
	if err == nil {
		t.Error("expected invalid mobile to fail validation")
	}
}

// TestValidatorService_RegisterCustom 测试运行时动态注册自定义校验器。
//
// 中文说明：
// - 通过 RegisterCustom 方法在运行时注册自定义校验规则。
// - 注册后应立即生效。
func TestValidatorService_RegisterCustom(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{
		Locale: "zh",
	}

	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// 运行时注册自定义校验规则
	err = svc.RegisterCustom("positive", func(ctx context.Context, field any) bool {
		n, ok := field.(int)
		if !ok {
			return false
		}
		return n > 0
	})
	if err != nil {
		t.Fatalf("failed to register custom validator: %v", err)
	}

	type ScoreReq struct {
		Score int `json:"score" validate:"positive"`
	}

	// 有效值
	err = svc.Validate(context.Background(), &ScoreReq{Score: 10})
	if err != nil {
		t.Errorf("expected positive score to pass, got: %v", err)
	}

	// 无效值
	err = svc.Validate(context.Background(), &ScoreReq{Score: -1})
	if err == nil {
		t.Error("expected negative score to fail validation")
	}
}

// ===================== 新增测试：Provider 层 =====================

// TestProvider_Contract 测试 validate Provider 契约。
//
// 中文说明：
// - Provider 名称应为 "validate"。
// - IsDefer 应为 true（延迟初始化）。
// - Provides 应返回 ValidatorKey。
func TestProvider_Contract(t *testing.T) {
	p := NewProvider()
	if p.Name() != "validate" {
		t.Errorf("expected provider name 'validate', got: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Error("expected IsDefer=true")
	}
	keys := p.Provides()
	if len(keys) != 1 || keys[0] != datacontract.ValidatorKey {
		t.Errorf("expected Provides=[%s], got: %v", datacontract.ValidatorKey, keys)
	}
}

// ===================== 新增测试：TranslateError 边界 =====================

// TestValidatorService_TranslateError_Nil 测试 TranslateError 输入 nil。
//
// 中文说明：
// - 传入 nil 应返回 nil。
func TestValidatorService_TranslateError_Nil(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	result := svc.TranslateError(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got: %v", result)
	}
}

// TestValidatorService_TranslateError_NonValidationError 测试 TranslateError 处理非 ValidationErrors 类型的错误。
//
// 中文说明：
// - 传入非 ValidationErrors 类型的错误，应包装为 BadRequest AppError。
func TestValidatorService_TranslateError_NonValidationError(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// 用 context.Canceled 测试非 ValidationErrors 类型的 error
	appErr := svc.TranslateError(context.Canceled)
	if appErr == nil {
		t.Fatal("expected non-nil AppError for non-validation error")
	}
	if appErr.GetStatus().Code != 400 {
		t.Errorf("expected code 400, got: %d", appErr.GetStatus().Code)
	}
}

// ===================== 新增测试：GetValidator / GetTranslator =====================

// TestValidatorService_GetValidator 测试获取底层 validator 实例。
//
// 中文说明：
// - GetValidator 应返回非 nil 的 validator.Validate 实例。
func TestValidatorService_GetValidator(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	v := svc.GetValidator()
	if v == nil {
		t.Error("expected non-nil validator instance")
	}
}

// TestValidatorService_GetTranslator 测试获取翻译器。
//
// 中文说明：
// - GetTranslator 应返回非 nil 的翻译器实例。
func TestValidatorService_GetTranslator(t *testing.T) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, err := NewValidatorService(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	trans := svc.GetTranslator()
	if trans == nil {
		t.Error("expected non-nil translator instance")
	}
}
