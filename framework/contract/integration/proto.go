// Application scenarios:
// - Define the proto generation contract used by code generation and route/service export flows.
// - Support generating proto definitions from proto files, services, and route declarations.
// - Provide one shared config and option model for generator implementations.
//
// 适用场景：
// - 定义代码生成和路由/服务导出流程使用的 proto 生成契约。
// - 支持从 proto 文件、服务定义和路由声明生成 proto。
// - 为生成器实现提供统一配置与选项模型。
package integration

import "context"

// ProtoGeneratorKey is the container key for the proto generator capability.
//
// ProtoGeneratorKey 是 proto generator 能力的容器键。
const ProtoGeneratorKey = "framework.proto.generator"

// ProtoGenerator defines the proto generation contract.
//
// ProtoGenerator 定义 proto 生成契约。
type ProtoGenerator interface {
	GenFromProto(ctx context.Context, opts ProtoGenOptions) error
	GenFromService(ctx context.Context, opts ServiceToProtoOptions) error
	GenFromRoute(ctx context.Context, opts RouteToProtoOptions) error
}

// ProtoGenOptions describes generation from proto files.
//
// ProtoGenOptions 描述从 proto 文件生成时的选项。
type ProtoGenOptions struct {
	ProtoFiles      []string
	ProtoDir        string
	OutputDir       string
	IncludeHTTP     bool
	Plugins         []string
	ImportPaths     []string
	GoOpt           string
	GoGrpcOpt       string
	GatewayOpt      string
	CustomPlugins   map[string]string
	JavaPackage     string
	CsharpNamespace string
}

// ServiceToProtoOptions describes generation from service definitions.
//
// ServiceToProtoOptions 描述从服务定义生成 proto 的选项。
type ServiceToProtoOptions struct {
	ServicePath       string
	OutputPath        string
	Package           string
	GoPackage         string
	ServiceName       string
	IncludeHTTP       bool
	HTTPAnnotations   map[string]HTTPRule
	ImportPaths       []string
	IncludeValidation bool
}

// RouteToProtoOptions describes generation from route declarations.
//
// RouteToProtoOptions 描述从路由声明生成 proto 的选项。
type RouteToProtoOptions struct {
	RouteFile   string
	HandlerFile string
	OutputPath  string
	Package     string
	GoPackage   string
	ServiceName string
	BasePath    string
	ImportPaths []string
}

// HTTPRule describes one HTTP binding rule.
//
// HTTPRule 描述一条 HTTP 绑定规则。
type HTTPRule struct {
	Method             string
	Path               string
	Body               string
	ResponseBody       string
	AdditionalBindings []*HTTPRule
}

// ProtoGeneratorConfig describes proto generation runtime configuration.
//
// ProtoGeneratorConfig 描述 proto 生成运行时配置。
type ProtoGeneratorConfig struct {
	Enabled               bool
	Strategy              string
	DefaultProtoDir       string
	DefaultOutputDir      string
	IncludeHTTPAnnotation bool
	ThirdPartyPaths       []string
}
