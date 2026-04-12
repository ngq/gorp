package proto

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/ngq/gorp/framework/contract"
)

// parseGinRoutes 解析 Gin 路由定义。
//
// 中文说明：
// - 解析 Go 文件的 AST，提取 Gin 路由注册代码；
// - 支持 GET/POST/PUT/DELETE/PATCH 等方法；
// - 支持分组路由（Group）；
// - 自动提取路径参数（:id → {id}）。
func (g *Generator) parseGinRoutes(filePath string) ([]contract.RouteDef, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	var routes []contract.RouteDef
	var currentGroup string

	// 遍历 AST
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// 检查是否是路由注册调用
			route := g.extractRouteFromCall(node, currentGroup)
			if route != nil {
				routes = append(routes, *route)
			}

			// 检查是否是 Group 调用
			group := g.extractGroupFromCall(node)
			if group != "" {
				currentGroup = group
			}
		}
		return true
	})

	return routes, nil
}

// extractRouteFromCall 从调用表达式中提取路由信息。
func (g *Generator) extractRouteFromCall(call *ast.CallExpr, currentGroup string) *contract.RouteDef {
	// 检查是否是方法调用
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	// 检查方法名是否是 HTTP 方法
	method := strings.ToUpper(sel.Sel.Name)
	httpMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true,
		"DELETE": true, "PATCH": true, "OPTIONS": true, "HEAD": true,
	}
	if !httpMethods[method] {
		return nil
	}

	// 检查参数数量
	if len(call.Args) < 2 {
		return nil
	}

	// 提取路径
	path := g.extractStringLiteral(call.Args[0])
	if path == "" {
		return nil
	}

	// 拼接分组路径
	fullPath := path
	if currentGroup != "" {
		fullPath = currentGroup + path
	}

	// 提取处理函数名
	handlerName := g.extractHandlerName(call.Args[1])

	return &contract.RouteDef{
		Method:       method,
		Path:         fullPath,
		HandlerName:  handlerName,
		Comments:     []string{},
		RequestType:  nil, // 后续从处理函数推断
		ResponseType: nil,
	}
}

// extractGroupFromCall 从调用表达式中提取路由分组。
func (g *Generator) extractGroupFromCall(call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	if sel.Sel.Name != "Group" {
		return ""
	}

	if len(call.Args) < 1 {
		return ""
	}

	return g.extractStringLiteral(call.Args[0])
}

// extractStringLiteral 提取字符串字面量。
func (g *Generator) extractStringLiteral(expr ast.Expr) string {
	basicLit, ok := expr.(*ast.BasicLit)
	if !ok {
		return ""
	}
	// 去掉引号
	return strings.Trim(basicLit.Value, `"`)
}

// extractHandlerName 提取处理函数名称。
func (g *Generator) extractHandlerName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.FuncLit:
		return "anonymousHandler"
	}
	return "unknownHandler"
}

// convertPathParams 转换路径参数格式。
//
// 中文说明：
// - Gin 格式：/users/:id
// - Proto 格式：/users/{id}
func (g *Generator) convertPathParams(path string) string {
	// 匹配 :param_name 并转换为 {param_name}
	re := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	return re.ReplaceAllString(path, `{$1}`)
}

// generateProtoFromRoutes 从路由生成 proto。
//
// 中文说明：
// - 根据路由定义生成 proto service；
// - 自动生成请求/响应 message；
// - 添加 google.api.http 注解；
// - 支持 Handler 类型推断（当提供 HandlerFile 时）。
func (g *Generator) generateProtoFromRoutes(routes []contract.RouteDef, opts contract.RouteToProtoOptions) string {
	var buf bytes.Buffer

	// 解析 Handler 类型（如果提供了 HandlerFile）
	handlerTypes := make(map[string]*HandlerTypeInfo)
	if opts.HandlerFile != "" {
		parser := NewHandlerParser()
		handlerNames := make([]string, len(routes))
		for i, route := range routes {
			handlerNames[i] = route.HandlerName
		}
		handlerTypes, _ = parser.ParseHandlers(opts.HandlerFile, handlerNames)
	}

	// 头部
	buf.WriteString("syntax = \"proto3\";\n\n")
	buf.WriteString(fmt.Sprintf("package %s;\n\n", opts.Package))
	buf.WriteString("import \"google/api/annotations.proto\";\n\n")
	buf.WriteString(fmt.Sprintf("option go_package = \"%s\";\n\n", opts.GoPackage))

	// 收集所有消息类型和已生成的消息
	generatedMsgs := make(map[string]bool)

	// 生成消息定义
	buf.WriteString("// Request/Response messages\n\n")
	for _, route := range routes {
		methodName := g.toCamelCase(route.HandlerName)
		reqName := methodName + "Request"
		respName := methodName + "Response"

		// 获取 Handler 类型信息
		handlerInfo := handlerTypes[route.HandlerName]

		// 生成请求消息
		if !generatedMsgs[reqName] {
			g.generateEnhancedRouteMessage(&buf, reqName, route, handlerInfo)
			generatedMsgs[reqName] = true
		}

		// 生成响应消息
		if !generatedMsgs[respName] {
			g.generateEnhancedResponseMessage(&buf, respName, handlerInfo)
			generatedMsgs[respName] = true
		}
	}

	// 生成服务定义
	buf.WriteString(fmt.Sprintf("service %s {\n", opts.ServiceName))
	for _, route := range routes {
		g.generateRouteMethod(&buf, route, opts)
	}
	buf.WriteString("}\n")

	return buf.String()
}

// generateEnhancedRouteMessage 生成增强的路由请求消息。
//
// 中文说明：
// - 包含路径参数；
// - 从 Handler 推断的请求类型；
// - 支持请求体字段。
func (g *Generator) generateEnhancedRouteMessage(buf *bytes.Buffer, msgName string, route contract.RouteDef, handlerInfo *HandlerTypeInfo) {
	buf.WriteString(fmt.Sprintf("message %s {\n", msgName))

	fieldNum := 1

	// 添加路径参数
	params := g.extractPathParams(route.Path)
	for _, param := range params {
		buf.WriteString(fmt.Sprintf("  string %s = %d;\n", param, fieldNum))
		fieldNum++
	}

	// 如果有 Handler 类型信息，添加请求体字段
	if handlerInfo != nil && handlerInfo.RequestType != nil {
		if handlerInfo.RequestType.Name != "" {
			// 引用已定义的消息类型
			buf.WriteString(fmt.Sprintf("  %s body = %d;\n", handlerInfo.RequestType.Name, fieldNum))
			fieldNum++
		}
	} else {
		// 添加通用请求体占位符
		if fieldNum == 1 {
			buf.WriteString("  // Request body fields\n")
		}
	}

	buf.WriteString("}\n\n")
}

// generateEnhancedResponseMessage 生成增强的响应消息。
//
// 中文说明：
// - 从 Handler 推断的响应类型；
// - 包含标准响应字段。
func (g *Generator) generateEnhancedResponseMessage(buf *bytes.Buffer, msgName string, handlerInfo *HandlerTypeInfo) {
	buf.WriteString(fmt.Sprintf("message %s {\n", msgName))

	fieldNum := 1

	// 如果有 Handler 类型信息，使用推断的类型
	if handlerInfo != nil && handlerInfo.ResponseType != nil && handlerInfo.ResponseType.Name != "" {
		// 引用已定义的消息类型
		buf.WriteString(fmt.Sprintf("  %s data = %d;\n", handlerInfo.ResponseType.Name, fieldNum))
		fieldNum++
	}

	// 添加标准响应字段
	buf.WriteString(fmt.Sprintf("  bool success = %d;\n", fieldNum))
	fieldNum++
	buf.WriteString(fmt.Sprintf("  string message = %d;\n", fieldNum))

	buf.WriteString("}\n\n")
}

// generateRouteMessage 生成路由消息定义。
func (g *Generator) generateRouteMessage(buf *bytes.Buffer, msgName string, routes []contract.RouteDef) {
	buf.WriteString(fmt.Sprintf("message %s {\n", msgName))

	// 根据消息类型生成不同字段
	if strings.HasSuffix(msgName, "Request") {
		// 请求消息：根据路径参数生成字段
		handlerName := strings.TrimSuffix(msgName, "Request")
		for _, route := range routes {
			if g.toCamelCase(route.HandlerName) == handlerName {
				// 提取路径参数
				params := g.extractPathParams(route.Path)
				for i, param := range params {
					buf.WriteString(fmt.Sprintf("  string %s = %d;\n", param, i+1))
				}
				// 添加通用请求体字段
				buf.WriteString("  // Add request body fields here\n")
				break
			}
		}
	} else {
		// 响应消息
		buf.WriteString("  // Add response fields here\n")
		buf.WriteString("  bool success = 1;\n")
		buf.WriteString("  string message = 2;\n")
	}

	buf.WriteString("}\n\n")
}

// generateRouteMethod 生成路由方法定义。
func (g *Generator) generateRouteMethod(buf *bytes.Buffer, route contract.RouteDef, opts contract.RouteToProtoOptions) {
	methodName := g.toCamelCase(route.HandlerName)
	reqName := methodName + "Request"
	respName := methodName + "Response"

	// 转换路径格式
	protoPath := g.convertPathParams(route.Path)
	if opts.BasePath != "" {
		protoPath = opts.BasePath + protoPath
	}

	buf.WriteString(fmt.Sprintf("  // %s %s\n", route.Method, route.Path))
	buf.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s) {\n", methodName, reqName, respName))
	buf.WriteString("    option (google.api.http) = {\n")

	// 根据 HTTP 方法生成注解
	switch route.Method {
	case "GET":
		buf.WriteString(fmt.Sprintf("      get: \"%s\"\n", protoPath))
	case "POST":
		buf.WriteString(fmt.Sprintf("      post: \"%s\"\n", protoPath))
		buf.WriteString("      body: \"*\"\n")
	case "PUT":
		buf.WriteString(fmt.Sprintf("      put: \"%s\"\n", protoPath))
		buf.WriteString("      body: \"*\"\n")
	case "DELETE":
		buf.WriteString(fmt.Sprintf("      delete: \"%s\"\n", protoPath))
	case "PATCH":
		buf.WriteString(fmt.Sprintf("      patch: \"%s\"\n", protoPath))
		buf.WriteString("      body: \"*\"\n")
	default:
		buf.WriteString(fmt.Sprintf("      %s: \"%s\"\n", strings.ToLower(route.Method), protoPath))
	}

	buf.WriteString("    };\n")
	buf.WriteString("  }\n\n")
}

// extractPathParams 提取路径参数。
func (g *Generator) extractPathParams(path string) []string {
	re := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := re.FindAllStringSubmatch(path, -1)

	var params []string
	for _, match := range matches {
		params = append(params, match[1])
	}
	return params
}

// toCamelCase 转换为 CamelCase。
func (g *Generator) toCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// 分割下划线
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}
	return result.String()
}