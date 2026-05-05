package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gingonic "github.com/gin-gonic/gin"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
)

type validateStub struct{}

func (validateStub) Validate(context.Context, any) error                          { return nil }
func (validateStub) ValidateVar(context.Context, any, string) error               { return nil }
func (validateStub) RegisterCustom(string, datacontract.CustomValidateFunc) error { return nil }
func (validateStub) SetLocale(string) error                                       { return nil }
func (validateStub) TranslateError(err error) resiliencecontract.AppError {
	return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
}

type loginRequest struct {
	UserName string `json:"username" validate:"required"`
}

func TestBindAndValidateJSONInHandler(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	r.POST("/login", func(c *gingonic.Context) {
		httpCtx := newHTTPContext(c)
		var req loginRequest
		if err := BindAndValidateJSON(httpCtx, validateStub{}, &req); err != nil {
			return
		}
		fromCtx, ok := supportcontract.FromValidatedBodyContext(httpCtx.Request().Context())
		if !ok {
			t.Fatal("expected validated body in request context")
		}
		storedReq, ok := fromCtx.(*loginRequest)
		if !ok {
			t.Fatalf("expected *loginRequest, got %T", fromCtx)
		}
		if storedReq.UserName != "alice" {
			t.Fatalf("expected username alice, got %s", storedReq.UserName)
		}
		c.JSON(http.StatusOK, map[string]any{"username": req.UserName})
	})

	body, _ := json.Marshal(map[string]any{"username": "alice"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBindAndValidateJSONStoresValidatedBodyInRequestContext(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	body, _ := json.Marshal(map[string]any{"username": "alice"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx, _ := gingonic.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	httpCtx := newHTTPContext(ctx)
	var payload loginRequest
	if err := BindAndValidateJSON(httpCtx, validateStub{}, &payload); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if payload.UserName != "alice" {
		t.Fatalf("expected username alice, got %s", payload.UserName)
	}
	fromReq, ok := supportcontract.FromValidatedBodyContext(httpCtx.Request().Context())
	if !ok {
		t.Fatal("expected validated body in HTTPContext request context")
	}
	storedReq, ok := fromReq.(*loginRequest)
	if !ok {
		t.Fatalf("expected *loginRequest, got %T", fromReq)
	}
	if storedReq.UserName != "alice" {
		t.Fatalf("expected username alice, got %s", storedReq.UserName)
	}
}

func TestBindAndValidateQuery(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	type query struct {
		Page int `form:"page" validate:"gte=1"`
	}
	r.GET("/items", func(c *gingonic.Context) {
		var req query
		if err := BindAndValidateQuery(newHTTPContext(c), validateStub{}, &req); err != nil {
			return
		}
		c.JSON(http.StatusOK, map[string]any{"page": req.Page})
	})

	req := httptest.NewRequest(http.MethodGet, "/items?page=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBindAndValidateForm(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	type form struct {
		Name string `form:"name" validate:"required"`
	}
	r.POST("/form", func(c *gingonic.Context) {
		var req form
		if err := BindAndValidate(newHTTPContext(c), validateStub{}, &req); err != nil {
			return
		}
		c.JSON(http.StatusOK, map[string]any{"name": req.Name})
	})

	req := httptest.NewRequest(http.MethodPost, "/form?name=bob", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
