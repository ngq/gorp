package facade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFacadeDependencyDirectionStaysKernelFacing(t *testing.T) {
	path := filepath.Join("facade.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read facade.go: %v", err)
	}
	text := string(content)

	for _, blocked := range []string{
		"cmd/",
		"/cmd/",
		"template",
		"starter",
		"framework/deploy",
	} {
		if strings.Contains(text, blocked) {
			t.Fatalf("facade should not depend on cli/template/starter semantics, blocked token: %s", blocked)
		}
	}

	for _, requiredImport := range []string{
		`"github.com/ngq/gorp/framework/bootstrap"`,
		`"github.com/ngq/gorp/framework/contract"`,
	} {
		if !strings.Contains(text, requiredImport) {
			t.Fatalf("facade should keep kernel-facing import: %s", requiredImport)
		}
	}
}
