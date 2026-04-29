package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSwaggerAndOpenAPIArtifactsStayInDocsDirectory(t *testing.T) {
	swaggerSource := string(mustReadFile(t, filepath.Join("cmd", "gorp", "cmd", "swagger.go")))
	openapiSource := string(mustReadFile(t, filepath.Join("cmd", "gorp", "cmd", "openapi.go")))

	if !strings.Contains(swaggerSource, `filepath.Join("docs")`) {
		t.Fatalf("swagger output dir should stay under docs/")
	}
	if !strings.Contains(openapiSource, `filepath.Join("docs", "swagger.json")`) {
		t.Fatalf("openapi input should stay docs/swagger.json")
	}
	if !strings.Contains(openapiSource, `filepath.Join("docs", "openapi.json")`) {
		t.Fatalf("openapi output should stay docs/openapi.json")
	}
}

func TestProtoGenerationArtifactsPreferProtoOrExplicitOutputDir(t *testing.T) {
	protoDir = "api/proto"
	outputDir = ""

	out := outputDir
	if out == "" {
		out = protoDir
	}
	if out != "api/proto" {
		t.Fatalf("expected proto output to default to proto dir, got %q", out)
	}

	outputDir = "gen/pb"
	out = outputDir
	if out != "gen/pb" {
		t.Fatalf("expected explicit output dir to win, got %q", out)
	}
}

func TestProtoHelpersDoNotTargetDocsOrManualDirectories(t *testing.T) {
	if err := ensureProtoDir(filepath.Join("api", "proto", "user.proto")); err != nil {
		t.Fatalf("ensureProtoDir failed: %v", err)
	}

	source := string(mustReadFile(t, filepath.Join("cmd", "gorp", "cmd", "proto.go")))
	if strings.Contains(source, "docs/manual") {
		t.Fatalf("proto command should not target docs/manual output")
	}
	if strings.Contains(source, "swagger.json") || strings.Contains(source, "openapi.json") {
		t.Fatalf("proto command should not target API doc artifacts")
	}
}
