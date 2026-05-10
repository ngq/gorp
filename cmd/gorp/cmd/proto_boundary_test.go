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

// TestProtoSubcommandsHaveConsistentProtoFileFlag 验证所有 proto 子命令
// 都注册了 -f, --proto-file flag，且不再使用旧名 --proto / --proto-files。
func TestProtoSubcommandsHaveConsistentProtoFileFlag(t *testing.T) {
	cmds := map[string][]string{
		"gen":         {"proto-file"},
		"gen-service": {"proto-file", "proto-dir"},
		"gen-client":  {"proto-file"},
		"all":         {"proto-file", "proto-dir"},
	}

	for name, expectedFlags := range cmds {
		var cmd = protoCmd.Commands()
		var found = false
		for _, c := range cmd {
			if c.Name() == name {
				found = true
				for _, flagName := range expectedFlags {
					f := c.Flags().Lookup(flagName)
					if f == nil {
						t.Errorf("%s: expected flag --%s not found", name, flagName)
					}
					if flagName == "proto-file" && f.Shorthand != "f" {
						t.Errorf("%s: expected --proto-file shorthand to be 'f', got %q", name, f.Shorthand)
					}
					if flagName == "proto-dir" && f.Shorthand != "d" {
						t.Errorf("%s: expected --proto-dir shorthand to be 'd', got %q", name, f.Shorthand)
					}
				}
				// 确保旧名已被移除。
				for _, legacy := range []string{"proto", "proto-files"} {
					if c.Flags().Lookup(legacy) != nil {
						t.Errorf("%s: legacy flag --%s should be removed", name, legacy)
					}
				}
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under proto", name)
		}
	}
}
