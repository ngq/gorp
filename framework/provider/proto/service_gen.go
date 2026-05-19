// Package proto provides proto generator implementation.
// This file implements service skeleton generation from proto files.
// Generates HTTP handler, gRPC service implementation and route registration.
//
// Proto 包提供 proto 生成器实现。
// 本文件实现从 proto 文件生成服务骨架代码。
// 生成 HTTP handler、gRPC service 实现和路由注册。
package proto

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// GenService generates HTTP handler, gRPC service skeleton and route registration from proto file.
// Enables proto-first workflow: proto → service implementation skeleton.
//
// GenService 从 proto 文件生成 HTTP handler、gRPC service skeleton 和路由注册。
// 支持闭环 proto-first 工作流：proto → 服务实现骨架。
func (g *Generator) GenService(ctx context.Context, opts integrationcontract.ServiceGenOptions) error {
	if opts.ProtoFile == "" {
		return errors.New("proto file path is required")
	}
	if opts.OutputDir == "" {
		return errors.New("output directory is required")
	}

	// 解析 proto 文件提取服务定义。
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

	// 如果指定了服务名则过滤。
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
		return fmt.Errorf("no services found in proto file: %s", opts.ProtoFile)
	}

	// 推导 proto package 路径，用于 import。
	protoPkg := deriveProtoPackage(opts.ProtoFile, opts.Module)

	for _, svc := range services {
		svcLower := strings.ToLower(svc.Name)

		if opts.IncludeHTTP {
			if err := genHandler(svc, svcLower, protoPkg, opts); err != nil {
				return fmt.Errorf("generate HTTP handler for %s failed: %w", svc.Name, err)
			}
		}

		if opts.IncludeGRPC {
			if err := genGRPCService(svc, svcLower, protoPkg, opts); err != nil {
				return fmt.Errorf("generate gRPC service for %s failed: %w", svc.Name, err)
			}
		}

		if opts.RegisterRoutes {
			if err := genRoutesRegistration(svc, svcLower, protoPkg, opts); err != nil {
				return fmt.Errorf("generate routes for %s failed: %w", svc.Name, err)
			}
		}
	}

	return nil
}

// deriveProtoPackage 从 proto 文件路径和 module 推导 Go import 路径。
// 优先从 proto 文件的 go_package 选项解析，其次从路径推导。
func deriveProtoPackage(protoFile, module string) string {
	// 优先解析 proto 文件中的 go_package 选项。
	if goPkg, ok := parseGoPackageFromProto(protoFile); ok {
		return goPkg
	}

	// 从路径推导：api/proto/user/v1/user.proto → module/proto/user/v1;userv1
	if module == "" {
		return ""
	}

	// 规范化路径分隔符。
	protoFile = filepath.ToSlash(protoFile)

	// 查找 proto 目录起始位置。
	for _, prefix := range []string{"api/proto/", "proto/"} {
		idx := strings.Index(protoFile, prefix)
		if idx >= 0 {
			rel := protoFile[idx+len(prefix):]
			// 去掉文件名，保留目录。
			dir := filepath.Dir(rel)
			dir = filepath.ToSlash(dir)
			parts := strings.Split(dir, "/")
			last := parts[len(parts)-1]
			// 生成 go_package 格式：module/proto/user/v1;userv1
			return module + "/proto/" + dir + ";" + last
		}
	}

	return ""
}

// parseGoPackageFromProto 从 proto 文件解析 go_package 选项。
// 返回的 import 路径不包含分号后的 package 别名。
func parseGoPackageFromProto(protoFile string) (string, bool) {
	content, err := os.ReadFile(protoFile)
	if err != nil {
		return "", false
	}

	// 匹配 go_package 选项：option go_package = "...";
	re := regexp.MustCompile(`option\s+go_package\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) >= 2 {
		goPkg := matches[1]
		// 去掉分号后的 package 别名（如 "example.com/proto/user/v1;userv1" → "example.com/proto/user/v1"）。
		if idx := strings.Index(goPkg, ";"); idx >= 0 {
			goPkg = goPkg[:idx]
		}
		return goPkg, true
	}
	return "", false
}

// genHandler 生成 handler skeleton。
func genHandler(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) error {
	handlerDir := filepath.Join(opts.OutputDir, "handler")
	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		return err
	}

	code := generateHandlerCode(svc, svcLower, protoPkg, opts)
	handlerFile := filepath.Join(handlerDir, svcLower+".go")
	return os.WriteFile(handlerFile, []byte(code), 0644)
}

// generateHandlerCode 生成 handler 代码。
// 生成的 handler 自动绑定请求参数、调用 ServiceServer、处理错误码映射。
// 用户只需实现 gRPC ServiceServer 接口，HTTP 层自动委托。
func generateHandlerCode(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) string {
	var buf strings.Builder

	buf.WriteString("// Package handler provides HTTP handlers for ")
	buf.WriteString(svc.Name)
	buf.WriteString(" service.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Package handler 提供 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 服务的 HTTP 处理器。\n")
	buf.WriteString("// 生成的 handler 自动委托调用 gRPC ServiceServer，实现双协议支持。\n")
	buf.WriteString("package handler\n\n")

	// Import 块。
	buf.WriteString("import (\n")
	buf.WriteString("\t\"net/http\"\n\n")
	if protoPkg != "" {
		buf.WriteString("\tpb \"")
		buf.WriteString(protoPkg)
		buf.WriteString("\"\n\n")
	}
	buf.WriteString("\t\"github.com/gin-gonic/gin\"\n")
	buf.WriteString("\t\"google.golang.org/grpc/codes\"\n")
	buf.WriteString("\t\"google.golang.org/grpc/status\"\n")
	// 如果启用校验，添加校验相关导入。
	if opts.IncludeValidation {
		buf.WriteString("\n\t\"github.com/ngq/gorp/validate\"\n")
	}
	buf.WriteString(")\n\n")

	// Handler struct。
	buf.WriteString("// ")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler handles HTTP requests for ")
	buf.WriteString(svc.Name)
	buf.WriteString(" service.\n")
	buf.WriteString("// 自动委托调用 gRPC ServiceServer，实现一套业务逻辑、双协议暴露。\n")
	buf.WriteString("//\n")
	buf.WriteString("// ")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler 处理 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 服务的 HTTP 请求。\n")
	buf.WriteString("// 自动委托调用 gRPC ServiceServer，实现一套业务逻辑、双协议暴露。\n")
	buf.WriteString("type ")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler struct {\n")
	buf.WriteString("\tsvc ")
	if protoPkg != "" {
		buf.WriteString("pb.")
	} else {
		buf.WriteString(svc.Name)
	}
	buf.WriteString(svc.Name)
	buf.WriteString("ServiceServer\n")
	buf.WriteString("}\n\n")

	// Constructor。
	buf.WriteString("// New")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler creates a new handler with gRPC service implementation.\n")
	buf.WriteString("//\n")
	buf.WriteString("// New")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler 使用 gRPC 服务实现创建新的 handler。\n")
	buf.WriteString("func New")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler(svc ")
	if protoPkg != "" {
		buf.WriteString("pb.")
	} else {
		buf.WriteString(svc.Name)
	}
	buf.WriteString(svc.Name)
	buf.WriteString("ServiceServer) *")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler {\n")
	buf.WriteString("\treturn &")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler{svc: svc}\n")
	buf.WriteString("}\n\n")

	// 生成 gRPC status → HTTP status 映射辅助函数。
	buf.WriteString("// grpcCodeToHTTPStatus 将 gRPC status code 映射为 HTTP status code。\n")
	buf.WriteString("//\n")
	buf.WriteString("// grpcCodeToHTTPStatus maps gRPC status code to HTTP status code.\n")
	buf.WriteString("func grpcCodeToHTTPStatus(code codes.Code) int {\n")
	buf.WriteString("\tswitch code {\n")
	buf.WriteString("\tcase codes.OK:\n")
	buf.WriteString("\t\treturn http.StatusOK\n")
	buf.WriteString("\tcase codes.InvalidArgument:\n")
	buf.WriteString("\t\treturn http.StatusBadRequest\n")
	buf.WriteString("\tcase codes.NotFound:\n")
	buf.WriteString("\t\treturn http.StatusNotFound\n")
	buf.WriteString("\tcase codes.AlreadyExists:\n")
	buf.WriteString("\t\treturn http.StatusConflict\n")
	buf.WriteString("\tcase codes.PermissionDenied:\n")
	buf.WriteString("\t\treturn http.StatusForbidden\n")
	buf.WriteString("\tcase codes.Unauthenticated:\n")
	buf.WriteString("\t\treturn http.StatusUnauthorized\n")
	buf.WriteString("\tcase codes.Unavailable:\n")
	buf.WriteString("\t\treturn http.StatusServiceUnavailable\n")
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn http.StatusInternalServerError\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	// 为每个方法生成 handler。
	for _, m := range svc.Methods {
		// 跳过流式方法，流式方法需要特殊处理。
		if m.InputStream || m.OutputStream {
			buf.WriteString("// ")
			buf.WriteString(m.Name)
			buf.WriteString(" is a streaming method and requires custom implementation.\n")
			buf.WriteString("//\n")
			buf.WriteString("// ")
			buf.WriteString(m.Name)
			buf.WriteString(" 是流式方法，需要自定义实现。\n")
			buf.WriteString("func (h *")
			buf.WriteString(svc.Name)
			buf.WriteString("Handler) ")
			buf.WriteString(m.Name)
			buf.WriteString("(c *gin.Context) {\n")
			buf.WriteString("\tc.JSON(http.StatusNotImplemented, gin.H{\"error\": \"streaming method not supported in HTTP\"})\n")
			buf.WriteString("}\n\n")
			continue
		}

		buf.WriteString("// ")
		buf.WriteString(m.Name)
		buf.WriteString(" handles ")
		buf.WriteString(m.Name)
		buf.WriteString(" HTTP request.\n")
		buf.WriteString("// 自动绑定请求、调用 ServiceServer、处理错误码映射。\n")
		buf.WriteString("//\n")
		buf.WriteString("// ")
		buf.WriteString(m.Name)
		buf.WriteString(" 处理 ")
		buf.WriteString(m.Name)
		buf.WriteString(" HTTP 请求。\n")
		buf.WriteString("// 自动绑定请求、调用 ServiceServer、处理错误码映射。\n")
		buf.WriteString("func (h *")
		buf.WriteString(svc.Name)
		buf.WriteString("Handler) ")
		buf.WriteString(m.Name)
		buf.WriteString("(c *gin.Context) {\n")

		if protoPkg != "" {
			// 请求绑定。
			buf.WriteString("\tvar req pb.")
			buf.WriteString(m.RequestType)
			buf.WriteString("\n")
			buf.WriteString("\tif err := c.ShouldBindJSON(&req); err != nil {\n")
			buf.WriteString("\t\tc.JSON(http.StatusBadRequest, gin.H{\"error\": err.Error()})\n")
			buf.WriteString("\t\treturn\n")
			buf.WriteString("\t}\n\n")

			// 调用 ServiceServer。
			// 如果启用校验，添加校验调用。
			if opts.IncludeValidation {
				buf.WriteString("\t// 请求校验。\n")
				buf.WriteString("\tif v, ok := req.(interface{ Validate() error }); ok {\n")
				buf.WriteString("\t\tif err := v.Validate(); err != nil {\n")
				buf.WriteString("\t\t\tc.JSON(http.StatusBadRequest, gin.H{\"error\": err.Error()})\n")
				buf.WriteString("\t\t\treturn\n")
				buf.WriteString("\t\t}\n")
				buf.WriteString("\t}\n\n")
			}
			buf.WriteString("\t// 调用 gRPC ServiceServer，业务逻辑在 service 层实现。\n")
			buf.WriteString("\tresp, err := h.svc.")
			buf.WriteString(m.Name)
			buf.WriteString("(c.Request.Context(), &req)\n")
			buf.WriteString("\tif err != nil {\n")
			buf.WriteString("\t\t// 将 gRPC status 映射为 HTTP status。\n")
			buf.WriteString("\t\tif st, ok := status.FromError(err); ok {\n")
			buf.WriteString("\t\t\thttpStatus := grpcCodeToHTTPStatus(st.Code())\n")
			buf.WriteString("\t\t\tc.JSON(httpStatus, gin.H{\"error\": st.Message()})\n")
			buf.WriteString("\t\t\treturn\n")
			buf.WriteString("\t\t}\n")
			buf.WriteString("\t\tc.JSON(http.StatusInternalServerError, gin.H{\"error\": err.Error()})\n")
			buf.WriteString("\t\treturn\n")
			buf.WriteString("\t}\n\n")

			// 返回响应。
			buf.WriteString("\tc.JSON(http.StatusOK, resp)\n")
		} else {
			buf.WriteString("\tc.JSON(http.StatusOK, gin.H{\"message\": \"TODO: implement ")
			buf.WriteString(m.Name)
			buf.WriteString("\"})\n")
		}
		buf.WriteString("}\n\n")
	}

	return buf.String()
}

// genGRPCService 生成 gRPC service skeleton。
func genGRPCService(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) error {
	serviceDir := filepath.Join(opts.OutputDir, "service")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	code := generateGRPCServiceCode(svc, svcLower, protoPkg, opts)
	serviceFile := filepath.Join(serviceDir, svcLower+".go")
	return os.WriteFile(serviceFile, []byte(code), 0644)
}

// generateGRPCServiceCode 生成 gRPC service 代码。
func generateGRPCServiceCode(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) string {
	var buf strings.Builder

	buf.WriteString("// Package service provides gRPC service implementation for ")
	buf.WriteString(svc.Name)
	buf.WriteString(".\n")
	buf.WriteString("//\n")
	buf.WriteString("// Package service 提供 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 的 gRPC 服务实现。\n")
	buf.WriteString("package service\n\n")

	buf.WriteString("import (\n")
	buf.WriteString("\t\"context\"\n\n")
	if protoPkg != "" {
		buf.WriteString("\tpb \"")
		buf.WriteString(protoPkg)
		buf.WriteString("\"\n\n")
	}
	buf.WriteString(")\n\n")

	// Service struct。
	buf.WriteString("// ")
	buf.WriteString(svc.Name)
	buf.WriteString("Service implements pb.")
	buf.WriteString(svc.Name)
	buf.WriteString("ServiceServer.\n")
	buf.WriteString("//\n")
	buf.WriteString("// ")
	buf.WriteString(svc.Name)
	buf.WriteString("Service 实现 pb.")
	buf.WriteString(svc.Name)
	buf.WriteString("ServiceServer。\n")
	buf.WriteString("type ")
	buf.WriteString(svc.Name)
	buf.WriteString("Service struct {\n")
	if protoPkg != "" {
		buf.WriteString("\tpb.Unimplemented")
		buf.WriteString(svc.Name)
		buf.WriteString("ServiceServer\n")
	}
	buf.WriteString("}\n\n")

	// Constructor。
	buf.WriteString("// New")
	buf.WriteString(svc.Name)
	buf.WriteString("Service creates a new service instance.\n")
	buf.WriteString("//\n")
	buf.WriteString("// New")
	buf.WriteString(svc.Name)
	buf.WriteString("Service 创建新的服务实例。\n")
	buf.WriteString("func New")
	buf.WriteString(svc.Name)
	buf.WriteString("Service() *")
	buf.WriteString(svc.Name)
	buf.WriteString("Service {\n")
	buf.WriteString("\treturn &")
	buf.WriteString(svc.Name)
	buf.WriteString("Service{}\n")
	buf.WriteString("}\n\n")

	// 为每个方法生成实现。
	for _, m := range svc.Methods {
		buf.WriteString("// ")
		buf.WriteString(m.Name)
		buf.WriteString(" implements ")
		buf.WriteString(m.Name)
		buf.WriteString(" gRPC method.\n")
		buf.WriteString("//\n")
		buf.WriteString("// ")
		buf.WriteString(m.Name)
		buf.WriteString(" 实现 ")
		buf.WriteString(m.Name)
		buf.WriteString(" gRPC 方法。\n")
		buf.WriteString("func (s *")
		buf.WriteString(svc.Name)
		buf.WriteString("Service) ")
		buf.WriteString(m.Name)
		buf.WriteString("(ctx context.Context, req *")
		if protoPkg != "" {
			buf.WriteString("pb.")
		}
		buf.WriteString(m.RequestType)
		buf.WriteString(") (*")
		if protoPkg != "" {
			buf.WriteString("pb.")
		}
		buf.WriteString(m.ResponseType)
		buf.WriteString(", error) {\n")
		buf.WriteString("\treturn &")
		if protoPkg != "" {
			buf.WriteString("pb.")
		}
		buf.WriteString(m.ResponseType)
		buf.WriteString("{}, nil\n")
		buf.WriteString("}\n\n")
	}

	return buf.String()
}

// genRoutesRegistration 生成路由注册代码。
func genRoutesRegistration(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) error {
	routesDir := filepath.Join(opts.OutputDir, "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return err
	}

	code := generateRoutesCode(svc, svcLower, protoPkg, opts)
	routesFile := filepath.Join(routesDir, svcLower+".go")
	return os.WriteFile(routesFile, []byte(code), 0644)
}

// generateRoutesCode 生成路由注册代码。
// 优先使用 proto annotation 中定义的路由，fallback 到方法名约定。
// 支持 gorp 自定义选项：认证、中间件、限流。
// 当方法声明了 gorp.auth 或 gorp.middleware 时，生成代码自动从
// MiddlewareRegistry 查找并挂载对应中间件，无需手动分组。
func generateRoutesCode(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) string {
	var buf strings.Builder

	// 检查是否有任何方法需要认证或自定义中间件。
	hasAuthRequired := false
	hasMiddleware := false
	for _, m := range svc.Methods {
		if m.AuthRequired || len(m.AuthRoles) > 0 {
			hasAuthRequired = true
		}
		if len(m.Middleware) > 0 {
			hasMiddleware = true
		}
	}
	// 需要中间件注册表来支持自动挂载。
	needRegistry := hasAuthRequired || hasMiddleware

	buf.WriteString("// Package routes provides route registration for ")
	buf.WriteString(svc.Name)
	buf.WriteString(" service.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Package routes 提供 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 服务的路由注册。\n")
	buf.WriteString("// 路由优先从 proto annotation 读取，fallback 到方法名约定。\n")
	buf.WriteString("// 支持 gorp 自定义选项：认证、中间件、限流。\n")
	if needRegistry {
		buf.WriteString("// 当 proto 中声明了 gorp.auth 或 gorp.middleware 时，\n")
		buf.WriteString("// 生成的代码会自动从 MiddlewareRegistry 查找并挂载对应中间件。\n")
	}
	buf.WriteString("package routes\n\n")

	buf.WriteString("import (\n")
	buf.WriteString("\t\"github.com/gin-gonic/gin\"\n\n")

	// handler import。
	handlerPkg := opts.Module
	if handlerPkg != "" {
		// 推导 handler 包路径。
		handlerRel := strings.TrimPrefix(opts.OutputDir, "/")
		buf.WriteString("\t\"")
		buf.WriteString(handlerPkg)
		buf.WriteString("/")
		buf.WriteString(handlerRel)
		buf.WriteString("/handler\"\n")
	}

	// 如果需要中间件注册表，添加 import。
	if needRegistry {
		buf.WriteString("\n\ttransportcontract \"github.com/ngq/gorp/framework/contract/transport\"\n")
	}
	buf.WriteString(")\n\n")

	// Register 函数签名。
	buf.WriteString("// Register")
	buf.WriteString(svc.Name)
	buf.WriteString("Routes registers ")
	buf.WriteString(svc.Name)
	buf.WriteString(" routes.\n")
	buf.WriteString("// 路由定义来源：proto annotation > 方法名约定。\n")
	buf.WriteString("// 认证/中间件配置来源：gorp 自定义选项。\n")
	buf.WriteString("//\n")
	buf.WriteString("// Register")
	buf.WriteString(svc.Name)
	buf.WriteString("Routes 注册 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 路由。\n")
	buf.WriteString("// 路由定义来源：proto annotation > 方法名约定。\n")
	buf.WriteString("// 认证/中间件配置来源：gorp 自定义选项。\n")

	// 函数签名：如果有中间件需求，增加 registry 参数。
	if needRegistry {
		buf.WriteString("func Register")
		buf.WriteString(svc.Name)
		buf.WriteString("Routes(r *gin.RouterGroup, h *handler.")
		buf.WriteString(svc.Name)
		buf.WriteString("Handler, registry transportcontract.MiddlewareRegistry) {\n")
	} else {
		buf.WriteString("func Register")
		buf.WriteString(svc.Name)
		buf.WriteString("Routes(r *gin.RouterGroup, h *handler.")
		buf.WriteString(svc.Name)
		buf.WriteString("Handler) {\n")
	}

	// 检查是否有任何方法定义了 HTTP annotation。
	hasAnnotations := false
	for _, m := range svc.Methods {
		if m.HTTPMethod != "" && m.HTTPPath != "" {
			hasAnnotations = true
			break
		}
	}

	if needRegistry {
		// 生成中间件查找代码，从注册表获取中间件实例。
		buf.WriteString(generateMiddlewareLookupCode(svc))
	}

	if hasAnnotations {
		// 使用 annotation 定义的路由，直接注册到根路由。
		buf.WriteString("\t// 路由从 proto annotation 解析。\n")
		for _, m := range svc.Methods {
			if m.HTTPMethod == "" || m.HTTPPath == "" {
				continue
			}
			// 将 proto 路径参数 {id} 转换为 Gin 路径参数 :id。
			ginPath := protoPathToGinPath(m.HTTPPath)

			// 生成路由注册，带中间件自动挂载。
			buf.WriteString(generateRouteWithMiddlewareAutoMount(m, ginPath, needRegistry))
		}
	} else {
		// Fallback 到方法名约定。
		buf.WriteString("\tgrp := r.Group(\"/")
		buf.WriteString(svcLower)
		buf.WriteString("\")\n\n")
		buf.WriteString("\t// 路由从方法名约定推导（proto 中未定义 annotation）。\n")
		for _, m := range svc.Methods {
			method, path := methodToHTTP(m.Name)
			buf.WriteString("\tgrp.")
			buf.WriteString(method)
			buf.WriteString("(\"")
			buf.WriteString(path)
			buf.WriteString("\", h.")
			buf.WriteString(m.Name)
			buf.WriteString(")\n")
		}
	}

	buf.WriteString("}\n")

	return buf.String()
}

// generateMiddlewareLookupCode 生成从中间件注册表查找中间件的代码。
// 预先查找所有需要的中间件，避免每个路由重复查找。
func generateMiddlewareLookupCode(svc ProtoService) string {
	var buf strings.Builder

	buf.WriteString("\t// 从中间件注册表查找 proto 中声明的中间件。\n")
	buf.WriteString("\t// 注册表由应用启动时初始化，通过 registry.Register(\"auth\", mw) 注册。\n")

	// 收集所有需要的中间件名称（去重）。
	authNeeded := false
	middlewareNames := map[string]bool{}
	for _, m := range svc.Methods {
		if m.AuthRequired || len(m.AuthRoles) > 0 {
			authNeeded = true
		}
		for _, mw := range m.Middleware {
			middlewareNames[mw] = true
		}
	}

	// 生成 auth 中间件查找。
	if authNeeded {
		buf.WriteString("\tauthMw, _ := registry.Lookup(\"auth\")\n")
	}

	// 生成其他中间件查找。
	for name := range middlewareNames {
		varName := sanitizeVarName(name) + "Mw"
		buf.WriteString("\t" + varName + ", _ := registry.Lookup(\"" + name + "\")\n")
	}

	buf.WriteString("\n")
	return buf.String()
}

// generateRouteWithMiddlewareAutoMount 生成单个路由注册代码，支持自动挂载中间件。
// 如果方法声明了 auth 或 middleware，生成带中间件的路由组。
func generateRouteWithMiddlewareAutoMount(m ProtoMethod, ginPath string, hasRegistry bool) string {
	var buf strings.Builder

	// 判断是否需要中间件。
	needsAuth := m.AuthRequired || len(m.AuthRoles) > 0
	needsMiddleware := len(m.Middleware) > 0

	if hasRegistry && (needsAuth || needsMiddleware) {
		// 生成带中间件的路由组。
		buf.WriteString("\t// [AUTH")
		if len(m.AuthRoles) > 0 {
			buf.WriteString(" roles:")
			buf.WriteString(strings.Join(m.AuthRoles, ","))
		}
		if len(m.Middleware) > 0 {
			buf.WriteString(" middleware:")
			buf.WriteString(strings.Join(m.Middleware, ","))
		}
		buf.WriteString("]\n")

		// 收集需要挂载的中间件变量名。
		var mwVars []string
		if needsAuth {
			mwVars = append(mwVars, "authMw")
		}
		for _, name := range m.Middleware {
			mwVars = append(mwVars, sanitizeVarName(name)+"Mw")
		}

		// 生成中间件过滤代码（跳过未注册的中间件）。
		buf.WriteString("\t_" + m.Name + "Mws := []gin.HandlerFunc{}\n")
		for _, v := range mwVars {
			buf.WriteString("\tif " + v + " != nil {\n")
			buf.WriteString("\t\t_" + m.Name + "Mws = append(_" + m.Name + "Mws, " + v + ")\n")
			buf.WriteString("\t}\n")
		}

		// 生成路由组注册。
		buf.WriteString("\tr.Group(\"\").Use(_" + m.Name + "Mws...)." + m.HTTPMethod + "(\"" + ginPath + "\", h." + m.Name + ")\n")
	} else {
		// 无中间件，直接注册。
		if m.AuthRequired || len(m.AuthRoles) > 0 || len(m.Middleware) > 0 {
			buf.WriteString("\t// [AUTH")
			if len(m.AuthRoles) > 0 {
				buf.WriteString(" roles:")
				buf.WriteString(strings.Join(m.AuthRoles, ","))
			}
			if len(m.Middleware) > 0 {
				buf.WriteString(" middleware:")
				buf.WriteString(strings.Join(m.Middleware, ","))
			}
			buf.WriteString("]\n")
		}

		buf.WriteString("\tr.")
		buf.WriteString(m.HTTPMethod)
		buf.WriteString("(\"")
		buf.WriteString(ginPath)
		buf.WriteString("\", h.")
		buf.WriteString(m.Name)
		buf.WriteString(")\n")
	}

	return buf.String()
}

// sanitizeVarName 将中间件名称转为合法的 Go 变量名前缀。
// 例如 "rate-limit" → "rateLimit"，"auth" → "auth"。
func sanitizeVarName(name string) string {
	var buf strings.Builder
	nextUpper := false
	for _, c := range name {
		if c == '-' || c == '_' || c == '.' {
			nextUpper = true
			continue
		}
		if nextUpper {
			buf.WriteRune(toUpper(c))
			nextUpper = false
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

// toUpper 将小写字母转为大写。
func toUpper(c rune) rune {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

// protoPathToGinPath 将 proto 路径参数格式转换为 Gin 格式。
// proto: /v1/users/{id} → gin: /v1/users/:id
// proto: /v1/users/{user_id}/posts/{post_id} → gin: /v1/users/:user_id/posts/:post_id
func protoPathToGinPath(protoPath string) string {
	// 使用正则替换 {xxx} 为 :xxx。
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllString(protoPath, ":$1")
}

// methodToHTTP 根据 gRPC 方法名推导 HTTP 方法和路径。
// 约定：CreateXxx → POST /xxx, GetXxx → GET /xxx, ListXxx → GET /xxx/list,
// UpdateXxx → PUT /xxx, DeleteXxx → DELETE /xxx。
func methodToHTTP(name string) (string, string) {
	switch {
	case strings.HasPrefix(name, "Create"):
		return "POST", "/" + strings.ToLower(strings.TrimPrefix(name, "Create"))
	case strings.HasPrefix(name, "Get"):
		return "GET", "/" + strings.ToLower(strings.TrimPrefix(name, "Get")) + "/:id"
	case strings.HasPrefix(name, "List"):
		return "GET", "/" + strings.ToLower(strings.TrimPrefix(name, "List")) + "/list"
	case strings.HasPrefix(name, "Update"):
		return "PUT", "/" + strings.ToLower(strings.TrimPrefix(name, "Update")) + "/:id"
	case strings.HasPrefix(name, "Delete"):
		return "DELETE", "/" + strings.ToLower(strings.TrimPrefix(name, "Delete")) + "/:id"
	default:
		return "POST", "/" + strings.ToLower(name)
	}
}
