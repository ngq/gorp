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
	// GenClient generates typed RPC client wrapper from proto file.
	// The generated wrapper provides type-safe method calls using the framework's RPCClient.
	//
	// GenClient 从 proto 文件生成类型化 RPC 客户端 wrapper。
	// 生成的 wrapper 使用框架的 RPCClient 提供类型安全的方法调用。
	GenClient(ctx context.Context, opts ClientGenOptions) error
	// GenService generates HTTP handler, gRPC service skeleton and route registration from proto file.
	// Enables proto-first workflow: proto → service implementation skeleton.
	//
	// GenService 从 proto 文件生成 HTTP handler、gRPC service skeleton 和路由注册。
	// 支持闭环 proto-first 工作流：proto → 服务实现骨架。
	GenService(ctx context.Context, opts ServiceGenOptions) error
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

// ClientGenOptions describes typed RPC client wrapper generation options.
//
// ClientGenOptions 描述类型化 RPC 客户端 wrapper 生成的选项。
type ClientGenOptions struct {
	// ProtoFile is the path to the proto file to parse.
	// ProtoFile 是要解析的 proto 文件路径。
	ProtoFile string

	// OutputFile is the path to write the generated Go client wrapper.
	// OutputFile 是写入生成的 Go 客户端 wrapper 的路径。
	OutputFile string

	// PackageName is the Go package name for the generated file.
	// PackageName 是生成文件的 Go package 名。
	PackageName string

	// ImportPaths are additional import paths for proto resolution.
	// ImportPaths 是 proto 解析的额外 import 路径。
	ImportPaths []string

	// ServiceName specifies which service to generate client for.
	// Empty means generate for all services in the proto file.
	// ServiceName 指定要生成客户端的服务。
	// 空表示为 proto 文件中的所有服务生成。
	ServiceName string

	// ClientPrefix is the prefix for generated client struct names.
	// Default is service name without "Service" suffix.
	// ClientPrefix 是生成的客户端 struct 名前缀。
	// 默认是去掉 "Service" 后缀的服务名。
	ClientPrefix string

	// UseGovernance indicates whether to inject governance middleware comments.
	// UseGovernance 表示是否注入治理中间件注释。
	UseGovernance bool
}

// ServiceGenOptions describes service skeleton generation from proto files.
// Supports proto-first workflow: proto → HTTP handler + gRPC service + route registration.
//
// ServiceGenOptions 描述从 proto 文件生成服务骨架的选项。
// 支持 proto-first 工作流：proto → HTTP handler + gRPC service + 路由注册。
type ServiceGenOptions struct {
	// ProtoFile is the path to the proto file to parse.
	// ProtoFile 是要解析的 proto 文件路径。
	ProtoFile string

	// OutputDir is the root directory for generated files.
	// Generated files will be placed under: OutputDir/handler/, OutputDir/service/, OutputDir/routes/.
	// OutputDir 是生成文件的根目录。
	// 生成的文件将放在：OutputDir/handler/、OutputDir/service/、OutputDir/routes/ 下。
	OutputDir string

	// PackageName is the Go package name for handler/service files.
	// PackageName 是 handler/service 文件的 Go package 名。
	PackageName string

	// Module is the Go module path (e.g., "example.com/myproject").
	// Used for import paths in generated code.
	// Module 是 Go module 路径（如 "example.com/myproject"）。
	// 用于生成代码中的 import 路径。
	Module string

	// ServiceName specifies which service to generate for.
	// Empty means generate for all services in the proto file.
	// ServiceName 指定要生成哪个服务的骨架。
	// 空表示为 proto 文件中的所有服务生成。
	ServiceName string

	// IncludeHTTP indicates whether to generate HTTP handler skeleton.
	// IncludeHTTP 是否生成 HTTP handler 骨架。
	IncludeHTTP bool

	// IncludeGRPC indicates whether to generate gRPC service skeleton.
	// IncludeGRPC 是否生成 gRPC service 骨架。
	IncludeGRPC bool

	// RegisterRoutes indicates whether to generate route registration code.
	// RegisterRoutes 是否生成路由注册代码。
	RegisterRoutes bool

	// ImportPaths are additional import paths for proto resolution.
	// ImportPaths 是 proto 解析的额外 import 路径。
	ImportPaths []string
}
