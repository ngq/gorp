package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutesTemplateUsesRootFacadeTypes(t *testing.T) {
	require.Contains(t, routesTpl, `gorp "github.com/ngq/gorp"`)
	require.Contains(t, routesTpl, "RegisterRoutes(r gorp.HTTPRouter)")
	require.Contains(t, routesTpl, "container, ok := gorp.FromContainerContext(ctx)")
	require.Contains(t, routesTpl, "container.Make(gorp.DBRuntimeKey)")
	require.NotContains(t, routesTpl, "framework/contract")
	require.NotContains(t, routesTpl, "contract.HTTPRouter")
	require.NotContains(t, routesTpl, "contract.Container")
	require.NotContains(t, routesTpl, "contract.DBRuntimeKey")
}

func TestEntRoutesTemplateUsesRootFacadeTypes(t *testing.T) {
	require.Contains(t, entRoutesTpl, `gorp "github.com/ngq/gorp"`)
	require.Contains(t, entRoutesTpl, "RegisterRoutes(r gorp.HTTPRouter)")
	require.Contains(t, entRoutesTpl, "container, ok := gorp.FromContainerContext(ctx)")
	require.Contains(t, entRoutesTpl, "container.Make(gorp.DBRuntimeKey)")
	require.NotContains(t, entRoutesTpl, "framework/contract")
	require.NotContains(t, entRoutesTpl, "contract.HTTPRouter")
	require.NotContains(t, entRoutesTpl, "contract.Container")
	require.NotContains(t, entRoutesTpl, "contract.DBRuntimeKey")
}

func TestCRUDTemplatesUseRootFacadeResponseHelpers(t *testing.T) {
	templates := []string{
		createTpl,
		updateTpl,
		getTpl,
		listTpl,
		deleteTpl,
		entCreateTpl,
		entUpdateTpl,
		entGetTpl,
		entListTpl,
		entDeleteTpl,
	}

	for _, tpl := range templates {
		require.Contains(t, tpl, `gorp "github.com/ngq/gorp"`)
		require.NotContains(t, tpl, "framework/provider/gin")
		require.Contains(t, tpl, "gorp.InternalError(")
		require.Contains(t, tpl, "gorp.Error(")
		require.True(
			t,
			containsAny(tpl, "gorp.BadRequest(", "gorp.Success(", "gorp.SuccessWithMessage(", "gorp.SuccessWithStatus("),
			"expected template to use root facade response helpers: %s",
			tpl,
		)
		require.True(
			t,
			containsAny(tpl, "api.mustRuntimeGorm(c.Context())", "api.mustEntClient(c.Context())"),
			"expected template to resolve runtime backend from request context: %s",
			tpl,
		)
	}
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func TestAppendGeneratedModuleRouteSupportsCurrentGoLayoutRoutes(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n"), 0o644))

	routesFile := filepath.Join(root, "routes.go")
	content := `package http

import (
	"example.com/demo/app/http/handler"
	"example.com/demo/internal/service"

	gorp "github.com/ngq/gorp"
)

func RegisterRoutes(r gorp.HTTPRouter, svc *service.Services) {
	demoHandler := handler.NewDemoHandler(svc.Demo)

	api := r.Group("/api/v1")
	{
		demos := api.Group("/demos")
		{
			demos.GET("", demoHandler.List)
		}
	}
}
`
	require.NoError(t, os.WriteFile(routesFile, []byte(content), 0o644))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	defer func() { _ = os.Chdir(wd) }()

	updated, err := appendGeneratedModuleRoute(routesFile, "user")
	require.NoError(t, err)
	require.True(t, updated)

	got, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(got)
	require.Contains(t, text, `generateduser "example.com/demo/app/http/module/generated/user"`)
	require.Contains(t, text, `generateduser.RegisterRoutes(r)`)
}

func TestAppendGeneratedModuleRouteSupportsRouterVariableNamedRouter(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n"), 0o644))

	routesFile := filepath.Join(root, "routes.go")
	content := `package http

import (
	gorp "github.com/ngq/gorp"
)

func RegisterRoutes(router gorp.HTTPRouter) {
	router.GET("/api/ping", nil)
}
`
	require.NoError(t, os.WriteFile(routesFile, []byte(content), 0o644))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	defer func() { _ = os.Chdir(wd) }()

	updated, err := appendGeneratedModuleRoute(routesFile, "user")
	require.NoError(t, err)
	require.True(t, updated)

	got, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(got)
	require.Contains(t, text, `generateduser "example.com/demo/app/http/module/generated/user"`)
	require.Contains(t, text, `generateduser.RegisterRoutes(router)`)
}
