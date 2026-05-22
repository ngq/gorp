// Package proto_test provides unit and integration tests for the proto generator.
//
// 适用场景：
// - 本文件包含测试辅助函数，供其他测试文件使用。
package proto

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func runGenFromService(t *testing.T, opts integrationcontract.ServiceToProtoOptions) string {
	t.Helper()
	content, err := genFromService(t, opts)
	if err != nil {
		t.Fatalf("GenFromService failed: %v", err)
	}
	return content
}

func genFromService(t *testing.T, opts integrationcontract.ServiceToProtoOptions) (string, error) {
	t.Helper()
	cfg := &integrationcontract.ProtoGeneratorConfig{DefaultProtoDir: filepath.Dir(opts.OutputPath)}
	gen, err := NewGenerator(cfg)
	if err != nil {
		return "", err
	}
	if err := gen.GenFromService(context.Background(), opts); err != nil {
		return "", err
	}
	content, err := os.ReadFile(opts.OutputPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func assertContainsAll(t *testing.T, content string, want ...string) {
	t.Helper()
	for _, item := range want {
		if !strings.Contains(content, item) {
			t.Fatalf("generated content missing %q\n%s", item, content)
		}
	}
}

func assertNotContainsAll(t *testing.T, content string, unwanted ...string) {
	t.Helper()
	for _, item := range unwanted {
		if strings.Contains(content, item) {
			t.Fatalf("generated content should not contain %q\n%s", item, content)
		}
	}
}
