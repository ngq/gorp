// Package validate_test provides fuzz tests for validation service.
//
// 适用场景：
// - 验证 ValidatorService 对各种边界输入的处理稳定性。
// - 发现潜在的 panic 或异常行为。
package validate_test

import (
	"context"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/provider/validate"
)

// FuzzValidateVar fuzzes ValidateVar with arbitrary strings and tags.
//
// FuzzValidateVar 对 ValidateVar 进行模糊测试，使用任意字符串和标签。
func FuzzValidateVar(f *testing.F) {
	// Create validator service
	// 创建验证服务
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: true,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	// Seed corpus with typical inputs
	f.Add("test@example.com", "required,email")
	f.Add("", "required")
	f.Add("hello", "min=3,max=10")
	f.Add("12345", "numeric")
	f.Add("http://example.com", "url")
	f.Add("192.168.1.1", "ip")
	f.Add("abc", "oneof=abc def ghi")
	f.Add("", "omitempty,email")
	f.Add("test", "")

	f.Fuzz(func(t *testing.T, value, tag string) {
		// ValidateVar should not panic for any input
		// ValidateVar 对任何输入都不应该 panic
		err := svc.ValidateVar(context.Background(), value, tag)
		_ = err // Result doesn't matter, we just verify no panic
	})
}

// FuzzValidateVar_Required fuzzes required validation.
//
// FuzzValidateVar_Required 对 required 验证进行模糊测试。
func FuzzValidateVar_Required(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("")
	f.Add("value")
	f.Add("  ")
	f.Add("\t\n")
	f.Add("非空值")

	f.Fuzz(func(t *testing.T, value string) {
		err := svc.ValidateVar(context.Background(), value, "required")
		// Empty values should fail required, non-empty should pass
		// 空值应该失败 required，非空值应该通过
		if value == "" && err == nil {
			t.Error("expected error for empty value with required tag")
		}
	})
}

// FuzzValidateVar_Email fuzzes email validation.
//
// FuzzValidateVar_Email 对 email 验证进行模糊测试。
func FuzzValidateVar_Email(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("test@example.com")
	f.Add("invalid")
	f.Add("@example.com")
	f.Add("test@")
	f.Add("test@example")
	f.Add("test.user+tag@example.co.uk")
	f.Add("中文@例子.中国")
	f.Add("test@example.com\n")
	f.Add("test@example.com\r\n")

	f.Fuzz(func(t *testing.T, email string) {
		err := svc.ValidateVar(context.Background(), email, "email")
		_ = err // Just verify no panic
	})
}

// FuzzValidateVar_MinMax fuzzes min/max validation.
//
// FuzzValidateVar_MinMax 对 min/max 验证进行模糊测试。
func FuzzValidateVar_MinMax(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("abc")
	f.Add("")
	f.Add("abcdefghijk")
	f.Add("a")

	f.Fuzz(func(t *testing.T, value string) {
		err := svc.ValidateVar(context.Background(), value, "min=1,max=100")
		_ = err // Just verify no panic
	})
}

// FuzzValidateVar_Numeric fuzzes numeric validation.
//
// FuzzValidateVar_Numeric 对 numeric 验证进行模糊测试。
func FuzzValidateVar_Numeric(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("12345")
	f.Add("-123")
	f.Add("3.14")
	f.Add("abc")
	f.Add("12abc34")
	f.Add("")
	f.Add("0")

	f.Fuzz(func(t *testing.T, value string) {
		err := svc.ValidateVar(context.Background(), value, "numeric")
		_ = err // Just verify no panic
	})
}

// FuzzValidateVar_URL fuzzes URL validation.
//
// FuzzValidateVar_URL 对 URL 验证进行模糊测试。
func FuzzValidateVar_URL(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("http://example.com")
	f.Add("https://example.com/path?query=1")
	f.Add("ftp://ftp.example.com")
	f.Add("invalid-url")
	f.Add("http://")
	f.Add("://missing-scheme.com")
	f.Add("http://example.com:8080/path")
	f.Add("http://中文域名.中国")

	f.Fuzz(func(t *testing.T, url string) {
		err := svc.ValidateVar(context.Background(), url, "url")
		_ = err // Just verify no panic
	})
}

// FuzzValidateVar_IP fuzzes IP validation.
//
// FuzzValidateVar_IP 对 IP 验证进行模糊测试。
func FuzzValidateVar_IP(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("192.168.1.1")
	f.Add("10.0.0.1")
	f.Add("255.255.255.255")
	f.Add("::1")
	f.Add("2001:db8::1")
	f.Add("invalid-ip")
	f.Add("256.256.256.256")
	f.Add("")

	f.Fuzz(func(t *testing.T, ip string) {
		err := svc.ValidateVar(context.Background(), ip, "ip")
		_ = err // Just verify no panic
	})
}

// FuzzValidateVar_OneOf fuzzes oneof validation.
//
// FuzzValidateVar_OneOf 对 oneof 验证进行模糊测试。
func FuzzValidateVar_OneOf(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("apple")
	f.Add("banana")
	f.Add("cherry")
	f.Add("invalid")
	f.Add("")
	f.Add("apple banana") // contains space

	f.Fuzz(func(t *testing.T, value string) {
		err := svc.ValidateVar(context.Background(), value, "oneof=apple banana cherry")
		_ = err // Just verify no panic
	})
}

// FuzzSetLocale fuzzes SetLocale with arbitrary strings.
//
// FuzzSetLocale 对 SetLocale 进行模糊测试，使用任意字符串。
func FuzzSetLocale(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: true,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("en")
	f.Add("zh")
	f.Add("fr")
	f.Add("")
	f.Add("EN")
	f.Add("ZH")
	f.Add("invalid-locale")

	f.Fuzz(func(t *testing.T, locale string) {
		err := svc.SetLocale(locale)
		// Only "en" and "zh" are supported
		// 只有 "en" 和 "zh" 被支持
		if err == nil && locale != "en" && locale != "zh" {
			t.Logf("SetLocale(%q) succeeded unexpectedly", locale)
		}
	})
}

// TestStruct for struct validation fuzz testing.
type fuzzTestStruct struct {
	Name  string `validate:"required" json:"name"`
	Email string `validate:"required,email" json:"email"`
	Age   int    `validate:"min=0,max=150" json:"age"`
}

// FuzzValidateStruct fuzzes struct validation.
//
// FuzzValidateStruct 对结构体验证进行模糊测试。
func FuzzValidateStruct(f *testing.F) {
	svc, err := validate.NewValidatorService(&datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: false,
	})
	if err != nil {
		f.Fatalf("failed to create validator service: %v", err)
	}

	f.Add("John Doe", "john@example.com", 25)
	f.Add("", "invalid-email", -5)
	f.Add("A", "test@test.com", 200)
	f.Add("Very Long Name That Might Cause Issues", "email@domain", 0)

	f.Fuzz(func(t *testing.T, name, email string, age int) {
		obj := &fuzzTestStruct{
			Name:  name,
			Email: email,
			Age:   age,
		}
		err := svc.Validate(context.Background(), obj)
		_ = err // Just verify no panic
	})
}
