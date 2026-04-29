package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestMultiFlatWireDocsLinkToDeployAssets(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwdocs")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwdocs",
		Module:          "example.com/mfwdocs",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	deployDoc, err := os.ReadFile(filepath.Join(projectDir, "docs", "deploy.md"))
	require.NoError(t, err)
	nextSteps, err := os.ReadFile(filepath.Join(projectDir, "docs", "next-steps.md"))
	require.NoError(t, err)

	readmeText := string(readme)
	deployText := string(deployDoc)
	nextText := string(nextSteps)

	require.Contains(t, readmeText, "docs/deploy.md")
	require.Contains(t, readmeText, "看：`docs/deploy.md`")
	require.NotContains(t, readmeText, "kubectl kustomize deploy/kubernetes/overlays/prod")
	require.Contains(t, readmeText, "修改 `proto/*.proto` 后")
	require.NotContains(t, readmeText, "母仓/单服务场景")
	require.Contains(t, readmeText, "多服务起步模板（Wire 装配版）")
	require.Contains(t, readmeText, "> 起步：")
	require.Contains(t, readmeText, "## 先看这三件事")
	require.Contains(t, readmeText, "- 创建项目：`gorp new multi-wire`")
	require.Contains(t, readmeText, "- 生成代码：`make generate`")
	require.Contains(t, readmeText, "目录：")
	require.Contains(t, readmeText, "## 日志建议")
	require.NotContains(t, readmeText, "安装代码生成工具")
	require.Contains(t, readmeText, "### 2. 统一生成代码")
	require.Equal(t, 1, strings.Count(readmeText, "### 5. 本地联调"))
	require.Contains(t, readmeText, "### 4. 运行测试")
	require.Contains(t, readmeText, "make test")
	require.Contains(t, readmeText, "### 5. 本地联调")
	require.Contains(t, readmeText, "make deploy-local")
	require.Contains(t, readmeText, "看：`docs/deploy.md`")
	require.NotContains(t, readmeText, "附带：")
	require.NotContains(t, readmeText, "验：")
	require.Contains(t, readmeText, "关键配置：")
	require.Contains(t, readmeText, "主线入口：")
	require.Contains(t, readmeText, "### 8. 验证 Proto-first gRPC 双服务链路")
	require.Contains(t, readmeText, "先确认一条服务间调用链路能跑通")
	require.Contains(t, readmeText, "关键配置：")
	require.Contains(t, readmeText, "主线入口：")
	require.Contains(t, readmeText, "### 9. 生成 proto 代码")
	require.Contains(t, readmeText, "修改 `proto/*.proto` 后")
	require.Contains(t, readmeText, "生成：")
	require.Contains(t, readmeText, "默认优先按服务间 Proto-first 接线验证；")
	require.Contains(t, readmeText, "这一步按需执行。")
	require.NotContains(t, readmeText, "third_party/")
	require.NotContains(t, readmeText, "消息与事件组件")

	structure, err := os.ReadFile(filepath.Join(projectDir, "docs", "structure.md"))
	require.NoError(t, err)
	structureText := string(structure)
	require.Contains(t, structureText, "其余基础设施接入细节，优先通过 framework capability / provider 解决")
	require.NotContains(t, structureText, "当前阶段为什么只新增 `docs/`")
	require.NotContains(t, structureText, "正确理解")

	require.Contains(t, readmeText, "- 统一执行 Wire 与 Proto 生成；")
	require.Contains(t, readmeText, "- 各服务统一由 `services/<service>/cmd/main.go` 启动；")
	require.Contains(t, readmeText, "- 通过 `deploy/compose/docker-compose.yaml` 拉起 Redis 与三组服务样板；")
	require.Contains(t, readmeText, "- 先构建 `user-service`、`order-service`、`product-service` 三个镜像；")
	require.Contains(t, readmeText, "- 优先使用 `gorp/log` 作为业务日志入口；")
	require.Contains(t, readmeText, "- gRPC 默认主线是 Proto-first")
	require.Contains(t, readmeText, "### 7. 部署资产")
	require.Contains(t, nextText, "docs/deploy.md")
	require.Contains(t, nextText, "建议按这个顺序继续：")
	require.Contains(t, nextText, "需要生成代码时执行 `make generate`")
	require.Contains(t, nextText, "默认附带 Docker 开发部署与 Kubernetes 部署资产")
	require.NotContains(t, nextText, "部署层不再需要你从母仓手工拷贝")
	require.Contains(t, deployText, "## Docker 开发部署")
	require.NotContains(t, deployText, "当前模板默认提供")
	require.Contains(t, deployText, "本地联调：")
	require.Contains(t, deployText, "会先构建三组服务镜像，再统一 tag 与 push。")
}
