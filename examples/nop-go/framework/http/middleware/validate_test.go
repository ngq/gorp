// Package middleware_test provides unit tests for validation middleware.
//
// 适用场景：
// - ValidateBodyMiddleware / ValidateQueryMiddleware 自动绑定与校验
// - BindAndValidateJSON / BindAndValidateQuery / BindAndValidate 便捷函数
// - 错误输出格式一致性
// - 字段名一致性
// - 国际化切换
// - 自定义校验器
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
)

// realValidator 是测试用的真实校验器，直接使用 go-playground/validator 实现 datacontract.Validator。
// 由于 middleware 包无法导入 validate provider 包（循环依赖），因此在测试中内联实现。
//
// realValidator 是测试中内联的真实校验器实现。
type realValidator struct {
	validate *validator.Validate
	trans    ut.Translator
	cfg      *datacontract.ValidatorConfig
}

// newRealValidator 创建真实校验器实例。
func newRealValidator(cfg *datacontract.ValidatorConfig) (*realValidator, error) {
	v := validator.New()

	// 注册 JSON tag 名函数，与 ValidatorService 行为一致
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	var trans ut.Translator
	switch cfg.Locale {
	case "zh":
		uni := ut.New(zh.New(), en.New())
		trans = uni.GetFallback()
		_ = zhTranslations.RegisterDefaultTranslations(v, trans)
	default:
		uni := ut.New(en.New(), zh.New())
		trans = uni.GetFallback()
		_ = enTranslations.RegisterDefaultTranslations(v, trans)
	}

	// 注册自定义规则
	for name, ruleCfg := range cfg.CustomRules {
		if ruleCfg.Fn != nil {
			_ = v.RegisterValidationCtx(name, func(ctx context.Context, fl validator.FieldLevel) bool {
				return ruleCfg.Fn(ctx, fl.Field().Interface())
			})
		}
	}

	return &realValidator{validate: v, trans: trans, cfg: cfg}, nil
}

func (rv *realValidator) Validate(ctx context.Context, obj any) error {
	err := rv.validate.StructCtx(ctx, obj)
	if err == nil {
		return nil
	}
	return rv.TranslateError(err)
}

func (rv *realValidator) ValidateVar(ctx context.Context, field any, tag string) error {
	err := rv.validate.VarCtx(ctx, field, tag)
	if err == nil {
		return nil
	}
	return rv.TranslateError(err)
}

func (rv *realValidator) RegisterCustom(name string, fn datacontract.CustomValidateFunc) error {
	return rv.validate.RegisterValidationCtx(name, func(ctx context.Context, fl validator.FieldLevel) bool {
		return fn(ctx, fl.Field().Interface())
	})
}

func (rv *realValidator) SetLocale(locale string) error {
	var trans ut.Translator
	switch locale {
	case "zh":
		uni := ut.New(zh.New(), en.New())
		trans = uni.GetFallback()
		_ = zhTranslations.RegisterDefaultTranslations(rv.validate, trans)
	case "en":
		uni := ut.New(en.New(), zh.New())
		trans = uni.GetFallback()
		_ = enTranslations.RegisterDefaultTranslations(rv.validate, trans)
	default:
		return fmt.Errorf("locale %s not supported", locale)
	}
	rv.trans = trans
	rv.cfg.Locale = locale
	return nil
}

func (rv *realValidator) TranslateError(err error) error {
	if err == nil {
		return nil
	}
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
	}

	if !rv.cfg.TranslateErrors {
		msgs := make([]string, len(validationErrors))
		for i, fe := range validationErrors {
			msgs[i] = fe.Error()
		}
		return resiliencecontract.BadRequest(
			resiliencecontract.ErrorReasonBadRequest,
			strings.Join(msgs, "; "),
		)
	}

	details := make([]datacontract.ValidationError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		details = append(details, datacontract.ValidationError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: fe.Translate(rv.trans),
			Value:   fe.Value(),
		})
	}

	msgs := make([]string, len(details))
	for i, d := range details {
		msgs[i] = d.Message
	}

	detailsJSON, _ := json.Marshal(details)

	return resiliencecontract.BadRequest(
		resiliencecontract.ErrorReasonBadRequest,
		strings.Join(msgs, "; "),
	).WithMetadata(map[string]string{
		"validation_errors": string(detailsJSON),
		"error_count":       fmt.Sprintf("%d", len(details)),
	})
}

// ===================== ValidateBodyMiddleware 测试 =====================

// TestValidateBodyMiddleware_Success 测试 ValidateBodyMiddleware 校验成功时放行请求。
//
// 中文说明：
// - 请求体合法时，中间件应放行并存储已校验对象到上下文。
func TestValidateBodyMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type CreateUserReq struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
	}

	router := NewTestEngine()
	router.POST("/users", ValidateBodyMiddleware(svc, &CreateUserReq{}), func(c *gin.Context) {
		// 从上下文获取已校验对象
		validatedBody, ok := supportcontract.FromValidatedBodyContext(c.Request.Context())
		if !ok {
			c.String(http.StatusInternalServerError, "missing validated body")
			return
		}
		body := validatedBody.(*CreateUserReq)
		c.String(http.StatusOK, "name=%s,email=%s", body.Name, body.Email)
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"alice","email":"alice@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "name=alice") {
		t.Errorf("expected response to contain 'name=alice', got: %s", recorder.Body.String())
	}
}

// TestValidateBodyMiddleware_ValidationFail 测试 ValidateBodyMiddleware 校验失败时返回统一错误。
//
// 中文说明：
// - 请求体不满足校验规则时，应返回 400 + ValidateErrorResponse 格式。
func TestValidateBodyMiddleware_ValidationFail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type CreateUserReq struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
	}

	router := NewTestEngine()
	router.POST("/users", ValidateBodyMiddleware(svc, &CreateUserReq{}), func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach here")
	})

	// 名字为空（违反 required），邮箱格式错误（违反 email）
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var resp ValidateErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != 400 {
		t.Errorf("expected code 400, got: %d", resp.Code)
	}
	if resp.Reason != "BAD_REQUEST" {
		t.Errorf("expected reason BAD_REQUEST, got: %s", resp.Reason)
	}
	// 有 details（因为 TranslateErrors=true）
	if resp.Details == "" {
		t.Error("expected non-empty details in validation error response")
	}
}

// TestValidateBodyMiddleware_InvalidJSON 测试 ValidateBodyMiddleware 处理无效 JSON。
//
// 中文说明：
// - 请求体不是合法 JSON 时，应返回 400 + 绑定错误信息。
func TestValidateBodyMiddleware_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type CreateUserReq struct {
		Name string `json:"name" validate:"required"`
	}

	router := NewTestEngine()
	router.POST("/users", ValidateBodyMiddleware(svc, &CreateUserReq{}), func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach here")
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", recorder.Code)
	}
}

// TestValidateBodyMiddleware_PanicsOnNonStructPointer 测试 ValidateBodyMiddleware 传入非结构体指针时 panic。
//
// 中文说明：
// - prototype 必须是结构体指针，否则应在注册时 panic，而不是在请求时才发现。
func TestValidateBodyMiddleware_PanicsOnNonStructPointer(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for non-struct pointer prototype")
		}
	}()

	cfg := &datacontract.ValidatorConfig{Locale: "zh"}
	svc, _ := newRealValidator(cfg)

	// 传入非结构体指针（string 指针）
	ValidateBodyMiddleware(svc, new(string))
}

// ===================== ValidateQueryMiddleware 测试 =====================

// TestValidateQueryMiddleware_Success 测试 ValidateQueryMiddleware 校验成功时放行请求。
//
// 中文说明：
// - 查询参数合法时，中间件应放行并存储已校验对象到上下文。
func TestValidateQueryMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type ListReq struct {
		Page int `form:"page" validate:"required,gte=1"`
		Size int `form:"size" validate:"required,gte=1,lte=100"`
	}

	router := NewTestEngine()
	router.GET("/items", ValidateQueryMiddleware(svc, &ListReq{}), func(c *gin.Context) {
		validatedBody, ok := supportcontract.FromValidatedBodyContext(c.Request.Context())
		if !ok {
			c.String(http.StatusInternalServerError, "missing validated body")
			return
		}
		body := validatedBody.(*ListReq)
		c.String(http.StatusOK, "page=%d,size=%d", body.Page, body.Size)
	})

	req := httptest.NewRequest(http.MethodGet, "/items?page=1&size=10", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", recorder.Code, recorder.Body.String())
	}
}

// TestValidateQueryMiddleware_ValidationFail 测试 ValidateQueryMiddleware 校验失败时返回统一错误。
//
// 中文说明：
// - 查询参数不满足校验规则时，应返回 400 + ValidateErrorResponse 格式。
func TestValidateQueryMiddleware_ValidationFail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type ListReq struct {
		Page int `form:"page" validate:"required,gte=1"`
		Size int `form:"size" validate:"required,gte=1,lte=100"`
	}

	router := NewTestEngine()
	router.GET("/items", ValidateQueryMiddleware(svc, &ListReq{}), func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach here")
	})

	// page=0 违反 gte=1
	req := httptest.NewRequest(http.MethodGet, "/items?page=0&size=10", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var resp ValidateErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != 400 {
		t.Errorf("expected code 400, got: %d", resp.Code)
	}
}

// ===================== BindAndValidateQuery 测试 =====================

// TestBindAndValidateQueryStoresValidatedBody 测试 BindAndValidateQuery 成功时存储已校验对象。
//
// 中文说明：
// - 查询参数绑定并校验成功后，应将对象存入请求上下文。
func TestBindAndValidateQueryStoresValidatedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validator := &stubValidator{
		validateFn: func(ctx context.Context, obj any) error { return nil },
	}

	router := NewTestEngine()
	router.GET("/search", func(c *gin.Context) {
		httpCtx := newContext(c)
		input := &struct {
			Q string `form:"q"`
		}{}
		if err := BindAndValidateQuery(httpCtx, validator, input); err != nil {
			c.String(http.StatusBadRequest, "error")
			return
		}
		validatedBody, ok := supportcontract.FromValidatedBodyContext(c.Request.Context())
		if !ok {
			c.String(http.StatusInternalServerError, "missing validated body")
			return
		}
		body := validatedBody.(*struct {
			Q string `form:"q"`
		})
		c.String(http.StatusOK, "q=%s", body.Q)
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "q=hello" {
		t.Errorf("expected q=hello, got: %s", recorder.Body.String())
	}
}

// ===================== BindAndValidate (form) 测试 =====================

// TestBindAndValidateReturnsUnifiedError 测试 BindAndValidate 校验失败返回统一错误。
//
// 中文说明：
// - 表单绑定后校验失败，应返回 400 + 统一错误格式。
func TestBindAndValidateReturnsUnifiedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validator := &stubValidator{
		validateFn: func(context.Context, any) error {
			return resiliencecontract.BadRequest(
				resiliencecontract.ErrorReasonBadRequest,
				"form validation failed",
			)
		},
	}

	router := NewTestEngine()
	router.POST("/form", func(c *gin.Context) {
		httpCtx := newContext(c)
		input := &struct {
			Name string `form:"name"`
		}{}
		_ = BindAndValidate(httpCtx, validator, input)
	})

	req := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader("name="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var resp ValidateErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "form validation failed" {
		t.Errorf("expected 'form validation failed', got: %s", resp.Message)
	}
}

// ===================== JSON 字段名一致性测试 =====================

// TestValidateBodyMiddleware_JSONFieldNameConsistency 测试中间件输出中使用 JSON 字段名。
//
// 中文说明：
// - 校验错误响应的 details 中应包含 JSON tag 名（如 username）而非 Go 字段名（如 Username）。
func TestValidateBodyMiddleware_JSONFieldNameConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type UserReq struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
	}

	router := NewTestEngine()
	router.POST("/users", ValidateBodyMiddleware(svc, &UserReq{}), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// 两个字段都校验失败
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"username":"","email":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var resp ValidateErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// details 中应包含 JSON 字段名
	if !strings.Contains(resp.Details, "username") {
		t.Errorf("expected details to contain JSON field name 'username', got: %s", resp.Details)
	}
	if !strings.Contains(resp.Details, "email") {
		t.Errorf("expected details to contain JSON field name 'email', got: %s", resp.Details)
	}
}

// ===================== 国际化测试 =====================

// TestValidateBodyMiddleware_LocaleZh 测试中文 locale 下的错误消息。
//
// 中文说明：
// - 使用中文 locale 配置，校验错误消息应包含中文翻译。
func TestValidateBodyMiddleware_LocaleZh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "zh",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type Req struct {
		Name string `json:"name" validate:"required"`
	}

	router := NewTestEngine()
	router.POST("/test", ValidateBodyMiddleware(svc, &Req{}), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var resp ValidateErrorResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)

	// 中文翻译应包含"必填"关键字
	if !strings.Contains(resp.Message, "必填") {
		t.Errorf("expected Chinese error message containing '必填', got: %s", resp.Message)
	}
}

// TestValidateBodyMiddleware_LocaleEn 测试英文 locale 下的错误消息。
//
// 中文说明：
// - 使用英文 locale 配置，校验错误消息应包含英文翻译。
func TestValidateBodyMiddleware_LocaleEn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &datacontract.ValidatorConfig{
		Locale:          "en",
		TranslateErrors: true,
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type Req struct {
		Name string `json:"name" validate:"required"`
	}

	router := NewTestEngine()
	router.POST("/test", ValidateBodyMiddleware(svc, &Req{}), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var resp ValidateErrorResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)

	// 英文翻译应包含 "required" 关键字
	if !strings.Contains(strings.ToLower(resp.Message), "required") {
		t.Errorf("expected English error message containing 'required', got: %s", resp.Message)
	}
}

// ===================== 自定义校验器测试 =====================

// TestValidateBodyMiddleware_CustomValidator 测试自定义校验规则在中间件中生效。
//
// 中文说明：
// - 注册自定义校验规则（如 mobile）后，ValidateBodyMiddleware 应能正确触发。
func TestValidateBodyMiddleware_CustomValidator(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
					return len(s) == 11 && s[0] == '1'
				},
			},
		},
	}
	svc, err := newRealValidator(cfg)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	type PhoneReq struct {
		Phone string `json:"phone" validate:"required,mobile"`
	}

	router := NewTestEngine()
	router.POST("/sms", ValidateBodyMiddleware(svc, &PhoneReq{}), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// 有效手机号
	req := httptest.NewRequest(http.MethodPost, "/sms", strings.NewReader(`{"phone":"13800138000"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Errorf("expected 200 for valid mobile, got %d", recorder.Code)
	}

	// 无效手机号
	req = httptest.NewRequest(http.MethodPost, "/sms", strings.NewReader(`{"phone":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid mobile, got %d", recorder.Code)
	}
}

// ===================== nil validator 测试 =====================

// TestBindAndValidateJSON_NilValidator 测试 validator 为 nil 时不执行校验。
//
// 中文说明：
// - 当 validator 参数为 nil 时，BindAndValidateJSON 只做绑定，不执行校验。
func TestBindAndValidateJSON_NilValidator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewTestEngine()
	router.POST("/test", func(c *gin.Context) {
		httpCtx := newContext(c)
		input := &struct {
			Name string `json:"name"`
		}{}
		if err := BindAndValidateJSON(httpCtx, nil, input); err != nil {
			c.String(http.StatusBadRequest, "error")
			return
		}
		c.String(http.StatusOK, "name=%s", input.Name)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
