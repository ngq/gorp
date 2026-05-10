// Package benchmark 提供 Validator 性能对比测试。
//
// 对比对象：
// - gorp ValidatorService（基于 go-playground/validator/v10）
// - 直接使用 go-playground/validator/v10
// - Kratos validator middleware（简单接口调用）
//
// 运行方式：
//
//	go test ./benchmark/... -bench=ValidatorCompare -benchmem
package benchmark

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/provider/validate"
)

// ============================================================
// 测试数据结构
// ============================================================

type CompareUser struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
}

type CompareComplexUser struct {
	Username  string `validate:"required,min=3,max=20,alphanum"`
	Email     string `validate:"required,email"`
	Age       int    `validate:"gte=0,lte=150"`
	Phone     string `validate:"required,e164"` // 国际电话号码格式
	Address   string `validate:"required,max=200"`
	Password  string `validate:"required,min=8,max=100,containsany=!@#$%^&*"`
	BirthDate string `validate:"required,datetime=2006-01-02"`
}

// ============================================================
// gorp ValidatorService
// ============================================================

// BenchmarkValidatorCompare_Gorp_Valid gorp 验证有效数据
func BenchmarkValidatorCompare_Gorp_Valid(b *testing.B) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, _ := validate.NewValidatorService(cfg)

	user := &CompareUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      25,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Validate(ctx, user)
	}
}

// BenchmarkValidatorCompare_Gorp_Invalid gorp 验证无效数据
func BenchmarkValidatorCompare_Gorp_Invalid(b *testing.B) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, _ := validate.NewValidatorService(cfg)

	user := &CompareUser{
		Username: "ab",
		Email:    "invalid",
		Age:      200,
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Validate(ctx, user)
	}
}

// BenchmarkValidatorCompare_Gorp_Complex_Valid gorp 验证复杂有效数据
func BenchmarkValidatorCompare_Gorp_Complex_Valid(b *testing.B) {
	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, _ := validate.NewValidatorService(cfg)

	user := &CompareComplexUser{
		Username:  "testuser123",
		Email:     "test@example.com",
		Age:       25,
		Phone:     "+8613800138000",
		Address:   "123 Main Street",
		Password:  "password123!",
		BirthDate: "1990-01-15",
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Validate(ctx, user)
	}
}

// ============================================================
// 直接使用 go-playground/validator（Kratos 底层也是用这个）
// ============================================================

// BenchmarkValidatorCompare_Raw_Valid 直接使用 validator.Validate
func BenchmarkValidatorCompare_Raw_Valid(b *testing.B) {
	v := validator.New()

	user := &CompareUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Struct(user)
	}
}

// BenchmarkValidatorCompare_Raw_Invalid 直接使用 validator.Validate（无效数据）
func BenchmarkValidatorCompare_Raw_Invalid(b *testing.B) {
	v := validator.New()

	user := &CompareUser{
		Username: "ab",
		Email:    "invalid",
		Age:      200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Struct(user)
	}
}

// BenchmarkValidatorCompare_Raw_Complex_Valid 直接使用 validator.Validate（复杂有效数据）
func BenchmarkValidatorCompare_Raw_Complex_Valid(b *testing.B) {
	v := validator.New()

	user := &CompareComplexUser{
		Username:  "testuser123",
		Email:     "test@example.com",
		Age:       25,
		Phone:     "+8613800138000",
		Address:   "123 Main Street",
		Password:  "password123!",
		BirthDate: "1990-01-15",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Struct(user)
	}
}

// ============================================================
// Kratos validator middleware（模拟）
// ============================================================

// BenchmarkValidatorCompare_KratosInterface Kratos validator 接口调用开销
func BenchmarkValidatorCompare_KratosInterface(b *testing.B) {
	// Kratos 的 validator middleware 只是调用 Validate() 接口
	// 真正的验证逻辑在外部（protovalidate 或 validator）

	user := &kratosMockValidator{
		data: &CompareUser{
			Username: "testuser",
			Email:    "test@example.com",
			Age:      25,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.Validate()
	}
}

// kratosMockValidator 模拟 Kratos 的 validator 接口
type kratosMockValidator struct {
	data *CompareUser
	v    *validator.Validate
}

func (m *kratosMockValidator) Validate() error {
	if m.v == nil {
		m.v = validator.New()
	}
	return m.v.Struct(m.data)
}

// BenchmarkValidatorCompare_KratosMiddleware Kratos middleware 开销（不含实际验证）
func BenchmarkValidatorCompare_KratosMiddleware(b *testing.B) {
	// 纯 middleware 接口检查开销（不含验证）
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Kratos middleware 只做类型断言
		var req any = &CompareUser{Username: "testuser", Email: "test@example.com", Age: 25}
		if _, ok := req.(interface{ Validate() error }); ok {
			// 接口检查成功（约 10 ns）
		}
	}
}

// ============================================================
// 对比：gorp vs raw vs Kratos
// ============================================================

// BenchmarkValidatorCompare_Overhead gorp wrapper 相对于 raw validator 的开销
func BenchmarkValidatorCompare_Overhead(b *testing.B) {
	b.Run("gorp", func(b *testing.B) {
		cfg := &datacontract.ValidatorConfig{Locale: "zh"}
		svc, _ := validate.NewValidatorService(cfg)
		user := &CompareUser{Username: "testuser", Email: "test@example.com", Age: 25}
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = svc.Validate(ctx, user)
		}
	})

	b.Run("raw", func(b *testing.B) {
		v := validator.New()
		user := &CompareUser{Username: "testuser", Email: "test@example.com", Age: 25}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.Struct(user)
		}
	})

	b.Run("kratos-interface", func(b *testing.B) {
		user := &kratosMockValidator{
			data: &CompareUser{Username: "testuser", Email: "test@example.com", Age: 25},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = user.Validate()
		}
	})
}

// ============================================================
// 禁用翻译的性能对比
// ============================================================

// BenchmarkValidatorCompare_TranslateDisabled 禁用翻译后的性能
func BenchmarkValidatorCompare_TranslateDisabled(b *testing.B) {
	b.Run("invalid_with_translate", func(b *testing.B) {
		cfg := &datacontract.ValidatorConfig{Locale: "zh", TranslateErrors: true}
		svc, _ := validate.NewValidatorService(cfg)
		user := &CompareUser{Username: "ab", Email: "invalid", Age: 200}
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = svc.Validate(ctx, user)
		}
	})

	b.Run("invalid_no_translate", func(b *testing.B) {
		cfg := &datacontract.ValidatorConfig{Locale: "zh", TranslateErrors: false}
		svc, _ := validate.NewValidatorService(cfg)
		user := &CompareUser{Username: "ab", Email: "invalid", Age: 200}
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = svc.Validate(ctx, user)
		}
	})

	b.Run("valid_with_translate", func(b *testing.B) {
		cfg := &datacontract.ValidatorConfig{Locale: "zh", TranslateErrors: true}
		svc, _ := validate.NewValidatorService(cfg)
		user := &CompareUser{Username: "testuser", Email: "test@example.com", Age: 25}
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = svc.Validate(ctx, user)
		}
	})

	b.Run("valid_no_translate", func(b *testing.B) {
		cfg := &datacontract.ValidatorConfig{Locale: "zh", TranslateErrors: false}
		svc, _ := validate.NewValidatorService(cfg)
		user := &CompareUser{Username: "testuser", Email: "test@example.com", Age: 25}
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = svc.Validate(ctx, user)
		}
	})
}