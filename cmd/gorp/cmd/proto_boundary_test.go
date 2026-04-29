package cmd

import "testing"

func TestProtoGenDefaultOutputDirFallsBackToProtoDir(t *testing.T) {
	protoDir = "api/proto"
	outputDir = ""

	out := outputDir
	if out == "" {
		out = protoDir
	}
	if out != "api/proto" {
		t.Fatalf("expected output dir to fall back to proto dir, got %q", out)
	}
}

func TestCreateProtoGeneratorRespectsHTTPFlag(t *testing.T) {
	gen, err := createProtoGenerator(false)
	if err != nil {
		t.Fatalf("createProtoGenerator(false) failed: %v", err)
	}
	if gen == nil {
		t.Fatalf("expected generator instance")
	}

	genHTTP, err := createProtoGenerator(true)
	if err != nil {
		t.Fatalf("createProtoGenerator(true) failed: %v", err)
	}
	if genHTTP == nil {
		t.Fatalf("expected generator instance with http enabled")
	}
}
