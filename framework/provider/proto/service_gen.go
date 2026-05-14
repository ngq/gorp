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
			if err := genHTTPHandler(svc, svcLower, protoPkg, opts); err != nil {
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

// genHTTPHandler 生成 HTTP handler skeleton。
func genHTTPHandler(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) error {
	handlerDir := filepath.Join(opts.OutputDir, "handler")
	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		return err
	}

	code := generateHTTPHandlerCode(svc, svcLower, protoPkg, opts)
	handlerFile := filepath.Join(handlerDir, svcLower+".go")
	return os.WriteFile(handlerFile, []byte(code), 0644)
}

// generateHTTPHandlerCode 生成 HTTP handler 代码。
func generateHTTPHandlerCode(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) string {
	var buf strings.Builder

	buf.WriteString("// Package handler provides HTTP handlers for ")
	buf.WriteString(svc.Name)
	buf.WriteString(" service.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Package handler 提供 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 服务的 HTTP 处理器。\n")
	buf.WriteString("package handler\n\n")

	buf.WriteString("import (\n")
	buf.WriteString("\t\"net/http\"\n\n")
	if protoPkg != "" {
		buf.WriteString("\tpb \"")
		buf.WriteString(protoPkg)
		buf.WriteString("\"\n\n")
	}
	buf.WriteString("\t\"github.com/gin-gonic/gin\"\n")
	buf.WriteString(")\n\n")

	// Handler struct。
	buf.WriteString("// ")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler handles HTTP requests for ")
	buf.WriteString(svc.Name)
	buf.WriteString(" service.\n")
	buf.WriteString("//\n")
	buf.WriteString("// ")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler 处理 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 服务的 HTTP 请求。\n")
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
	buf.WriteString("Handler creates a new handler.\n")
	buf.WriteString("//\n")
	buf.WriteString("// New")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler 创建新的 handler。\n")
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

	// 为每个方法生成 handler。
	for _, m := range svc.Methods {
		buf.WriteString("// ")
		buf.WriteString(m.Name)
		buf.WriteString(" handles ")
		buf.WriteString(m.Name)
		buf.WriteString(" HTTP request.\n")
		buf.WriteString("//\n")
		buf.WriteString("// ")
		buf.WriteString(m.Name)
		buf.WriteString(" 处理 ")
		buf.WriteString(m.Name)
		buf.WriteString(" HTTP 请求。\n")
		buf.WriteString("func (h *")
		buf.WriteString(svc.Name)
		buf.WriteString("Handler) ")
		buf.WriteString(m.Name)
		buf.WriteString("(c *gin.Context) {\n")
		if protoPkg != "" {
			buf.WriteString("\tvar req pb.")
			buf.WriteString(m.RequestType)
			buf.WriteString("\n")
			buf.WriteString("\tif err := c.ShouldBindJSON(&req); err != nil {\n")
			buf.WriteString("\t\tc.JSON(http.StatusBadRequest, gin.H{\"error\": err.Error()})\n")
			buf.WriteString("\t\treturn\n")
			buf.WriteString("\t}\n\n")
			buf.WriteString("\t// TODO: implement business logic\n")
			buf.WriteString("\t// resp, err := h.svc.")
			buf.WriteString(m.Name)
			buf.WriteString("(c.Request.Context(), &req)\n")
			buf.WriteString("\t// if err != nil {\n")
			buf.WriteString("\t// \tc.JSON(http.StatusInternalServerError, gin.H{\"error\": err.Error()})\n")
			buf.WriteString("\t// \treturn\n")
			buf.WriteString("\t// }\n\n")
		}
		buf.WriteString("\tc.JSON(http.StatusOK, gin.H{\"message\": \"TODO: implement ")
		buf.WriteString(m.Name)
		buf.WriteString("\"})\n")
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
		buf.WriteString("\t// TODO: implement business logic\n")
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
func generateRoutesCode(svc ProtoService, svcLower, protoPkg string, opts integrationcontract.ServiceGenOptions) string {
	var buf strings.Builder

	buf.WriteString("// Package routes provides route registration for ")
	buf.WriteString(svc.Name)
	buf.WriteString(" service.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Package routes 提供 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 服务的路由注册。\n")
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
	buf.WriteString(")\n\n")

	// Register 函数。
	buf.WriteString("// Register")
	buf.WriteString(svc.Name)
	buf.WriteString("Routes registers ")
	buf.WriteString(svc.Name)
	buf.WriteString(" routes.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Register")
	buf.WriteString(svc.Name)
	buf.WriteString("Routes 注册 ")
	buf.WriteString(svc.Name)
	buf.WriteString(" 路由。\n")
	buf.WriteString("func Register")
	buf.WriteString(svc.Name)
	buf.WriteString("Routes(r *gin.RouterGroup, h *handler.")
	buf.WriteString(svc.Name)
	buf.WriteString("Handler) {\n")
	buf.WriteString("\tgrp := r.Group(\"/")
	buf.WriteString(svcLower)
	buf.WriteString("\")\n\n")

	// 为每个方法生成路由注册。
	// 基于 HTTP annotation 约定：方法名推导 HTTP 方法和路径。
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

	buf.WriteString("}\n")

	return buf.String()
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
		return "GET", ""
	case strings.HasPrefix(name, "Update"):
		return "PUT", "/" + strings.ToLower(strings.TrimPrefix(name, "Update")) + "/:id"
	case strings.HasPrefix(name, "Delete"):
		return "DELETE", "/" + strings.ToLower(strings.TrimPrefix(name, "Delete")) + "/:id"
	default:
		return "POST", "/" + strings.ToLower(name)
	}
}
