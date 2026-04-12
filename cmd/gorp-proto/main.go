// gorp-proto - Proto 生成器 CLI 工具
//
// 支持三种工作流：
// - from-service: Go Service 接口 → Proto
// - from-proto: Proto → pb.go (调用 protoc)
// - from-route: Gin 路由 → Proto (实验性)
//
// 安装:
//   go install github.com/ngq/gorp/cmd/gorp-proto@latest
//
// 使用:
//   gorp-proto from-service --service-path ./service.go --output ./proto/
//   gorp-proto from-proto --proto-dir ./proto --output ./pb/
package main

import (
	"os"

	"github.com/ngq/gorp/cmd/gorp-proto/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}