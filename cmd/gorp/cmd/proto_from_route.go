package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ngq/gorp/framework/contract"
	"github.com/spf13/cobra"
)

var (
	// from-route flags
	routeFile   string
	handlerFile string
	basePath    string
)

// protoFromRouteCmd from-route 子命令：从 Gin 路由生成 Proto 文件（实验性）
var protoFromRouteCmd = &cobra.Command{
	Use:   "from-route",
	Short: "从 Gin 路由生成 Proto 文件（实验性）",
	Long: `从 Gin 路由定义反向生成 Proto 文件。

此功能为实验性功能，当前状态：
- ✅ 路由解析框架完成
- ⚠️ Handler 类型推断需要手动指定或使用约定

解析内容：
- Gin 路由定义（GET/POST/PUT/DELETE）
- 路径参数提取（如 /users/{id}）
- 自动生成 HTTP 注解

使用:
  gorp proto from-route -r ./routes.go -o ./proto/
  gorp proto from-route -r ./routes.go -h ./handlers.go -o ./proto/api.proto

示例 Gin 路由:
  r.GET("/users/:id", handler.GetUser)
  r.POST("/users", handler.CreateUser)
  r.PUT("/users/:id", handler.UpdateUser)
  r.DELETE("/users/:id", handler.DeleteUser)`,
	RunE: runProtoFromRoute,
}

func init() {
	protoCmd.AddCommand(protoFromRouteCmd)

	protoFromRouteCmd.Flags().StringVarP(&routeFile, "route-file", "r", "", "Gin 路由定义文件路径（必填）")
	protoFromRouteCmd.Flags().StringVarP(&handlerFile, "handler-file", "h", "", "Handler 定义文件路径（用于类型推断）")
	protoFromRouteCmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出 Proto 文件路径（必填）")
	protoFromRouteCmd.Flags().StringVarP(&protoPackage, "package", "p", "", "Proto package 名称")
	protoFromRouteCmd.Flags().StringVar(&goPackage, "go-package", "", "Go package 路径")
	protoFromRouteCmd.Flags().StringVar(&serviceName, "service-name", "", "Service 名称")
	protoFromRouteCmd.Flags().StringVar(&basePath, "base-path", "", "HTTP 基础路径（如 /api/v1）")
	protoFromRouteCmd.Flags().StringSliceVar(&importPathsS, "import-paths", nil, "额外的 import 路径")

	protoFromRouteCmd.MarkFlagRequired("route-file")
	protoFromRouteCmd.MarkFlagRequired("output")
}

func runProtoFromRoute(cmd *cobra.Command, args []string) error {
	// 确保输出目录存在
	if err := ensureProtoDir(outputDir); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 推断默认值
	pkg := protoPackage
	if pkg == "" {
		base := filepath.Base(outputDir)
		if ext := filepath.Ext(base); ext != "" {
			pkg = base[:len(base)-len(ext)]
		} else {
			pkg = "api"
		}
	}

	gopkg := goPackage
	if gopkg == "" {
		gopkg = fmt.Sprintf("./%s;%s", filepath.Dir(outputDir), pkg)
		fmt.Printf("⚠️  go-package 未指定，使用推断值: %s\n", gopkg)
	}

	sn := serviceName
	if sn == "" {
		sn = pkg + "Service"
		fmt.Printf("⚠️  service-name 未指定，使用推断值: %s\n", sn)
	}

	// 创建生成器
	gen, err := createProtoGenerator(true) // from-route 默认包含 HTTP 注解
	if err != nil {
		return fmt.Errorf("创建生成器失败: %w", err)
	}

	// 构建选项
	opts := contract.RouteToProtoOptions{
		RouteFile:   routeFile,
		HandlerFile: handlerFile,
		OutputPath:  outputDir,
		Package:     pkg,
		GoPackage:   gopkg,
		ServiceName: sn,
		BasePath:    basePath,
		ImportPaths: importPathsS,
	}

	// 执行生成
	fmt.Printf("🔄 正在解析路由: %s\n", routeFile)
	fmt.Printf("⚠️  此功能为实验性功能，请求/响应类型可能需要手动调整\n")

	if err := gen.GenFromRoute(context.Background(), opts); err != nil {
		printProtoError("Route→Proto", err)
		return err
	}

	printProtoSuccess("Route→Proto", outputDir)
	return nil
}
