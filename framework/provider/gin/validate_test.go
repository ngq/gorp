package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gingonic "github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

type validateStub struct{}

func (validateStub) Validate(context.Context, any) error { return nil }
func (validateStub) ValidateVar(context.Context, any, string) error { return nil }
func (validateStub) RegisterCustom(string, contract.CustomValidateFunc) error { return nil }
func (validateStub) SetLocale(string) error { return nil }
func (validateStub) TranslateError(err error) contract.AppError {
	return contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
}

type loginRequest struct {
	UserName string `json:"username" validate:"required"`
}

func TestValidateBody(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	r.POST("/login", func(c *gingonic.Context) {
		var req loginRequest
		if err := ValidateBody(c, validateStub{}, &req); err != nil {
			return
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

func TestValidateQuery(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	type query struct {
		Page int `form:"page" validate:"gte=1"`
	}
	r.GET("/items", func(c *gingonic.Context) {
		var req query
		if err := ValidateQuery(c, validateStub{}, &req); err != nil {
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

func TestValidateForm(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	r := gingonic.New()
	type form struct {
		Name string `form:"name" validate:"required"`
	}
	r.POST("/form", func(c *gingonic.Context) {
		var req form
		if err := ValidateForm(c, validateStub{}, &req); err != nil {
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
