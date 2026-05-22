// Package gin_test provides unit tests for Gin-backed context output helpers.
//
// 适用场景：
// - 验证基于 Gin 的 context 输出助手行为一致。
// - 防止字符串、XML、Data 和 Redirect 输出在 provider 适配层回归。
// - 在主线抽象演进时，保持 provider 级上下文输出行为稳定。
package gin

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	ginpkg "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type xmlPayload struct {
	XMLName xml.Name `xml:"reply"`
	Message string   `xml:"message"`
}

// TestContextString verifies string output through the Gin-backed context helper.
//
// TestContextString 验证基于 Gin 的 context 字符串输出助手。
func TestContextString(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newContext(ctx)
	httpCtx.String(http.StatusOK, "success")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "success", rec.Body.String())
}

// TestContextXML verifies XML output through the Gin-backed context helper.
//
// TestContextXML 验证基于 Gin 的 context XML 输出助手。
func TestContextXML(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newContext(ctx)
	httpCtx.XML(http.StatusOK, xmlPayload{Message: "ok"})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "<reply>")
	require.Contains(t, rec.Body.String(), "<message>ok</message>")
}

// TestContextData verifies binary data output through the Gin-backed context helper.
//
// TestContextData 验证基于 Gin 的 context Data 输出助手。
func TestContextData(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newContext(ctx)
	httpCtx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("pong"))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "pong", rec.Body.String())
	require.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
}

// TestContextRedirect verifies redirect output through the Gin-backed context helper.
//
// TestContextRedirect 验证基于 Gin 的 context Redirect 输出助手。
func TestContextRedirect(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newContext(ctx)
	httpCtx.Redirect(http.StatusFound, "/callback")

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/callback", rec.Header().Get("Location"))
}
