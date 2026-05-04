package integration

import "context"

const ProtoGeneratorKey = "framework.proto.generator"

type ProtoGenerator interface {
	GenFromProto(ctx context.Context, opts ProtoGenOptions) error
	GenFromService(ctx context.Context, opts ServiceToProtoOptions) error
	GenFromRoute(ctx context.Context, opts RouteToProtoOptions) error
}

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

type HTTPRule struct {
	Method             string
	Path               string
	Body               string
	ResponseBody       string
	AdditionalBindings []*HTTPRule
}

type ProtoGeneratorConfig struct {
	Enabled               bool
	Strategy              string
	DefaultProtoDir       string
	DefaultOutputDir      string
	IncludeHTTPAnnotation bool
	ThirdPartyPaths       []string
}
