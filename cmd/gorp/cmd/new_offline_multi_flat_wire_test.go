package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestRenderMultiFlatWireTemplateIncludesTests(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwverify")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwverify",
		Module:          "example.com/mfwverify",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	err := renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data)
	require.NoError(t, err)

	mustExist := []string{
		"services/user/internal/biz/user_test.go",
		"services/order/internal/biz/order_test.go",
		"services/product/internal/biz/product_test.go",
		"services/user/internal/server/http_test.go",
		"services/order/internal/server/http_test.go",
		"services/product/internal/server/http_test.go",
		"services/user/internal/service/grpc_smoke_test.go",
		"services/order/internal/service/grpc_smoke_test.go",
		"deploy/compose/docker-compose.yaml",
		"deploy/docker/Dockerfile.user",
		"deploy/docker/Dockerfile.order",
		"deploy/docker/Dockerfile.product",
		"docs/structure.md",
		"docs/component-tree.md",
		"docs/next-steps.md",
		"Makefile",
	}

	for _, rel := range mustExist {
		path := filepath.Join(projectDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", rel, err)
		}
	}

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "多服务起步模板（Wire 装配版）")
	require.Contains(t, readmeText, "> 起步：")
	require.Contains(t, readmeText, "## 先看这三件事")
	require.Contains(t, readmeText, "- 创建项目：`gorp new multi-wire`")
	require.Contains(t, readmeText, "- 生成代码：`make generate`")
	require.NotContains(t, readmeText, "安装代码生成工具")
	require.Contains(t, readmeText, "### 2. 统一生成代码")
	require.Contains(t, readmeText, "## 继续往下时再看")
	require.Contains(t, readmeText, "## 目录速览")
	require.Contains(t, readmeText, "目录：")
	require.Contains(t, readmeText, "## 日志建议")
	require.NotContains(t, readmeText, "## 技术栈")
	require.NotContains(t, readmeText, "母仓/单服务场景")
	require.Equal(t, 1, strings.Count(readmeText, "### 5. 本地联调"))
	require.Contains(t, readmeText, "### 4. 运行测试")
	require.Contains(t, readmeText, "make test")
	require.Contains(t, readmeText, "### 5. 本地联调")
	require.Contains(t, readmeText, "make deploy-local")
	require.Contains(t, readmeText, "附带：")
	require.Contains(t, readmeText, "看：")
	require.Contains(t, readmeText, "验：")
	require.Contains(t, readmeText, "### 8. 验证 Proto-first gRPC 双服务链路")
	require.Contains(t, readmeText, "建议：")
	require.Contains(t, readmeText, "配：")
	require.Contains(t, readmeText, "主线：")
	require.Contains(t, readmeText, "### 9. 生成 proto 代码")
	require.Contains(t, readmeText, "修改 `proto/*.proto` 后")
	require.Contains(t, readmeText, "生成：")
	require.Contains(t, readmeText, "默认优先按服务间 Proto-first 接线验证；")
	require.Contains(t, readmeText, "这一步按需执行。")
	require.Contains(t, readmeText, "有请求上下文时优先使用 `log.Ctx(ctx)`；")
	require.Less(t, strings.Index(readmeText, "### 4. 运行测试"), strings.Index(readmeText, "### 5. 本地联调"))
	require.Less(t, strings.Index(readmeText, "### 5. 本地联调"), strings.Index(readmeText, "### 6. Harbor 推送镜像"))
	require.Contains(t, readmeText, "- 统一执行 Wire 与 Proto 生成；")
	require.Contains(t, readmeText, "- 服务入口由 `services/<service>/cmd/main.go` 承接。")
	require.Contains(t, readmeText, "- 各服务统一由 `services/<service>/cmd/main.go` 启动；")
	require.Contains(t, readmeText, "- 传输层注册在 `internal/server`，业务编排在 `internal/service`。")
	require.Contains(t, readmeText, "- 通过 `deploy/compose/docker-compose.yaml` 拉起 Redis 与三组服务样板；")
	require.Contains(t, readmeText, "- 先构建 `user-service`、`order-service`、`product-service` 三个镜像；")
	require.Contains(t, readmeText, "- 优先使用 `gorp/log` 作为业务日志入口；")
	require.Contains(t, readmeText, "- 传输层注册在 `internal/server`，业务编排在 `internal/service`。")
	require.NotContains(t, readmeText, "不再通过 `services/*/start.go`")
	require.NotContains(t, readmeText, "capability 调用统一放在 `internal/service`")

	userSvc, err := os.ReadFile(filepath.Join(projectDir, "services", "user", "internal", "service", "service.go"))
	require.NoError(t, err)
	require.Contains(t, string(userSvc), "framework/provider/grpc")
	require.NotContains(t, string(userSvc), "github.com/ngq/gorp/app/grpc")

	serverFiles := []string{
		"services/user/internal/server/http.go",
		"services/order/internal/server/http.go",
		"services/product/internal/server/http.go",
	}
	for _, rel := range serverFiles {
		content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		require.Contains(t, string(content), "framework/provider/gin", rel)
		require.Contains(t, string(content), "ginprovider", rel)
	}

	userPB, err := os.ReadFile(filepath.Join(projectDir, "proto", "user", "v1", "user.pb.go"))
	require.NoError(t, err)
	userPBText := string(userPB)
	require.NotContains(t, userPBText, "{{.ModuleName}}")
	require.Contains(t, userPBText, "example.com/mfwverify/proto/user/v1;userv1")

	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)
}
