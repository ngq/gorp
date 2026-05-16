// Package proto provides proto generator implementation.
// Core implementation of ProtoGenerator contract.
// Supports protoc invocation, Go AST parsing, Gin route parsing.
//
// Proto 包提供 proto 生成器实现。
// ProtoGenerator 契约的核心实现。
// 支持 protoc 调用、Go AST 解析、Gin 路由解析。
package proto

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// Generator 实现 ProtoGenerator 接口。
//
// 中文说明：
// - 支持三种工作流：Proto-first / Service-first / Route-first；
// - Proto-first：调用 protoc 生成 Go 代码；
// - Service-first：解析 Go AST 生成 Proto；
// - Route-first：解析 Gin 路由生成 Proto。
type Generator struct {
	cfg *integrationcontract.ProtoGeneratorConfig
}

// NewGenerator 创建 Generator。
func NewGenerator(cfg *integrationcontract.ProtoGeneratorConfig) (*Generator, error) {
	return &Generator{cfg: cfg}, nil
}

// GenFromProto 从 proto 文件生成 Go 代码（标准 protoc 流程）。
//
// 中文说明：
// - 调用 protoc 命令生成 Go 代码；
// - 支持 --go_out 和 --go-grpc_out；
// - 可选支持 --grpc-gateway_out（IncludeHTTP）。
func (g *Generator) GenFromProto(ctx context.Context, opts integrationcontract.ProtoGenOptions) error {
	// 设置默认值
	if opts.ProtoDir == "" {
		opts.ProtoDir = g.cfg.DefaultProtoDir
	}

	// 构建 protoc 参数
	args := []string{}

	// 添加导入路径
	for _, path := range opts.ImportPaths {
		args = append(args, "-I", path)
	}
	args = append(args, "-I", opts.ProtoDir)

	// 添加第三方路径
	for _, path := range g.cfg.ThirdPartyPaths {
		args = append(args, "-I", path)
	}

	// Go 代码输出
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = opts.ProtoDir
	}
	args = append(args, "--go_out="+outputDir)

	// --go_opt 参数（支持自定义配置）
	goOpt := opts.GoOpt
	if goOpt == "" {
		goOpt = "paths=source_relative"
	}
	args = append(args, "--go_opt="+goOpt)

	// gRPC 代码输出
	args = append(args, "--go-grpc_out="+outputDir)

	// --go-grpc_opt 参数（支持自定义配置）
	goGrpcOpt := opts.GoGrpcOpt
	if goGrpcOpt == "" {
		goGrpcOpt = "paths=source_relative"
	}
	args = append(args, "--go-grpc_opt="+goGrpcOpt)

	// HTTP 注解支持（grpc-gateway）
	if opts.IncludeHTTP || g.cfg.IncludeHTTPAnnotation {
		args = append(args, "--grpc-gateway_out="+outputDir)

		// --grpc-gateway_opt 参数（支持自定义配置）
		gatewayOpt := opts.GatewayOpt
		if gatewayOpt == "" {
			gatewayOpt = "paths=source_relative"
		}
		args = append(args, "--grpc-gateway_opt="+gatewayOpt)
	}

	// 自定义插件
	for pluginName, pluginOpt := range opts.CustomPlugins {
		args = append(args, "--"+pluginName+"_out="+outputDir)
		if pluginOpt != "" {
			args = append(args, "--"+pluginName+"_opt="+pluginOpt)
		}
	}

	// 添加额外插件（简单格式）
	for _, plugin := range opts.Plugins {
		args = append(args, "--"+plugin+"_out="+outputDir)
	}

	// 添加 proto 文件
	if len(opts.ProtoFiles) == 0 {
		// 扫描目录下所有 proto 文件
		files, err := g.scanProtoFiles(opts.ProtoDir)
		if err != nil {
			return fmt.Errorf("proto: scan proto files: %w", err)
		}
		args = append(args, files...)
	} else {
		args = append(args, opts.ProtoFiles...)
	}

	// 执行 protoc 命令
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("proto: protoc failed: %w", err)
	}

	return nil
}

// GenFromService 从 Go Service 接口反向生成 proto 文件。
//
// 中文说明：
// - 解析 Go AST 提取接口定义；
// - 自动扫描同 package 下全部非测试 Go 文件，收集完整 struct 定义；
// - 对 request/response 可达的自定义类型构建递归闭包；
// - 对无法解析的自定义类型直接报错，而不是生成 placeholder。
func (g *Generator) GenFromService(ctx context.Context, opts integrationcontract.ServiceToProtoOptions) error {
	_ = ctx

	// 解析 Go 文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, opts.ServicePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("proto: parse service file: %w", err)
	}

	// 提取服务定义
	services := g.extractServices(f)
	if len(services) == 0 {
		return fmt.Errorf("proto: no service interface found in %s", opts.ServicePath)
	}

	// 使用第一个服务（或匹配 ServiceName）
	var svc *integrationcontract.ServiceDef
	if opts.ServiceName != "" {
		for i := range services {
			if services[i].Name == opts.ServiceName {
				svc = &services[i]
				break
			}
		}
		if svc == nil {
			return fmt.Errorf("proto: service %s not found", opts.ServiceName)
		}
	} else {
		svc = &services[0]
	}

	// 扫描同 package 下全部非测试 Go 文件，构建基础 structDefs。
	structDefs, err := g.collectPackageStructDefs(opts.ServicePath)
	if err != nil {
		return fmt.Errorf("proto: collect package structs: %w", err)
	}

	// 从 service request/response 出发，构建根类型集合。
	initialTypes := g.collectServiceRootTypes(svc)

	// 使用 resolver 补齐跨包递归类型闭包。
	if len(opts.ImportPaths) > 0 && len(initialTypes) > 0 {
		resolver := NewTypeResolver(opts.ImportPaths)
		resolved, err := resolver.ResolveAllTypes(opts.ServicePath, initialTypes)
		if err != nil {
			return fmt.Errorf("proto: resolve imported types: %w", err)
		}
		for typeName, typeDef := range resolved {
			if _, exists := structDefs[typeName]; !exists {
				structDefs[typeName] = typeDef
			}
		}
	}

	// 前置校验：所有可达自定义类型都必须可解析。
	if err := g.validateReachableTypes(svc, structDefs); err != nil {
		return err
	}

	// 生成 proto 内容（传入结构体定义）
	protoContent := g.generateProtoContent(svc, opts, structDefs)

	// 写入文件
	if err := os.WriteFile(opts.OutputPath, []byte(protoContent), 0644); err != nil {
		return fmt.Errorf("proto: write proto file: %w", err)
	}

	// 格式化 proto 文件（可选）
	g.formatProtoFile(opts.OutputPath)

	return nil
}

// GenFromRoute 从 Gin 路由生成 proto 文件。
//
// 中文说明：
// - 解析 Gin 路由定义；
// - 推断请求/响应类型；
// - 自动添加 HTTP 注解。
func (g *Generator) GenFromRoute(ctx context.Context, opts integrationcontract.RouteToProtoOptions) error {
	// 解析路由文件
	routes, err := g.parseGinRoutes(opts.RouteFile)
	if err != nil {
		return fmt.Errorf("proto: parse route file: %w", err)
	}

	if len(routes) == 0 {
		return fmt.Errorf("proto: no routes found in %s", opts.RouteFile)
	}

	// 生成 proto 内容
	protoContent := g.generateProtoFromRoutes(routes, opts)

	// 写入文件
	if err := os.WriteFile(opts.OutputPath, []byte(protoContent), 0644); err != nil {
		return fmt.Errorf("proto: write proto file: %w", err)
	}

	// 格式化 proto 文件
	g.formatProtoFile(opts.OutputPath)

	return nil
}

// scanProtoFiles 扫描目录下所有 proto 文件。
func (g *Generator) scanProtoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// extractStructDefs 从 Go AST 提取所有结构体定义。
//
// 中文说明：
// - 解析文件中的所有 type StructName struct 定义；
// - 提取每个结构体的字段信息和注释；
// - 返回结构体名到类型定义的映射（包含字段和注释）。
func (g *Generator) extractStructDefs(f *ast.File) map[string]*integrationcontract.TypeDef {
	structDefs := make(map[string]*integrationcontract.TypeDef)

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// 提取结构体注释（从 GenDecl.Doc）
			var comments []string
			if genDecl.Doc != nil {
				for _, c := range genDecl.Doc.List {
					text := strings.TrimPrefix(c.Text, "//")
					text = strings.TrimSpace(text)
					if text != "" {
						comments = append(comments, text)
					}
				}
			}

			// 提取结构体字段
			fields := g.extractStructFields(structType)

			structDefs[typeSpec.Name.Name] = &integrationcontract.TypeDef{
				Name:     typeSpec.Name.Name,
				Fields:   fields,
				Comments: comments,
			}
		}
	}

	return structDefs
}

// extractStructFields 从结构体 AST 提取字段定义。
//
// 中文说明：
// - 遍历结构体的所有字段；
// - 解析字段名、类型和 tag；
// - 生成 FieldDef 列表。
func (g *Generator) extractStructFields(structType *ast.StructType) []integrationcontract.FieldDef {
	var fields []integrationcontract.FieldDef

	if structType.Fields == nil {
		return fields
	}

	for _, field := range structType.Fields.List {
		// 提取字段名
		var fieldName string
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		} else {
			// 嵌入字段（匿名字段），跳过
			continue
		}

		// 提取字段类型
		typeDef := g.extractType(field.Type)

		// 提取 JSON tag 和 remark tag
		var jsonName, protoName, remark string
		if field.Tag != nil {
			tagValue := strings.Trim(field.Tag.Value, "`")
			jsonName = g.extractTagValue(tagValue, "json")
			protoName = g.extractTagValue(tagValue, "proto")
			// 优先提取 remark tag（用于字段描述）
			remark = g.extractTagValue(tagValue, "remark")
		}

		// 如果没有 JSON tag，使用字段名转 snake_case
		if jsonName == "" {
			jsonName = fieldName
		}
		// 如果没有 proto tag，使用 JSON 名转 snake_case
		if protoName == "" {
			protoName = g.toSnakeCase(jsonName)
		}

		// 提取注释（包括字段前注释和字段后注释）
		// 如果有 remark tag，优先使用 remark，不提取 // 注释
		var comments []string
		if remark == "" {
			// 字段前注释（Doc，字段上一行）
			if field.Doc != nil {
				for _, c := range field.Doc.List {
					text := strings.TrimPrefix(c.Text, "//")
					text = strings.TrimSpace(text)
					if text != "" {
						comments = append(comments, text)
					}
				}
			}
			// 字段后注释（Comment，字段同一行末尾）
			if field.Comment != nil {
				for _, c := range field.Comment.List {
					text := strings.TrimPrefix(c.Text, "//")
					text = strings.TrimSpace(text)
					if text != "" {
						comments = append(comments, text)
					}
				}
			}
		}

		fields = append(fields, integrationcontract.FieldDef{
			Name:      fieldName,
			JSONName:  jsonName,
			ProtoName: protoName,
			Type:      typeDef,
			Comments:  comments,
			Remark:    remark,
		})
	}

	return fields
}

// extractTagValue 从 struct tag 中提取指定键的值。
//
// 中文说明：
// - 解析 Go struct tag 字符串；
// - 支持提取 json、proto、remark 等标签；
// - remark 标签的值可能包含逗号（如 "状态:0-草稿,1-发布"）。
func (g *Generator) extractTagValue(tagStr, key string) string {
	// 对于 remark 标签，值可能包含逗号，需要特殊处理
	if key == "remark" {
		// 格式: remark:"value" - value 可以包含逗号
		re := regexp.MustCompile(key + `:"([^"]*)"`)
		matches := re.FindStringSubmatch(tagStr)
		if len(matches) > 1 {
			return matches[1]
		}
		return ""
	}

	// 其他标签（json、proto 等）: key:"value,options" 或 key:value
	// 值部分不含逗号，逗号后是选项
	re := regexp.MustCompile(key + `:"([^,"]+)(?:,[^"]*)?"`)
	matches := re.FindStringSubmatch(tagStr)
	if len(matches) > 1 {
		return matches[1]
	}
	// 尝试无引号格式
	re2 := regexp.MustCompile(key + `:([^,\s]+)`)
	matches2 := re2.FindStringSubmatch(tagStr)
	if len(matches2) > 1 {
		return matches2[1]
	}
	return ""
}

// toSnakeCase 将字符串转换为 snake_case。
//
// 中文说明：
// - CamelCase → camel_case；
// - 用于 proto 字段命名。
func (g *Generator) toSnakeCase(s string) string {
	// 处理已经是 snake_case 的情况
	if strings.Contains(s, "_") {
		return s
	}

	var result strings.Builder
	for i, ch := range s {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteByte(byte(ch))
	}
	return strings.ToLower(result.String())
}

// extractServices 从 Go AST 提取服务接口。
//
// 中文说明：
// - 解析 Go 文件中的接口定义；
// - 提取接口名称、注释和方法列表；
// - 返回 ServiceDef 列表。
func (g *Generator) extractServices(f *ast.File) []integrationcontract.ServiceDef {
	var services []integrationcontract.ServiceDef

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			svc := integrationcontract.ServiceDef{
				Name: typeSpec.Name.Name,
			}

			// 提取接口注释（优先使用 typeSpec.Doc，其次使用 genDecl.Doc）
			var doc *ast.CommentGroup
			if typeSpec.Doc != nil {
				doc = typeSpec.Doc
			} else if genDecl.Doc != nil {
				doc = genDecl.Doc
			}
			if doc != nil {
				for _, c := range doc.List {
					text := strings.TrimPrefix(c.Text, "//")
					text = strings.TrimSpace(text)
					if text != "" {
						svc.Comments = append(svc.Comments, text)
					}
				}
			}

			// 提取方法
			for _, method := range ifaceType.Methods.List {
				if len(method.Names) == 0 {
					continue
				}

				fnType, ok := method.Type.(*ast.FuncType)
				if !ok {
					continue
				}

				methodDef := g.extractMethod(fnType, method.Names[0].Name, method)
				svc.Methods = append(svc.Methods, methodDef)
			}

			if len(svc.Methods) > 0 {
				services = append(services, svc)
			}
		}
	}

	return services
}

// extractMethod 提取方法定义。
//
// 中文说明：
// - 从接口方法签名中提取请求和响应类型；
// - 自动跳过 context.Context 参数和 error 返回值；
// - 解析指针、切片等复合类型；
// - 提取方法注释。
func (g *Generator) extractMethod(fnType *ast.FuncType, name string, field *ast.Field) integrationcontract.MethodDef {
	method := integrationcontract.MethodDef{
		Name: name,
	}

	// 提取方法注释
	if field.Doc != nil {
		for _, c := range field.Doc.List {
			text := strings.TrimPrefix(c.Text, "//")
			text = strings.TrimSpace(text)
			if text != "" {
				method.Comments = append(method.Comments, text)
			}
		}
	}

	// 提取请求类型（第一个参数）
	if fnType.Params != nil && len(fnType.Params.List) > 0 {
		// 跳过 context.Context 参数
		for i, param := range fnType.Params.List {
			if g.isContextType(param.Type) {
				continue
			}
			if i > 0 || len(fnType.Params.List) == 1 {
				method.RequestType = g.extractType(param.Type)
				break
			}
		}
	}

	// 提取响应类型（返回值）
	if fnType.Results != nil && len(fnType.Results.List) > 0 {
		for _, result := range fnType.Results.List {
			if g.isErrorType(result.Type) {
				continue
			}
			method.ResponseType = g.extractType(result.Type)
			break
		}
	}

	return method
}

// extractType 提取类型定义。
//
// 中文说明：
// - 从 AST 表达式提取类型信息；
// - 支持简单类型、指针类型、切片类型、Map 类型；
// - 支持跨包类型（如 time.Time）；
// - Map 类型会提取 key 和 value 的完整类型信息。
func (g *Generator) extractType(expr ast.Expr) *integrationcontract.TypeDef {
	switch t := expr.(type) {
	case *ast.Ident:
		return &integrationcontract.TypeDef{
			Name: t.Name,
		}
	case *ast.SelectorExpr:
		return &integrationcontract.TypeDef{
			Name:    t.Sel.Name,
			Package: g.getPackageName(t.X),
		}
	case *ast.StarExpr:
		inner := g.extractType(t.X)
		inner.IsPointer = true
		return inner
	case *ast.ArrayType:
		inner := g.extractType(t.Elt)
		inner.IsSlice = true
		return inner
	case *ast.MapType:
		// 完整提取 Map 的 key 和 value 类型
		return &integrationcontract.TypeDef{
			IsMap:    true,
			MapKey:   g.extractType(t.Key),
			MapValue: g.extractType(t.Value),
		}
	default:
		return &integrationcontract.TypeDef{
			Name: "any",
		}
	}
}

// getPackageName 获取包名。
//
// 中文说明：
// - 从 SelectorExpr 中提取包名部分；
// - 例如 time.Time 返回 "time"。
func (g *Generator) getPackageName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// isContextType 判断是否为 context.Context 类型。
//
// 中文说明：
// - 用于识别并跳过方法的第一个参数；
// - Proto RPC 方法不需要传递 Context。
func (g *Generator) isContextType(expr ast.Expr) bool {
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "Context" {
			return true
		}
	}
	return false
}

// isErrorType 判断是否为 error 类型。
//
// 中文说明：
// - 用于识别并跳过方法的 error 返回值；
// - Proto RPC 方法不直接返回 error，由 gRPC 框架处理。
func (g *Generator) isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "error"
	}
	return false
}

// collectPackageStructDefs 收集 service 所在 package 的全部结构体定义。
//
// 中文说明：
// - 扫描 service 文件所在目录下全部非 `_test.go` 文件；
// - 只合并与 service 文件同 package 的 Go 文件；
// - 统一复用 generator 的字段提取逻辑，保证同包解析口径一致。
func (g *Generator) collectPackageStructDefs(servicePath string) (map[string]*integrationcontract.TypeDef, error) {
	fset := token.NewFileSet()
	serviceFile, err := parser.ParseFile(fset, servicePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse service file: %w", err)
	}

	serviceDir := filepath.Dir(servicePath)
	files, err := filepath.Glob(filepath.Join(serviceDir, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("scan package go files: %w", err)
	}

	structDefs := make(map[string]*integrationcontract.TypeDef)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse package file %s: %w", file, err)
		}
		if parsed.Name == nil || serviceFile.Name == nil || parsed.Name.Name != serviceFile.Name.Name {
			continue
		}

		for typeName, typeDef := range g.extractStructDefs(parsed) {
			structDefs[typeName] = typeDef
		}
	}

	return structDefs, nil
}

// collectServiceRootTypes 收集 service 方法签名中的根类型。
//
// 中文说明：
// - 只收集 request / response 的根自定义类型；
// - 内置类型会被过滤；
// - 返回值用于驱动跨包递归闭包解析。
func (g *Generator) collectServiceRootTypes(svc *integrationcontract.ServiceDef) []string {
	if svc == nil {
		return nil
	}

	var roots []string
	for _, method := range svc.Methods {
		roots = append(roots, g.collectRootTypeName(method.RequestType)...)
		roots = append(roots, g.collectRootTypeName(method.ResponseType)...)
	}
	return uniqueStrings(roots)
}

// collectRootTypeName 从单个类型中提取根自定义类型名。
func (g *Generator) collectRootTypeName(typeDef *integrationcontract.TypeDef) []string {
	if typeDef == nil {
		return nil
	}
	if typeDef.IsMap {
		refs := g.collectRootTypeName(typeDef.MapKey)
		refs = append(refs, g.collectRootTypeName(typeDef.MapValue)...)
		return uniqueStrings(refs)
	}

	fullName := fullTypeName(typeDef)
	if fullName == "" || g.isBuiltInType(fullName) || g.isBuiltInType(typeDef.Name) {
		return nil
	}
	return []string{fullName}
}

// validateReachableTypes 校验 service 可达的所有自定义类型都已经可解析。
//
// 中文说明：
// - 从每个方法的 request / response 出发递归遍历字段；
// - 一旦发现缺失类型或 proto 不支持的 map 组合，立即报错；
// - 错误信息带上方法和字段来源，避免生成半成品 proto。
func (g *Generator) validateReachableTypes(svc *integrationcontract.ServiceDef, structDefs map[string]*integrationcontract.TypeDef) error {
	if svc == nil {
		return nil
	}

	seen := make(map[string]bool)
	for _, method := range svc.Methods {
		if err := g.validateReachableType(method.Name, "request", nil, method.RequestType, structDefs, seen); err != nil {
			return err
		}
		if err := g.validateReachableType(method.Name, "response", nil, method.ResponseType, structDefs, seen); err != nil {
			return err
		}
	}
	return nil
}

// validateReachableType 递归校验单个类型节点是否可安全生成 proto。
func (g *Generator) validateReachableType(methodName, position string, fieldPath []string, typeDef *integrationcontract.TypeDef, structDefs map[string]*integrationcontract.TypeDef, seen map[string]bool) error {
	if typeDef == nil {
		return nil
	}

	if typeDef.IsMap {
		if typeDef.MapKey == nil || typeDef.MapValue == nil {
			return fmt.Errorf("proto: method %s %s field %s uses incomplete map type", methodName, position, g.describeFieldPath(fieldPath))
		}
		if !g.isValidProtoMapKey(typeDef.MapKey) {
			return fmt.Errorf("proto: method %s %s field %s uses unsupported map key type %s", methodName, position, g.describeFieldPath(fieldPath), fullTypeName(typeDef.MapKey))
		}
		if typeDef.MapValue.IsMap || (typeDef.MapValue.IsSlice && !g.isBytesField(typeDef.MapValue)) {
			return fmt.Errorf("proto: method %s %s field %s uses unsupported map value type %s", methodName, position, g.describeFieldPath(fieldPath), g.describeType(typeDef.MapValue))
		}
		if err := g.validateReachableType(methodName, position, fieldPath, typeDef.MapValue, structDefs, seen); err != nil {
			return err
		}
		return nil
	}

	fullName := fullTypeName(typeDef)
	if fullName == "" || g.isBuiltInType(fullName) || g.isBuiltInType(typeDef.Name) {
		return nil
	}

	lookupName := typeDef.Name
	if lookupName == "" {
		lookupName = fullName
	}
	if seen[lookupName] {
		return nil
	}

	structDef, ok := structDefs[lookupName]
	if !ok {
		return fmt.Errorf("proto: method %s %s field %s references unresolved type %s", methodName, position, g.describeFieldPath(fieldPath), fullName)
	}
	seen[lookupName] = true

	for _, field := range structDef.Fields {
		childPath := append(append([]string{}, fieldPath...), field.Name)
		if err := g.validateReachableType(methodName, position, childPath, field.Type, structDefs, seen); err != nil {
			return err
		}
	}
	return nil
}

// isValidProtoMapKey 判断 map key 是否符合 proto3 约束。
func (g *Generator) isValidProtoMapKey(typeDef *integrationcontract.TypeDef) bool {
	if typeDef == nil || typeDef.IsMap || typeDef.IsSlice {
		return false
	}

	switch fullTypeName(typeDef) {
	case "string", "int32", "int64", "uint32", "uint64", "bool":
		return true
	case "int", "uint":
		return true
	default:
		return false
	}
}

// describeFieldPath 生成错误信息中的字段路径。
func (g *Generator) describeFieldPath(fieldPath []string) string {
	if len(fieldPath) == 0 {
		return "<root>"
	}
	return strings.Join(fieldPath, ".")
}

// describeType 生成人类可读的类型描述。
func (g *Generator) describeType(typeDef *integrationcontract.TypeDef) string {
	if typeDef == nil {
		return "<nil>"
	}
	if typeDef.IsMap {
		return fmt.Sprintf("map[%s]%s", g.describeType(typeDef.MapKey), g.describeType(typeDef.MapValue))
	}

	name := fullTypeName(typeDef)
	if name == "" {
		name = typeDef.Name
	}
	if name == "" {
		name = "<anonymous>"
	}
	if typeDef.IsSlice {
		name = "[]" + name
	}
	if typeDef.IsPointer {
		name = "*" + name
	}
	return name
}

// serviceUsesProtoAny 判断当前 service 可达类型里是否使用了 google.protobuf.Any。
func (g *Generator) serviceUsesProtoAny(svc *integrationcontract.ServiceDef, structDefs map[string]*integrationcontract.TypeDef) bool {
	if svc == nil {
		return false
	}
	seen := make(map[string]bool)
	for _, method := range svc.Methods {
		if g.typeUsesProtoAny(method.RequestType, structDefs, seen) || g.typeUsesProtoAny(method.ResponseType, structDefs, seen) {
			return true
		}
	}
	return false
}

// typeUsesProtoAny 递归判断类型树中是否包含 any / interface{}。
func (g *Generator) typeUsesProtoAny(typeDef *integrationcontract.TypeDef, structDefs map[string]*integrationcontract.TypeDef, seen map[string]bool) bool {
	if typeDef == nil {
		return false
	}
	if typeDef.IsMap {
		return g.typeUsesProtoAny(typeDef.MapKey, structDefs, seen) || g.typeUsesProtoAny(typeDef.MapValue, structDefs, seen)
	}
	if g.goTypeToProtoType(typeDef) == "google.protobuf.Any" {
		return true
	}

	fullName := fullTypeName(typeDef)
	if fullName == "" || g.isBuiltInType(fullName) || g.isBuiltInType(typeDef.Name) {
		return false
	}

	lookupName := typeDef.Name
	if lookupName == "" {
		lookupName = fullName
	}
	if seen[lookupName] {
		return false
	}
	seen[lookupName] = true

	structDef, ok := structDefs[lookupName]
	if !ok {
		return false
	}
	for _, field := range structDef.Fields {
		if g.typeUsesProtoAny(field.Type, structDefs, seen) {
			return true
		}
	}
	return false
}

// generateProtoContent 生成 proto 文件内容。
//
// 中文说明：
// - 生成 proto 文件头部、message 定义和 service 定义；
// - 使用 structDefs 填充 message 字段；
// - 避免重复生成相同的 message。
func (g *Generator) generateProtoContent(svc *integrationcontract.ServiceDef, opts integrationcontract.ServiceToProtoOptions, structDefs map[string]*integrationcontract.TypeDef) string {
	var buf bytes.Buffer

	// 头部
	buf.WriteString("syntax = \"proto3\";\n\n")
	buf.WriteString(fmt.Sprintf("package %s;\n\n", opts.Package))

	// HTTP 注解导入
	if opts.IncludeHTTP {
		buf.WriteString("import \"google/api/annotations.proto\";\n")
	}
	if g.serviceUsesProtoAny(svc, structDefs) {
		buf.WriteString("import \"google/protobuf/any.proto\";\n")
	}
	if opts.IncludeHTTP || g.serviceUsesProtoAny(svc, structDefs) {
		buf.WriteString("\n")
	}

	// Go package
	buf.WriteString(fmt.Sprintf("option go_package = \"%s\";\n\n", opts.GoPackage))

	// 收集所有需要生成的消息类型（避免重复）
	generatedMsgs := make(map[string]bool)

	// Messages
	buf.WriteString("// Messages generated from Go types\n\n")
	for _, method := range svc.Methods {
		// 生成请求类型
		if method.RequestType != nil && method.RequestType.Name != "" {
			msgName := method.RequestType.Name
			if !generatedMsgs[msgName] {
				generatedMsgs[msgName] = true
				g.generateMessage(&buf, msgName, structDefs, generatedMsgs)
			}
		}
		// 生成响应类型
		if method.ResponseType != nil && method.ResponseType.Name != "" {
			msgName := method.ResponseType.Name
			if !generatedMsgs[msgName] {
				generatedMsgs[msgName] = true
				g.generateMessage(&buf, msgName, structDefs, generatedMsgs)
			}
		}
	}

	// Service
	// 写入 Service 注释（如果有）
	if len(svc.Comments) > 0 {
		for _, c := range svc.Comments {
			buf.WriteString(fmt.Sprintf("// %s\n", strings.TrimSpace(c)))
		}
	} else {
		// 默认注释
		buf.WriteString(fmt.Sprintf("// %s service definition.\n", svc.Name))
	}
	buf.WriteString(fmt.Sprintf("service %s {\n", svc.Name))
	for _, method := range svc.Methods {
		g.generateMethod(&buf, method, opts)
	}
	buf.WriteString("}\n")

	return buf.String()
}

// generateMessage 生成 message 定义。
//
// 中文说明：
// - 根据类型名称从 structDefs 获取类型定义（包含字段和注释）；
// - 先完整输出当前 message，再递归输出依赖 message；
// - 这样可以避免把嵌套 message 插入到父 message 体内。
func (g *Generator) generateMessage(buf *bytes.Buffer, name string, structDefs map[string]*integrationcontract.TypeDef, generatedMsgs map[string]bool) {
	structDef := structDefs[name]
	if structDef == nil {
		return
	}

	var nestedMessages []string

	// 生成结构体注释（如果有）
	if len(structDef.Comments) > 0 {
		for _, c := range structDef.Comments {
			buf.WriteString(fmt.Sprintf("// %s\n", strings.TrimSpace(c)))
		}
	}

	buf.WriteString(fmt.Sprintf("message %s {\n", name))

	fieldNum := 1
	for _, field := range structDef.Fields {
		if field.Type == nil {
			continue
		}

		nestedMessages = append(nestedMessages, g.collectNestedMessageNames(field.Type, structDefs, generatedMsgs)...)

		protoType := g.goTypeToProtoType(field.Type)
		fieldDef := g.buildProtoFieldDefinition(protoType, field, fieldNum)
		buf.WriteString(fieldDef)
		fieldNum++
	}

	buf.WriteString("}\n\n")

	for _, nestedName := range uniqueStrings(nestedMessages) {
		if generatedMsgs[nestedName] {
			continue
		}
		generatedMsgs[nestedName] = true
		g.generateMessage(buf, nestedName, structDefs, generatedMsgs)
	}
}

// collectNestedMessageNames 收集字段依赖的嵌套自定义消息名。
func (g *Generator) collectNestedMessageNames(typeDef *integrationcontract.TypeDef, structDefs map[string]*integrationcontract.TypeDef, generatedMsgs map[string]bool) []string {
	if typeDef == nil {
		return nil
	}
	if typeDef.IsMap {
		return g.collectNestedMessageNames(typeDef.MapValue, structDefs, generatedMsgs)
	}

	fullName := fullTypeName(typeDef)
	if fullName == "" || g.isBuiltInType(fullName) || g.isBuiltInType(typeDef.Name) {
		return nil
	}

	msgName := typeDef.Name
	if msgName == "" {
		msgName = fullName
	}
	if generatedMsgs[msgName] {
		return nil
	}
	if _, found := structDefs[msgName]; !found {
		return nil
	}
	return []string{msgName}
}

// isBytesField 判断字段是否应映射为 proto bytes。
func (g *Generator) isBytesField(typeDef *integrationcontract.TypeDef) bool {
	if typeDef == nil {
		return false
	}
	return typeDef.IsSlice && typeDef.Name == "byte" && typeDef.Package == ""
}

// buildProtoFieldDefinition 根据字段类型构造 proto 字段声明。
func (g *Generator) buildProtoFieldDefinition(protoType string, field integrationcontract.FieldDef, fieldNum int) string {
	var fieldDef string
	switch {
	case field.Type.IsMap:
		fieldDef = fmt.Sprintf("  %s %s = %d", protoType, field.ProtoName, fieldNum)
	case field.Type.IsSlice && !g.isBytesField(field.Type):
		fieldDef = fmt.Sprintf("  repeated %s %s = %d", protoType, field.ProtoName, fieldNum)
	default:
		fieldDef = fmt.Sprintf("  %s %s = %d", protoType, field.ProtoName, fieldNum)
	}

	if field.Remark != "" {
		return fmt.Sprintf("%s; // %s\n", fieldDef, field.Remark)
	}
	if len(field.Comments) > 0 {
		return fmt.Sprintf("%s; // %s\n", fieldDef, strings.Join(field.Comments, "; "))
	}
	return fieldDef + ";\n"
}

// goTypeToProtoType 将 Go 类型转换为 Proto 类型。
//
// 中文说明：
// - 基本类型映射：string→string, int→int32, uint64→uint64 等；
// - Map 类型生成 map<key, value> 语法；
// - selector 类型按 Package+Name 一起判断；
// - 切片类型由调用方处理（使用 repeated）。
func (g *Generator) goTypeToProtoType(typeDef *integrationcontract.TypeDef) string {
	if typeDef == nil {
		return "string"
	}

	if typeDef.IsMap {
		keyType := g.goTypeToProtoType(typeDef.MapKey)
		valueType := g.goTypeToProtoType(typeDef.MapValue)
		return fmt.Sprintf("map<%s, %s>", keyType, valueType)
	}

	if protoType, ok := protoBuiltinType(fullTypeName(typeDef)); ok {
		return protoType
	}
	if protoType, ok := protoBuiltinType(typeDef.Name); ok {
		return protoType
	}
	return typeDef.Name
}

// isBuiltInType 判断是否为内置类型。
//
// 中文说明：
// - 用于区分内置类型和自定义结构体；
// - 自定义结构体需要生成对应的 message 定义。
func (g *Generator) isBuiltInType(typeName string) bool {
	_, ok := protoBuiltinType(typeName)
	return ok
}

// generateMethod 生成 service 方法。
//
// 中文说明：
// - 生成 rpc 方法定义；
// - 包含方法注释（从 Go 代码提取）；
// - 添加 HTTP 注解（如果配置）。
func (g *Generator) generateMethod(buf *bytes.Buffer, method integrationcontract.MethodDef, opts integrationcontract.ServiceToProtoOptions) {
	reqName := method.Name + "Request"
	if method.RequestType != nil && method.RequestType.Name != "" {
		reqName = method.RequestType.Name
	}

	respName := method.Name + "Response"
	if method.ResponseType != nil && method.ResponseType.Name != "" {
		respName = method.ResponseType.Name
	}

	// 写入方法注释（如果有）
	if len(method.Comments) > 0 {
		for _, c := range method.Comments {
			buf.WriteString(fmt.Sprintf("  // %s\n", strings.TrimSpace(c)))
		}
	} else {
		// 默认注释：方法名
		buf.WriteString(fmt.Sprintf("  // %s RPC method\n", method.Name))
	}

	buf.WriteString(fmt.Sprintf("  rpc %s(%s) returns (%s)", method.Name, reqName, respName))

	// HTTP 注解
	if opts.IncludeHTTP {
		if rule, ok := opts.HTTPAnnotations[method.Name]; ok {
			buf.WriteString(" {\n")
			buf.WriteString("    option (google.api.http) = {\n")
			buf.WriteString(fmt.Sprintf("      %s: \"%s\"\n", strings.ToLower(rule.Method), rule.Path))
			if rule.Body != "" {
				buf.WriteString(fmt.Sprintf("      body: \"%s\"\n", rule.Body))
			}
			buf.WriteString("    };\n")
			buf.WriteString("  }\n")
		} else {
			buf.WriteString(";\n")
		}
	} else {
		buf.WriteString(";\n")
	}
}

// formatProtoFile 格式化 proto 文件。
// Uses buf format if available; gracefully degrades when buf is not installed.
func (g *Generator) formatProtoFile(path string) {
	cmd := exec.Command("buf", "format", "-w", path)
	if err := cmd.Run(); err != nil {
		// buf not installed or format failed — non-fatal, the proto file is still valid.
		// buf 未安装或格式化失败——非致命，proto 文件仍然有效。
		fmt.Printf("[proto:info] buf format skipped for %s: %v\n", path, err)
	}
}

// GetConfig 获取当前配置。
func (g *Generator) GetConfig() *integrationcontract.ProtoGeneratorConfig {
	return g.cfg
}

// isValidGoIdentifier 检查字符串是否为有效的 Go 标识符。
// 有效的 Go 标识符格式：[A-Za-z_][A-Za-z0-9_]*
func isValidGoIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	// 检查第一个字符
	firstChar := name[0]
	if !(firstChar >= 'A' && firstChar <= 'Z') && !(firstChar >= 'a' && firstChar <= 'z') && firstChar != '_' {
		return false
	}

	// 检查后续字符
	for i := 1; i < len(name); i++ {
		char := name[i]
		if !(char >= 'A' && char <= 'Z') && !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '_' {
			return false
		}
	}

	return true
}
