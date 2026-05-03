package gorp

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRootExportsStayFacadeThin(t *testing.T) {
	path := filepath.Join("gorp.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read gorp.go: %v", err)
	}
	text := string(content)

	for _, blocked := range []string{"BootHTTPService", "RunHTTP", "NewHTTPServiceRuntime", "MustMake"} {
		if strings.Contains(text, blocked) {
			t.Fatalf("root export should stay facade-thin, blocked token found: %s", blocked)
		}
	}
	for _, blockedSemantic := range []string{"starter", "template", "framework/deploy", "cmd/"} {
		if strings.Contains(text, blockedSemantic) {
			t.Fatalf("root export should not include non-facade semantic token: %s", blockedSemantic)
		}
	}
	importBlock := text
	if m := regexp.MustCompile(`(?s)import\\s*\\(.*?\\)`).FindString(text); m != "" {
		importBlock = m
	}
	for _, blockedImport := range []string{"framework/bootstrap", "framework/container"} {
		if strings.Contains(importBlock, blockedImport) {
			t.Fatalf("root export should stay facade-thin, blocked import found: %s", blockedImport)
		}
	}
	for _, requiredImport := range []string{`"github.com/ngq/gorp/framework/facade"`, `"github.com/ngq/gorp/framework/contract"`} {
		if !strings.Contains(importBlock, requiredImport) {
			t.Fatalf("root export should keep thin forward imports, missing: %s", requiredImport)
		}
	}

	for _, required := range []string{"func Run(", "func Start(", "func RunContext(", "func BuildHTTPRuntime(", "func Build(", "func HTTP(", "func WithoutHTTP(", "func Module(", "func Modules(", "func WithModule(", "func WithProviders(", "func WithMigrate(", "func WithSetup(", "func WithHTTPRoutes("} {
		if !strings.Contains(text, required) {
			t.Fatalf("exports missing required facade entry: %s", required)
		}
	}
	for _, requiredTypeAlias := range []string{"type MigrateFunc =", "type SetupFunc =", "type HTTPRouteRegistrar ="} {
		if !strings.Contains(text, requiredTypeAlias) {
			t.Fatalf("exports missing required facade type alias: %s", requiredTypeAlias)
		}
	}
	for _, requiredTypeAlias := range []string{"type HTTPRuntime =", "type HTTPServiceOptions =", "type ServiceProvider =", "type Option ="} {
		if !strings.Contains(text, requiredTypeAlias) {
			t.Fatalf("exports missing required facade/common alias: %s", requiredTypeAlias)
		}
	}
	for _, requiredVar := range []string{"ErrServiceNameRequired", "ErrNoServiceDeclared", "ErrHTTPRouteRegistrationFailed", "ErrHTTPRuntimeUnavailable", "ErrSetupFailed", "ErrMigrateFailed", "ErrStartupCanceled", "ErrHTTPServiceRunFailed", "ErrHTTPRuntimeBuildFailed"} {
		if !strings.Contains(text, requiredVar) {
			t.Fatalf("exports missing required facade error: %s", requiredVar)
		}
	}
}
