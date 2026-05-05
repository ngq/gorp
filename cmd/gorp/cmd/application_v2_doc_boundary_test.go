package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplicationV2DocDeclaresApplicationFirstUserLayering(t *testing.T) {
	content, err := os.ReadFile("E:/project/gin_plantfrom/.private-docs/manual/v2/application/gorp-platform-kernel-above-application-复杂能力定义与简化路径.zh-CN.md")
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "## 5.4 用户分层使用口径")
	require.Contains(t, text, "application 是统一主入口，面向普通用户，也面向大多数高级用户；kernel 是扩展与回退层，不是高级用户默认入口。")
	require.Contains(t, text, "Kratos 不是“复杂场景才使用的 application”，而是“默认高层入口 + 允许复杂场景下探底层”的框架。")
	require.Contains(t, text, "普通用户和大多数高级用户，都先走 application-first 默认路径")
	require.Contains(t, text, "## 9.1 借鉴范围：只借鉴能力域，不复制框架形态")
	require.Contains(t, text, "Kratos 只作为能力对照表，不作为 gorp 的架构蓝图。")
	require.Contains(t, text, "## 9.2 禁止事项：不把 v2 写成 Kratos 翻版方案")
	require.Contains(t, text, "v2 的目标不是“补齐 Kratos 的目录”")
	require.Contains(t, text, "kernel capability -> application declaration -> starter consumption")
	require.Contains(t, text, "## 9.4 contrib 候选维护原则与缺口清单")
	require.Contains(t, text, "Kratos contrib 只用于帮助我们盘点缺口，不用于反向定义 gorp contrib 地图。")
	require.Contains(t, text, "不按 `.tmp/kratos/contrib` 逐目录补齐")
	require.Contains(t, text, "contrib/errortracker/sentry")
	require.Contains(t, text, "contrib/encoding/msgpack")
	require.Contains(t, text, "## 9.5 当前 contrib 占位能力清单")
	require.Contains(t, text, "`contrib/registry/kubernetes`")
	require.Contains(t, text, "`contrib/configsource/apollo`")
	require.Contains(t, text, "`contrib/dtm/dtmsdk`")
	require.Contains(t, text, "### 9.5.5 与 `framework/provider` 的相关性对照")
	require.Contains(t, text, "`framework/provider/configsource/local`")
	require.Contains(t, text, "`framework/provider/discovery/noop`")
	require.Contains(t, text, "它们是“有意设计的默认实现”，不是“没做完的外部能力适配”。")
	require.Contains(t, text, "这些 contrib 的能力位是合理的，但当前不能把“目录存在”视为“能力完成”。")
}

func TestApplicationV2ChecklistKeepsExternalReferenceBoundary(t *testing.T) {
	content, err := os.ReadFile("E:/project/gin_plantfrom/.private-docs/manual/v2/application/gorp-platform-kernel-above-application-P0-执行checklist（按文件级）.zh-CN.md")
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "J. 外部参照边界治理")
	require.Contains(t, text, "Kratos 只作为能力对照表，不作为 gorp 架构蓝图")
	require.Contains(t, text, "禁止以 Kratos 顶层目录反推 gorp 顶层目录")
	require.Contains(t, text, "transport / middleware / proto-first 只作为能力域参照，不作为目录复制任务")
	require.Contains(t, text, "kernel capability -> application declaration -> starter consumption")
	require.Contains(t, text, "10 个能力域全部完成")
}

func TestRuntimeV2DocDeclaresHostLifecycleBoundary(t *testing.T) {
	content, err := os.ReadFile("E:/project/gin_plantfrom/.private-docs/manual/v2/application/gorp-runtime-定义清单与语义说明.zh-CN.md")
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "## 0.0.1 Host / Lifecycle 边界声明（HTTP / Cron / gRPC）")
	require.Contains(t, text, "Host 是 kernel 层生命周期编排能力，不是业务侧运行入口托管")
	require.Contains(t, text, "HTTP 已经消费它做优雅关闭；Cron / gRPC 具备接入适配器，但不等于默认由 application 托管")
	require.Contains(t, text, "`framework/bootstrap/http_service.go` 已经使用 `container.MakeHost(...)`")
}
