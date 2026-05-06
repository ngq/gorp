// Application scenarios:
// - Verify that the Gin-backed HTTP context output helpers behave consistently.
// - Guard the provider adapter against regressions in string, XML, data, and redirect output.
// - Keep provider-level context output behavior stable while mainline abstractions evolve.
//
// 适用场景：
// - 验证基于 Gin 的 HTTP context 输出助手行为一致。
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

// TestHTTPContextString verifies string output through the Gin-backed HTTP context helper.
//
// TestHTTPContextString 验证基于 Gin 的 HTTP context 字符串输出助手。
func TestHTTPContextString(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newHTTPContext(ctx)
	httpCtx.String(http.StatusOK, "success")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "success", rec.Body.String())
}

// TestHTTPContextXML verifies XML output through the Gin-backed HTTP context helper.
//
// TestHTTPContextXML 验证基于 Gin 的 HTTP context XML 输出助手。
func TestHTTPContextXML(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newHTTPContext(ctx)
	httpCtx.XML(http.StatusOK, xmlPayload{Message: "ok"})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "<reply>")
	require.Contains(t, rec.Body.String(), "<message>ok</message>")
}

// TestHTTPContextData verifies binary data output through the Gin-backed HTTP context helper.
//
// TestHTTPContextData 验证基于 Gin 的 HTTP context Data 输出助手。
func TestHTTPContextData(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newHTTPContext(ctx)
	httpCtx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("pong"))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "pong", rec.Body.String())
	require.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
}

// TestHTTPContextRedirect verifies redirect output through the Gin-backed HTTP context helper.
//
// TestHTTPContextRedirect 验证基于 Gin 的 HTTP context Redirect 输出助手。
func TestHTTPContextRedirect(t *testing.T) {
	ginpkg.SetMode(ginpkg.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	httpCtx := newHTTPContext(ctx)
	httpCtx.Redirect(http.StatusFound, "/callback")

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/callback", rec.Header().Get("Location"))
}
