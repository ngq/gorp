package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gingonic "github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestRouterRoutesIncludesChildGroupsAndAllMethods(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)
	engine := gingonic.New()
	root := newRouter(&engine.RouterGroup, engine)
	api := root.Group("/api").Group("/v1")
	handler := func(c transportcontract.Context) { c.Status(http.StatusNoContent) }

	api.GET("/users/:id", handler)
	api.PATCH("/users/:id", handler)
	api.HEAD("/users/:id", handler)
	api.OPTIONS("/users/:id", handler)

	want := []transportcontract.RouteInfo{
		{Method: http.MethodGet, Path: "/api/v1/users/:id"},
		{Method: http.MethodPatch, Path: "/api/v1/users/:id"},
		{Method: http.MethodHead, Path: "/api/v1/users/:id"},
		{Method: http.MethodOptions, Path: "/api/v1/users/:id"},
	}
	require.Equal(t, want, root.Routes())
	require.Equal(t, want, api.Routes())

	snapshot := root.Routes()
	snapshot[0].Path = "/mutated"
	require.Equal(t, "/api/v1/users/:id", root.Routes()[0].Path)
}

func TestRouterMountRecordsGetAndHead(t *testing.T) {
	engine := gingonic.New()
	router := newRouter(&engine.RouterGroup, engine)
	router.Mount("/status", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	require.Equal(t, []transportcontract.RouteInfo{
		{Method: http.MethodGet, Path: "/status"},
		{Method: http.MethodHead, Path: "/status"},
	}, router.Routes())
}

func TestRouterNoRouteAndNoMethod(t *testing.T) {
	engine := gingonic.New()
	router := newRouter(&engine.RouterGroup, engine)
	router.GET("/known", func(c transportcontract.Context) { c.Status(http.StatusNoContent) })
	router.NoRoute(func(c transportcontract.Context) {
		c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	})
	router.NoMethod(func(c transportcontract.Context) {
		c.JSON(http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	})

	notFound := httptest.NewRecorder()
	engine.ServeHTTP(notFound, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, http.StatusNotFound, notFound.Code)

	notAllowed := httptest.NewRecorder()
	engine.ServeHTTP(notAllowed, httptest.NewRequest(http.MethodPost, "/known", nil))
	require.Equal(t, http.StatusMethodNotAllowed, notAllowed.Code)
}
