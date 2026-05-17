package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// createCmd 是 create 子命令的父命令。
var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create service scaffolds for custom development",
	GroupID: commandGroupCodegen,
	Long: `Create service scaffolds for custom development.

The create command generates lightweight service skeletons that let you
organize your own business directory structure. Unlike proto-generated
services with DDD structure, these skeletons provide minimal HTTP contract
setup with mono governance capabilities pre-configured.

Available subcommands:
  gorp create service <name>  Create a custom service skeleton

This is designed for:
  - Tool services (WeChat adapter, Alipay integration, etc.)
  - API gateways
  - Custom business services with non-DDD structure

For standard business services following DDD patterns, use proto generation
with the multi-service templates instead.`,
}

// createServiceCmd 创建自定义 service 骨架。
var createServiceCmd = &cobra.Command{
	Use:   "service <name>",
	Short: "Create a custom service skeleton with mono governance",
	Long: `Create a custom service skeleton with mono governance.

This generates a lightweight service structure without DDD layers.
You can freely organize your business directories and code structure.

Generated structure:
  services/<name>/
  ├── cmd/app/main.go      # Entry point with mono governance
  ├── config/app.yaml      # Configuration file
  ├── handler/             # HTTP handlers
  │   └── handler.go       # Sample handler
  ├── routes.go            # Route registration
  ├── go.mod               # Go module
  └── README.md

The generated service includes:
  - HTTP contract mode (gorp.HTTPContext) by default
  - Mono governance (RequestIdentity, Logging, Recovery, Timeout, Metrics)
  - Health check endpoint

Use --http=gin for Gin native mode if preferred.

Examples:
  gorp create service user-api
  gorp create service wechat-adapter --http=gin
  gorp create service payment-gateway`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateService,
}

var createServiceHTTP string
var createServicePath string

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createServiceCmd)

	createServiceCmd.Flags().StringVar(&createServiceHTTP, "http", "", "HTTP mode: contract (default), gin")
	createServiceCmd.Flags().StringVar(&createServicePath, "path", "services", "parent directory for the service (default: services)")
}

func runCreateService(cmd *cobra.Command, args []string) error {
	serviceName := strings.TrimSpace(args[0])
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}

	// 校验服务名
	if err := validateServiceName(serviceName); err != nil {
		return err
	}

	// 规范化 HTTP 模式
	httpMode := normalizeHTTPMode(createServiceHTTP)

	// 确定目标目录
	targetDir := filepath.Join(createServicePath, serviceName)

	// 检查目录是否已存在
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory already exists: %s", targetDir)
	}

	// 获取模块路径
	in := bufio.NewReader(cmd.InOrStdin())
	modulePath, err := promptModulePath(in, cmd.OutOrStdout(), serviceName)
	if err != nil {
		return err
	}

	// 生成骨架
	if err := generateServiceSkeleton(targetDir, serviceName, modulePath, httpMode); err != nil {
		return fmt.Errorf("generate service skeleton: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created service: %s\n", targetDir)
	fmt.Fprintf(cmd.OutOrStdout(), "Next:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  cd %s\n", targetDir)
	fmt.Fprintf(cmd.OutOrStdout(), "  go mod tidy\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  go run ./cmd/app\n")

	return nil
}

// validateServiceName 校验服务名是否合法。
func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("service name must not contain path separators")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid service name: %s", name)
	}
	return nil
}

// promptModulePath 提示用户输入模块路径。
func promptModulePath(r *bufio.Reader, out io.Writer, serviceName string) (string, error) {
	defaultModule := fmt.Sprintf("github.com/example/%s", serviceName)
	fmt.Fprintf(out, "请输入模块路径 (默认: %s): ", defaultModule)

	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return defaultModule, nil
	}
	return line, nil
}

// generateServiceSkeleton 生成 service 骨架文件。
func generateServiceSkeleton(targetDir, serviceName, modulePath, httpMode string) error {
	// 创建目录结构
	dirs := []string{
		filepath.Join(targetDir, "cmd/app"),
		filepath.Join(targetDir, "config"),
		filepath.Join(targetDir, "handler"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// 生成文件
	data := serviceSkeletonData{
		ServiceName: serviceName,
		ModulePath:  modulePath,
		HTTPMode:    httpMode,
		IsGinMode:   httpMode == "gin",
	}

	files := map[string]string{
		"cmd/app/main.go":      serviceMainTemplate,
		"config/app.yaml":      serviceConfigTemplate,
		"handler/handler.go":   serviceHandlerTemplate,
		"routes.go":            serviceRoutesTemplate,
		"go.mod":               serviceGoModTemplate,
		"README.md":            serviceReadmeTemplate,
	}

	for filename, template := range files {
		filePath := filepath.Join(targetDir, filename)
		content := executeServiceTemplate(template, data)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write file %s: %w", filePath, err)
		}
	}

	return nil
}

// serviceSkeletonData 是模板数据。
type serviceSkeletonData struct {
	ServiceName string
	ModulePath  string
	HTTPMode    string
	IsGinMode   bool
}

// executeServiceTemplate 简单模板替换。
func executeServiceTemplate(template string, data serviceSkeletonData) string {
	result := template
	result = strings.ReplaceAll(result, "{{.ServiceName}}", data.ServiceName)
	result = strings.ReplaceAll(result, "{{.ModulePath}}", data.ModulePath)
	result = strings.ReplaceAll(result, "{{.HTTPMode}}", data.HTTPMode)
	if data.IsGinMode {
		result = strings.ReplaceAll(result, "{{.HTTPContextType}}", "*gin.Context")
		result = strings.ReplaceAll(result, "{{.HTTPRouterType}}", "*gin.Engine")
		result = strings.ReplaceAll(result, "{{.ImportGorp}}", "")
		result = strings.ReplaceAll(result, "{{.ImportGin}}", "\n\t\"github.com/gin-gonic/gin\"")
	} else {
		result = strings.ReplaceAll(result, "{{.HTTPContextType}}", "gorp.HTTPContext")
		result = strings.ReplaceAll(result, "{{.HTTPRouterType}}", "gorp.HTTPRouter")
		result = strings.ReplaceAll(result, "{{.ImportGorp}}", "")
		result = strings.ReplaceAll(result, "{{.ImportGin}}", "")
	}
	return result
}

// 模板定义
const serviceMainTemplate = `package main

import (
	"fmt"
	"os"

	gorp "github.com/ngq/gorp"{{.ImportGin}}
)

func main() {
	if err := gorp.Run(
		"{{.ServiceName}}",
		gorp.WithMonoGovernance(),
		gorp.WithSetup(setup),
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setup(rt *gorp.HTTPRuntime) error {
	// 在这里初始化你的业务依赖
	// 例如：数据库连接、Redis 客户端、第三方 SDK 等

	// 注册 HTTP 路由
	registerRoutes(rt.Router, rt)

	// 如需注册 gRPC 服务：
	// registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	// if err != nil {
	//     return err
	// }
	// return registrar.RegisterProto(func(server *grpc.Server) error {
	//     // pb.RegisterYourServiceServer(server, yourService)
	//     return nil
	// })

	return nil
}
`

const serviceConfigTemplate = `# {{.ServiceName}} 配置文件
service:
  name: {{.ServiceName}}

server:
  http:
    addr: :8080

# 微服务治理配置（按需启用）
# tracing:
#   enabled: true
#   endpoint: localhost:4317
#
# discovery:
#   enabled: true
#   endpoints:
#     - localhost:2379
`

const serviceHandlerTemplate = `package handler

import ({{.ImportGin}}
)

// Hello 示例 handler，可根据需要修改或删除。
// 你可以自由组织 handler 目录结构，例如：
// - handler/user.go
// - handler/order.go
// - handler/wechat/adapter.go
func Hello(c {{.HTTPContextType}}) {
	c.JSON(200, map[string]any{
		"message": "hello from {{.ServiceName}}",
	})
}
`

const serviceRoutesTemplate = `package main

import (
	gorp "github.com/ngq/gorp"{{.ImportGin}}
	"{{.ModulePath}}/handler"
)

// registerRoutes 注册 HTTP 路由。
// 在这里添加你的业务路由。
func registerRoutes(r {{.HTTPRouterType}}, rt *gorp.HTTPRuntime) {
	// 示例路由
	r.GET("/hello", handler.Hello)

	// 健康检查
	r.GET("/healthz", func(c {{.HTTPContextType}}) {
		c.JSON(200, map[string]any{
			"status":  "healthy",
			"service": "{{.ServiceName}}",
		})
	})

	// 在这里添加更多路由...
}
`

const serviceGoModTemplate = `module {{.ModulePath}}

go 1.22

require github.com/ngq/gorp v0.0.0
`

const serviceReadmeTemplate = `# {{.ServiceName}}

自定义服务骨架，默认使用单体治理（轻量）。

## 目录结构

` + "```" + `
services/{{.ServiceName}}/
├── cmd/app/main.go      # 服务入口，单体治理
├── config/app.yaml      # 配置文件
├── handler/             # HTTP handlers（自由发挥）
│   └── handler.go
├── routes.go            # 路由注册
├── go.mod
└── README.md
` + "```" + `

## 启动

` + "```bash" + `
cd services/{{.ServiceName}}
go mod tidy
go run ./cmd/app
` + "```" + `

## 业务代码组织

handler 目录由你自由发挥，例如：

` + "```" + `
handler/
├── user.go           # 用户相关
├── order.go          # 订单相关
├── wechat/           # 企业微信适配
│   └── adapter.go
└── alipay/           # 支付宝适配
    └── adapter.go
` + "```" + `

如需 DDD 结构，可自行创建：
` + "```" + `
internal/
├── biz/      # 业务逻辑
├── data/     # 数据访问
└── service/  # 服务层
` + "```" + `

## 治理能力

默认使用单体治理（轻量）：
- RequestIdentity、Logging、Recovery、Timeout、Metrics

如需微服务治理，修改 main.go：
` + "```go" + `
gorp.Run("{{.ServiceName}}",
    gorp.GRPC(),
    gorp.WithMicroserviceGovernance(), // 替换 WithMonoGovernance
    gorp.WithSetup(setup),
)
` + "```" + `

微服务治理包括：
- Tracing（OpenTelemetry）
- Service Discovery（etcd）
- Circuit Breaker（Sentinel）
`
