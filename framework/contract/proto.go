package contract

import (
	"context"
)

// ProtoGeneratorKey 是 Proto 生成器在容器中的绑定 key。
const ProtoGeneratorKey = "framework.proto.generator"

// ProtoGenerator Proto 生成器接口。
//
// 中文说明：
// - 支持三种工作流：Proto-first / Service-first / Route-first；
// - Proto-first：标准 protoc 生成 Go 代码；
// - Service-first：从 Go Service 接口生成 Proto；
// - Route-first：从 Gin 路由生成 Proto。
type ProtoGenerator interface {
	// GenFromProto 从 proto 文件生成 Go 代码（标准 protoc 流程）。
	//
	// 中文说明：
	// - 执行 protoc 命令生成 Go 代码；
	// - 支持生成 gRPC 服务代码；
	// - 可选生成 HTTP/gRPC 转码注解。
	GenFromProto(ctx context.Context, opts ProtoGenOptions) error

	// GenFromService 从 Go Service 接口反向生成 proto 文件。
	//
	// 中文说明：
	// - 解析 Go AST 提取接口定义；
	// - 推断请求/响应类型；
	// - 生成 proto service 和 messages；
	// - 可选添加 HTTP 注解。
	GenFromService(ctx context.Context, opts ServiceToProtoOptions) error

	// GenFromRoute 从 Gin 路由生成 proto 文件（HTTP-first）。
	//
	// 中文说明：
	// - 解析 Gin 路由定义；
	// - 推断请求/响应类型；
	// - 生成 proto service 和 messages；
	// - 自动添加 HTTP 注解。
	GenFromRoute(ctx context.Context, opts RouteToProtoOptions) error
}

// ProtoGenOptions Proto-first 生成选项。
//
// 中文说明：
// - 用于标准 protoc 生成流程；
// - 支持额外的 protoc 插件。
type ProtoGenOptions struct {
	// ProtoFiles proto 文件路径列表
	ProtoFiles []string

	// ProtoDir proto 文件目录（默认 api/proto）
	ProtoDir string

	// OutputDir 输出目录（默认与 proto 同目录）
	OutputDir string

	// IncludeHTTP 是否生成 google.api.http 注解
	IncludeHTTP bool

	// Plugins 额外的 protoc 插件
	Plugins []string

	// ImportPaths protoc 导入路径
	ImportPaths []string

	// GoOpt --go_opt 参数（如 "paths=source_relative,module=example.com"）
	GoOpt string

	// GoGrpcOpt --go-grpc_opt 参数
	GoGrpcOpt string

	// GatewayOpt --grpc-gateway_opt 参数
	GatewayOpt string

	// CustomPlugins 自定义插件（插件名 -> 参数）
	CustomPlugins map[string]string

	// JavaPackage option java_package
	JavaPackage string

	// CsharpNamespace option csharp_namespace
	CsharpNamespace string
}

// ServiceToProtoOptions Service-first 生成选项。
//
// 中文说明：
// - 从 Go Service 接口生成 Proto；
// - 支持 HTTP 注解生成。
type ServiceToProtoOptions struct {
	// ServicePath Go Service 接口文件路径
	ServicePath string

	// OutputPath 输出的 proto 文件路径
	OutputPath string

	// Package proto package 名称
	Package string

	// GoPackage Go package 路径
	GoPackage string

	// ServiceName 服务名称（默认从接口名推断）
	ServiceName string

	// IncludeHTTP 是否生成 google.api.http 注解
	IncludeHTTP bool

	// HTTPAnnotations HTTP 路径映射（方法名 -> HTTP 规则）
	HTTPAnnotations map[string]HTTPRule

	// ImportPaths 额外的 import 路径（用于跨包类型解析）
	ImportPaths []string

	// IncludeValidation 是否生成验证规则注解
	IncludeValidation bool
}

// RouteToProtoOptions Route-first 生成选项。
//
// 中文说明：
// - 从 Gin 路由生成 Proto；
// - 自动推断 HTTP 注解。
type RouteToProtoOptions struct {
	// RouteFile Gin 路由定义文件路径
	RouteFile string

	// HandlerFile Handler 定义文件路径（用于类型推断）
	HandlerFile string

	// OutputPath 输出的 proto 文件路径
	OutputPath string

	// Package proto package 名称
	Package string

	// GoPackage Go package 路径
	GoPackage string

	// ServiceName gRPC service 名称
	ServiceName string

	// BasePath HTTP 基础路径（如 "/api/v1"）
	BasePath string

	// ImportPaths 额外的 import 路径
	ImportPaths []string
}

// HTTPRule HTTP/gRPC 转码规则。
//
// 中文说明：
// - 定义 gRPC 方法到 HTTP 接口的映射；
// - 遵循 google.api.http 规范。
type HTTPRule struct {
	// Method HTTP 方法：GET/POST/PUT/DELETE/PATCH
	Method string

	// Path HTTP 路径模板（如 "/v1/users/{user_id}"）
	Path string

	// Body 请求体字段名（"*" 表示全部）
	Body string

	// ResponseBody 响应体字段名
	ResponseBody string

	// AdditionalBindings 额外的 HTTP 绑定
	AdditionalBindings []*HTTPRule
}

// ProtoGeneratorConfig Proto 生成器配置。
//
// 中文说明：
// - 定义 Proto 生成器的启用状态和默认配置。
type ProtoGeneratorConfig struct {
	// Enabled 是否启用
	Enabled bool

	// Strategy 策略：noop/protoc
	Strategy string

	// DefaultProtoDir 默认 proto 目录
	DefaultProtoDir string

	// DefaultOutputDir 默认输出目录
	DefaultOutputDir string

	// IncludeHTTPAnnotation 默认是否包含 HTTP 注解
	IncludeHTTPAnnotation bool

	// ThirdPartyPaths 第三方 proto 文件路径
	ThirdPartyPaths []string
}

// ServiceDef 服务定义（AST 解析结果）。
type ServiceDef struct {
	// Name 服务名称
	Name string

	// Methods 方法定义列表
	Methods []MethodDef

	// Comments 注释
	Comments []string
}

// MethodDef 方法定义。
type MethodDef struct {
	// Name 方法名称
	Name string

	// RequestType 请求类型
	RequestType *TypeDef

	// ResponseType 响应类型
	ResponseType *TypeDef

	// Comments 注释
	Comments []string

	// HTTPRule HTTP 规则（可选）
	HTTPRule *HTTPRule

	// RequestStream 是否为客户端流式 RPC
	RequestStream bool

	// ResponseStream 是否为服务端流式 RPC
	ResponseStream bool
}

// TypeDef 类型定义。
type TypeDef struct {
	// Name 类型名称
	Name string

	// Package 所属包
	Package string

	// IsPointer 是否为指针类型
	IsPointer bool

	// IsSlice 是否为切片类型
	IsSlice bool

	// IsMap 是否为 map 类型
	IsMap bool

	// MapKey Map 的 key 类型（仅当 IsMap=true 时有效）
	MapKey *TypeDef

	// MapValue Map 的 value 类型（仅当 IsMap=true 时有效）
	MapValue *TypeDef

	// Fields 字段列表（仅结构体）
	Fields []FieldDef

	// Comments 结构体注释（仅结构体）
	Comments []string

	// IsEnum 是否为枚举类型
	IsEnum bool

	// EnumValues 枚举值（仅当 IsEnum=true 时有效）
	EnumValues []EnumValue
}

// FieldDef 字段定义。
type FieldDef struct {
	// Name 字段名称
	Name string

	// JSONName JSON tag 名称
	JSONName string

	// ProtoName proto 字段名（snake_case）
	ProtoName string

	// Type 字段类型
	Type *TypeDef

	// Tag 验证标签
	Tag string

	// Remark 字段描述（优先从 remark tag 获取）
	Remark string

	// Comments 注释（来自行尾注释）
	Comments []string

	// ProtoNumber proto 字段编号（自动分配）
	ProtoNumber int

	// ValidationRules 验证规则（从 binding/validate tag 解析）
	ValidationRules []ValidationRule

	// DefaultValue 默认值
	DefaultValue string

	// IsOptional 是否为 optional 字段
	IsOptional bool
}

// EnumValue 枚举值定义。
type EnumValue struct {
	// Name 枚举名称
	Name string

	// Value 枚举值
	Value int32

	// Comments 注释
	Comments []string
}

// ValidationRule 验证规则。
type ValidationRule struct {
	// Rule 规则名称（如 "required", "min", "max", "email"）
	Rule string

	// Value 规则值（如 min=6 中的 6）
	Value interface{}

	// Message 错误消息
	Message string
}

// RouteDef 路由定义。
type RouteDef struct {
	// Method HTTP 方法
	Method string

	// Path HTTP 路径
	Path string

	// HandlerName 处理函数名称
	HandlerName string

	// RequestType 请求类型（推断）
	RequestType *TypeDef

	// ResponseType 响应类型（推断）
	ResponseType *TypeDef

	// Comments 注释
	Comments []string

	// HandlerFile Handler 所在文件路径
	HandlerFile string
}

// ImportDef Import 定义。
type ImportDef struct {
	// Path import 路径
	Path string

	// Public 是否为 public import
	Public bool

	// Weak 是否为 weak import
	Weak bool
}
