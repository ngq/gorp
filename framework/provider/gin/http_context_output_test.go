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
