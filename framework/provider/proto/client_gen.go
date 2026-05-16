// Package proto provides proto generator implementation.
// This file implements typed RPC client wrapper generation from proto files.
//
// Proto 包提供 proto 生成器实现。
// 本文件实现从 proto 文件生成类型化 RPC 客户端 wrapper。
package proto

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ProtoService describes one service parsed from proto file.
//
// ProtoService 描述从 proto 文件解析出的一个服务。
type ProtoService struct {
	Name    string
	Methods []ProtoMethod
}

// ProtoMethod describes one method parsed from proto service.
//
// ProtoMethod 描述从 proto 服务解析出的一个方法。
type ProtoMethod struct {
	Name         string
	RequestType  string
	ResponseType string
	InputStream  bool
	OutputStream bool
	// HTTP routing info from google.api.http annotation.
	// HTTP 路由信息，从 google.api.http 注解解析。
	HTTPMethod string // GET, POST, PUT, DELETE, PATCH
	HTTPPath   string // /v1/users/{id}
	HTTPBody   string // "*" or field name for body mapping
	// Auth and middleware info from gorp options.
	// 认证和中间件信息，从 gorp 选项解析。
	AuthRequired bool     // 是否要求认证
	AuthRoles    []string // 要求的角色
	AuthSkip     bool     // 跳过认证
	Middleware   []string // 中间件名称列表
	RateLimitRPS int32    // 每秒请求数限制
	RateLimitRPM int32    // 每分钟请求数限制
	RateLimitKey string   // 限流键
}

// GenClient generates typed RPC client wrapper from proto file.
// Parses the proto file to extract service and method definitions,
// then generates a Go file with typed client struct and methods.
//
// GenClient 从 proto 文件生成类型化 RPC 客户端 wrapper。
// 解析 proto 文件提取服务和方法的定义，
// 然后生成包含类型化客户端 struct 和方法的 Go 文件。
func (g *Generator) GenClient(ctx context.Context, opts integrationcontract.ClientGenOptions) error {
	if opts.ProtoFile == "" {
		return errors.New("proto file path is required")
	}
	if opts.OutputFile == "" {
		return errors.New("output file path is required")
	}
	if opts.PackageName == "" {
		// Derive package name from output directory
		// 从输出目录推导 package 名
		opts.PackageName = filepath.Base(filepath.Dir(opts.OutputFile))
		if opts.PackageName == "" || opts.PackageName == "." {
			opts.PackageName = "client"
		}
	}

	// Parse proto file to extract services
	// 解析 proto 文件提取服务
	services, err := parseProtoFile(opts.ProtoFile)
	if err != nil {
		return fmt.Errorf("parse proto file failed: %w", err)
	}

	// 校验所有服务名和方法名的标识符合法性，防止代码注入。
	for _, svc := range services {
		if !isValidGoIdentifier(svc.Name) {
			return fmt.Errorf("invalid service name %q: must match [A-Za-z_][A-Za-z0-9_]*", svc.Name)
		}
		for _, m := range svc.Methods {
			if !isValidGoIdentifier(m.Name) {
				return fmt.Errorf("invalid method name %q in service %q: must match [A-Za-z_][A-Za-z0-9_]*", m.Name, svc.Name)
			}
		}
	}

	// Filter service if specified
	// 如果指定了服务名则过滤
	if opts.ServiceName != "" {
		filtered := []ProtoService{}
		for _, svc := range services {
			if svc.Name == opts.ServiceName {
				filtered = append(filtered, svc)
			}
		}
		services = filtered
	}

	if len(services) == 0 {
		return errors.New("no services found in proto file")
	}

	// Generate client wrapper code
	// 生成客户端 wrapper 代码
	code := generateClientWrapperCode(services, opts)

	// Write output file
	// 写入输出文件
	outputDir := filepath.Dir(opts.OutputFile)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("create output directory failed: %w", err)
		}
	}

	if err := os.WriteFile(opts.OutputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("write output file failed: %w", err)
	}

	return nil
}

// parseProtoFile parses a proto file and extracts service definitions.
// Uses regex-based parsing for simplicity and portability.
// Also parses google.api.http annotations for HTTP routing.
//
// parseProtoFile 解析 proto 文件并提取服务定义。
// 使用基于正则的解析以保持简单和可移植性。
// 同时解析 google.api.http 注解获取 HTTP 路由信息。
func parseProtoFile(protoFile string) ([]ProtoService, error) {
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return nil, fmt.Errorf("read proto file failed: %w", err)
	}

	services := []ProtoService{}
	contentStr := string(content)

	// Regex patterns for service and method definitions
	// 用于匹配 service 和 method 定义的正则模式
	// 匹配 service 块，包含方法定义
	servicePattern := regexp.MustCompile(`service\s+(\w+)\s*\{([\s\S]*?)\n\}`)
	// 匹配 rpc 方法定义，包括方法体（option 块）
	methodPattern := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(stream\s+)?(\w+)\s*\)\s*returns\s*\(\s*(stream\s+)?(\w+)\s*\)\s*(\{([^}]*)\})?;?`)

	// Find all services
	// 查找所有服务
	serviceMatches := servicePattern.FindAllStringSubmatch(contentStr, -1)
	for _, svcMatch := range serviceMatches {
		serviceName := svcMatch[1]
		serviceBody := svcMatch[2]

		methods := []ProtoMethod{}

		// Find all methods in the service
		// 查找服务中的所有方法
		methodMatches := methodPattern.FindAllStringSubmatch(serviceBody, -1)
		for i := range methodMatches {
			methodMatch := methodMatches[i]
			methodName := methodMatch[1]
			inputStream := strings.TrimSpace(methodMatch[2]) != ""
			requestType := methodMatch[3]
			outputStream := strings.TrimSpace(methodMatch[4]) != ""
			responseType := methodMatch[5]
			methodBody := methodMatch[7] // 可能为空

			// Parse google.api.http annotation from method body
			// 从方法体解析 google.api.http 注解
			httpMethod, httpPath, httpBody := parseHTTPAnnotation(methodBody)

			// Parse gorp auth/middleware options from method body
			// 从方法体解析 gorp 认证/中间件选项
			authRequired, authRoles, authSkip, middleware, rateLimitRPS, rateLimitRPM, rateLimitKey := parseGorpOptions(methodBody)

			methods = append(methods, ProtoMethod{
				Name:         methodName,
				RequestType:  requestType,
				ResponseType: responseType,
				InputStream:  inputStream,
				OutputStream: outputStream,
				HTTPMethod:   httpMethod,
				HTTPPath:     httpPath,
				HTTPBody:     httpBody,
				AuthRequired: authRequired,
				AuthRoles:    authRoles,
				AuthSkip:     authSkip,
				Middleware:   middleware,
				RateLimitRPS: rateLimitRPS,
				RateLimitRPM: rateLimitRPM,
				RateLimitKey: rateLimitKey,
			})
		}

		services = append(services, ProtoService{
			Name:    serviceName,
			Methods: methods,
		})
	}

	return services, nil
}

// parseHTTPAnnotation parses google.api.http annotation from rpc method body.
// Returns HTTP method, path, and body field name.
//
// parseHTTPAnnotation 从 rpc 方法体解析 google.api.http 注解。
// 返回 HTTP 方法、路径和 body 字段名。
func parseHTTPAnnotation(methodBody string) (httpMethod, httpPath, httpBody string) {
	if methodBody == "" {
		return "", "", ""
	}

	// 匹配 google.api.http 注解块
	// option (google.api.http) = { get: "/v1/users/{id}" };
	httpPattern := regexp.MustCompile(`\(google\.api\.http\)\s*=\s*\{([^}]+)\}`)
	httpMatch := httpPattern.FindStringSubmatch(methodBody)
	if len(httpMatch) < 2 {
		return "", "", ""
	}

	httpBlock := httpMatch[1]

	// 解析 HTTP 方法：get/post/put/delete/patch
	// 格式：get: "/v1/users/{id}" 或 post: "/v1/users" body: "*"
	methodPatterns := []struct {
		method string
		regex  *regexp.Regexp
	}{
		{"GET", regexp.MustCompile(`get:\s*"([^"]+)"`)},
		{"POST", regexp.MustCompile(`post:\s*"([^"]+)"`)},
		{"PUT", regexp.MustCompile(`put:\s*"([^"]+)"`)},
		{"DELETE", regexp.MustCompile(`delete:\s*"([^"]+)"`)},
		{"PATCH", regexp.MustCompile(`patch:\s*"([^"]+)"`)},
	}

	for _, mp := range methodPatterns {
		if match := mp.regex.FindStringSubmatch(httpBlock); len(match) >= 2 {
			httpMethod = mp.method
			httpPath = match[1]
			break
		}
	}

	// 解析 body 字段：body: "*" 或 body: "user"
	bodyPattern := regexp.MustCompile(`body:\s*"([^"]+)"`)
	if match := bodyPattern.FindStringSubmatch(httpBlock); len(match) >= 2 {
		httpBody = match[1]
	}

	return httpMethod, httpPath, httpBody
}

// parseGorpOptions parses gorp custom options from rpc method body.
// Returns auth and middleware configuration.
//
// parseGorpOptions 从 rpc 方法体解析 gorp 自定义选项。
// 返回认证和中间件配置。
func parseGorpOptions(methodBody string) (authRequired bool, authRoles []string, authSkip bool, middleware []string, rateLimitRPS, rateLimitRPM int32, rateLimitKey string) {
	if methodBody == "" {
		return false, nil, false, nil, 0, 0, ""
	}

	// 解析 (gorp.auth) 选项
	// option (gorp.auth) = { required: true, roles: ["admin", "user"] };
	authPattern := regexp.MustCompile(`\(gorp\.auth\)\s*=\s*\{([^}]+)\}`)
	if authMatch := authPattern.FindStringSubmatch(methodBody); len(authMatch) >= 2 {
		authBlock := authMatch[1]

		// 解析 required: true/false
		if match := regexp.MustCompile(`required:\s*(true|false)`).FindStringSubmatch(authBlock); len(match) >= 2 {
			authRequired = match[1] == "true"
		}

		// 解析 roles: ["admin", "user"]
		rolesPattern := regexp.MustCompile(`roles:\s*\[([^\]]*)\]`)
		if match := rolesPattern.FindStringSubmatch(authBlock); len(match) >= 2 {
			roleStrs := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(match[1], -1)
			for _, r := range roleStrs {
				authRoles = append(authRoles, r[1])
			}
		}

		// 解析 skip: true/false
		if match := regexp.MustCompile(`skip:\s*(true|false)`).FindStringSubmatch(authBlock); len(match) >= 2 {
			authSkip = match[1] == "true"
		}
	}

	// 解析 middleware 选项
	// option (gorp.middleware) = "auth", "logging";
	// 或 option (gorp.middleware) = ["auth", "logging"];
	middlewarePattern := regexp.MustCompile(`\(gorp\.middleware\)\s*=\s*(?:"([^"]+)"|\[([^\]]*)\])`)
	if match := middlewarePattern.FindStringSubmatch(methodBody); len(match) >= 2 {
		if match[1] != "" {
			// 单个中间件："auth"
			middleware = []string{match[1]}
		} else if match[2] != "" {
			// 多个中间件：["auth", "logging"]
			mwStrs := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(match[2], -1)
			for _, mw := range mwStrs {
				middleware = append(middleware, mw[1])
			}
		}
	}

	// 解析 rate_limit 选项
	// option (gorp.rate_limit) = { requests_per_second: 100 };
	rateLimitPattern := regexp.MustCompile(`\(gorp\.rate_limit\)\s*=\s*\{([^}]+)\}`)
	if match := rateLimitPattern.FindStringSubmatch(methodBody); len(match) >= 2 {
		rateBlock := match[1]

		// 解析 requests_per_second
		if m := regexp.MustCompile(`requests_per_second:\s*(\d+)`).FindStringSubmatch(rateBlock); len(m) >= 2 {
			rateLimitRPS = parseInt32(m[1])
		}

		// 解析 requests_per_minute
		if m := regexp.MustCompile(`requests_per_minute:\s*(\d+)`).FindStringSubmatch(rateBlock); len(m) >= 2 {
			rateLimitRPM = parseInt32(m[1])
		}

		// 解析 key
		if m := regexp.MustCompile(`key:\s*"([^"]+)"`).FindStringSubmatch(rateBlock); len(m) >= 2 {
			rateLimitKey = m[1]
		}
	}

	return authRequired, authRoles, authSkip, middleware, rateLimitRPS, rateLimitRPM, rateLimitKey
}

// parseInt32 parses a string to int32.
// Uses strconv.ParseInt for proper validation: rejects non-numeric characters
// and reports overflow instead of silently truncating.
//
// parseInt32 将字符串解析为 int32。
// 使用 strconv.ParseInt 进行正确校验：拒绝非数字字符并报告溢出。
func parseInt32(s string) int32 {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}

// generateClientWrapperCode generates Go code for typed RPC client wrapper.
//
// generateClientWrapperCode 生成类型化 RPC 客户端 wrapper 的 Go 代码。
func generateClientWrapperCode(services []ProtoService, opts integrationcontract.ClientGenOptions) string {
	var buf strings.Builder

	// File header
	// 文件头部
	buf.WriteString("// Code generated by gorp proto gen-client. DO NOT EDIT.\n")
	buf.WriteString("//\n")
	buf.WriteString("// 生成代码由 gorp proto gen-client 生成。请勿修改。\n")
	buf.WriteString("\n")
	buf.WriteString("package " + opts.PackageName + "\n\n")

	// Imports
	// 导入
	buf.WriteString("import (\n")
	buf.WriteString("\t\"context\"\n")
	buf.WriteString("\n")
	buf.WriteString("\ttransportcontract \"github.com/ngq/gorp/framework/contract/transport\"\n")
	buf.WriteString(")\n\n")

	// Generate pb import path based on proto file location
	// 根据 proto 文件位置生成 pb 导入路径
	protoDir := filepath.Dir(opts.ProtoFile)
	protoBase := filepath.Base(opts.ProtoFile)
	protoPkg := strings.TrimSuffix(protoBase, ".proto")

	// For pb import, we assume standard protoc output structure
	// 对于 pb 导入，假设使用标准 protoc 输出结构
	buf.WriteString("// Import generated protobuf types.\n")
	buf.WriteString("// 导入生成的 protobuf 类型。\n")
	buf.WriteString("//\n")
	buf.WriteString("// You may need to adjust the import path based on your protoc configuration.\n")
	buf.WriteString("// 可能需要根据 protoc 配置调整导入路径。\n")
	buf.WriteString("//\n")
	buf.WriteString("// Example: import pb \"" + protoDir + "/" + protoPkg + "\"\n")
	buf.WriteString("// 示例: import pb \"" + protoDir + "/" + protoPkg + "\"\n\n")

	// Generate client for each service
	// 为每个服务生成客户端
	for _, svc := range services {
		clientName := opts.ClientPrefix
		if clientName == "" {
			// Remove "Service" suffix if present
			// 如果存在 "Service" 后缀则去掉
			clientName = strings.TrimSuffix(svc.Name, "Service")
			if clientName == "" {
				clientName = svc.Name
			}
		}
		clientName += "Client"

		// Client struct
		// 客户端结构
		buf.WriteString("// " + clientName + " provides typed RPC client methods for " + svc.Name + ".\n")
		buf.WriteString("// Uses gorp's RPCClient with built-in governance capabilities.\n")
		buf.WriteString("//\n")
		buf.WriteString("// " + clientName + " 提供 " + svc.Name + " 的类型化 RPC 客户端方法。\n")
		buf.WriteString("// 使用 gorp 的 RPCClient，内置治理能力。\n")
		buf.WriteString("type " + clientName + " struct {\n")
		buf.WriteString("\tclient transportcontract.RPCClient\n")
		buf.WriteString("\tserviceName string\n")
		buf.WriteString("}\n\n")

		// Constructor
		// 构造函数
		buf.WriteString("// New" + clientName + " creates a new typed client wrapper.\n")
		buf.WriteString("// The RPCClient should be obtained from gorp container.\n")
		buf.WriteString("//\n")
		buf.WriteString("// New" + clientName + " 创建新的类型化客户端 wrapper。\n")
		buf.WriteString("// RPCClient 应从 gorp 容器获取。\n")
		buf.WriteString("func New" + clientName + "(client transportcontract.RPCClient) *" + clientName + " {\n")
		buf.WriteString("\treturn &" + clientName + "{\n")
		buf.WriteString("\t\tclient:      client,\n")
		buf.WriteString("\t\tserviceName: \"" + svc.Name + "\",\n")
		buf.WriteString("\t}\n")
		buf.WriteString("}\n\n")

		// Generate methods
		// 生成方法
		for _, method := range svc.Methods {
			// Skip streaming methods as they require special handling
			// 跳过流式方法，因为需要特殊处理
			if method.InputStream || method.OutputStream {
				buf.WriteString("// " + method.Name + " is a streaming method and requires special handling.\n")
				buf.WriteString("// Please use the raw RPCClient.Conn() for streaming RPC.\n")
				buf.WriteString("//\n")
				buf.WriteString("// " + method.Name + " 是流式方法，需要特殊处理。\n")
				buf.WriteString("// 请使用原始 RPCClient.Conn() 进行流式 RPC。\n\n")
				continue
			}

			buf.WriteString("// " + method.Name + " calls the " + method.Name + " RPC method.\n")
			buf.WriteString("// Automatically applies governance middleware (timeout, retry, etc).\n")
			buf.WriteString("//\n")
			buf.WriteString("// " + method.Name + " 调用 " + method.Name + " RPC 方法。\n")
			buf.WriteString("// 自动应用治理中间件（超时、重试等）。\n")
			buf.WriteString("func (c *" + clientName + ") " + method.Name + "(ctx context.Context, req *" + method.RequestType + ") (*" + method.ResponseType + ", error) {\n")
			buf.WriteString("\tresp := &" + method.ResponseType + "{}\n")
			buf.WriteString("\terr := c.client.Call(ctx, c.serviceName, \"" + method.Name + "\", req, resp)\n")
			buf.WriteString("\treturn resp, err\n")
			buf.WriteString("}\n\n")
		}
	}

	// Governance integration comment if requested
	// 如果请求则添加治理集成注释
	if opts.UseGovernance {
		buf.WriteString("// Governance Integration\n")
		buf.WriteString("// 治理集成\n")
		buf.WriteString("//\n")
		buf.WriteString("// The generated clients automatically benefit from gorp's governance capabilities:\n")
		buf.WriteString("// 生成的客户端自动享有 gorp 的治理能力：\n")
		buf.WriteString("//\n")
		buf.WriteString("// - Timeout: Configurable per-request timeout via RPCClient middleware.\n")
		buf.WriteString("// - Retry: Automatic retry on transient failures.\n")
		buf.WriteString("// - Circuit Breaker: Protection against cascading failures.\n")
		buf.WriteString("// - Tracing: Distributed tracing with OpenTelemetry.\n")
		buf.WriteString("// - Metadata: Automatic metadata propagation between services.\n")
		buf.WriteString("// - Service Auth: Token-based service-to-service authentication.\n")
		buf.WriteString("//\n")
		buf.WriteString("// Configure governance in your gorp application config:\n")
		buf.WriteString("// 在 gorp 应用配置中配置治理：\n")
		buf.WriteString("//\n")
		buf.WriteString("// governance:\n")
		buf.WriteString("//   mode: microservice\n")
		buf.WriteString("//   providers:\n")
		buf.WriteString("//     tracing: otel\n")
		buf.WriteString("//     circuitbreaker: sentinel\n")
		buf.WriteString("//\n")
	}

	return buf.String()
}
