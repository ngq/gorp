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

	"github.com/ngq/gorp/framework/contract"
)

// Generator 实现 ProtoGenerator 接口。
//
// 中文说明：
// - 支持三种工作流：Proto-first / Service-first / Route-first；
// - Proto-first：调用 protoc 生成 Go 代码；
// - Service-first：解析 Go AST 生成 Proto；
// - Route-first：解析 Gin 路由生成 Proto。
type Generator struct {
	cfg *contract.ProtoGeneratorConfig
}

// NewGenerator 创建 Generator。
func NewGenerator(cfg *contract.ProtoGeneratorConfig) (*Generator, error) {
	return &Generator{cfg: cfg}, nil
}

// GenFromProto 从 proto 文件生成 Go 代码（标准 protoc 流程）。
//
// 中文说明：
// - 调用 protoc 命令生成 Go 代码；
// - 支持 --go_out 和 --go-grpc_out；
// - 可选支持 --grpc-gateway_out（IncludeHTTP）。
func (g *Generator) GenFromProto(ctx context.Context, opts contract.ProtoGenOptions) error {
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
// - 提取请求/响应类型的完整字段定义；
// - 支持跨文件类型解析（当提供 ImportPaths 时）；
// - 生成 proto service 和 messages。
func (g *Generator) GenFromService(ctx context.Context, opts contract.ServiceToProtoOptions) error {
	// 解析 Go 文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, opts.ServicePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("proto: parse service file: %w", err)
	}

	// 提取当前文件的结构体定义
	structDefs := g.extractStructDefs(f)

	// 如果提供了 ImportPaths，尝试解析跨文件类型
	if len(opts.ImportPaths) > 0 {
		resolver := NewTypeResolver(opts.ImportPaths)

		// 收集所有需要解析的类型
		var typesToResolve []string
		for _, typeDef := range structDefs {
			for _, field := range typeDef.Fields {
				if field.Type != nil && field.Type.Name != "" && !g.isBuiltInType(field.Type.Name) {
					typesToResolve = append(typesToResolve, field.Type.Name)
				}
			}
		}

		// 解析跨文件类型
		if len(typesToResolve) > 0 {
			baseDir := filepath.Dir(opts.ServicePath)
			for _, importPath := range opts.ImportPaths {
				resolved, err := resolver.ResolveTypesFromImport(importPath, typesToResolve, baseDir)
				if err == nil {
					for typeName, resolvedTypeDef := range resolved {
						if _, exists := structDefs[typeName]; !exists {
							structDefs[typeName] = resolvedTypeDef
						}
					}
				}
			}
		}
	}

	// 提取服务定义
	services := g.extractServices(f)

	if len(services) == 0 {
		return fmt.Errorf("proto: no service interface found in %s", opts.ServicePath)
	}

	// 使用第一个服务（或匹配 ServiceName）
	var svc *contract.ServiceDef
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
func (g *Generator) GenFromRoute(ctx context.Context, opts contract.RouteToProtoOptions) error {
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
func (g *Generator) extractStructDefs(f *ast.File) map[string]*contract.TypeDef {
	structDefs := make(map[string]*contract.TypeDef)

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

			structDefs[typeSpec.Name.Name] = &contract.TypeDef{
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
func (g *Generator) extractStructFields(structType *ast.StructType) []contract.FieldDef {
	var fields []contract.FieldDef

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

		fields = append(fields, contract.FieldDef{
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
func (g *Generator) extractServices(f *ast.File) []contract.ServiceDef {
	var services []contract.ServiceDef

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

			svc := contract.ServiceDef{
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
func (g *Generator) extractMethod(fnType *ast.FuncType, name string, field *ast.Field) contract.MethodDef {
	method := contract.MethodDef{
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
func (g *Generator) extractType(expr ast.Expr) *contract.TypeDef {
	switch t := expr.(type) {
	case *ast.Ident:
		return &contract.TypeDef{
			Name: t.Name,
		}
	case *ast.SelectorExpr:
		return &contract.TypeDef{
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
		return &contract.TypeDef{
			IsMap:    true,
			MapKey:   g.extractType(t.Key),
			MapValue: g.extractType(t.Value),
		}
	default:
		return &contract.TypeDef{
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

// generateProtoContent 生成 proto 文件内容。
//
// 中文说明：
// - 生成 proto 文件头部、message 定义和 service 定义；
// - 使用 structDefs 填充 message 字段；
// - 避免重复生成相同的 message。
func (g *Generator) generateProtoContent(svc *contract.ServiceDef, opts contract.ServiceToProtoOptions, structDefs map[string]*contract.TypeDef) string {
	var buf bytes.Buffer

	// 头部
	buf.WriteString("syntax = \"proto3\";\n\n")
	buf.WriteString(fmt.Sprintf("package %s;\n\n", opts.Package))

	// HTTP 注解导入
	if opts.IncludeHTTP {
		buf.WriteString("import \"google/api/annotations.proto\";\n\n")
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
				g.generateMessage(&buf, method.RequestType, msgName, structDefs, generatedMsgs)
				generatedMsgs[msgName] = true
			}
		}
		// 生成响应类型
		if method.ResponseType != nil && method.ResponseType.Name != "" {
			msgName := method.ResponseType.Name
			if !generatedMsgs[msgName] {
				g.generateMessage(&buf, method.ResponseType, msgName, structDefs, generatedMsgs)
				generatedMsgs[msgName] = true
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
// - 将 Go 类型映射到 Proto 类型；
// - 处理嵌套结构体（递归生成）；
// - 正确处理 Map 类型和切片类型；
// - 生成结构体注释（如果有）。
func (g *Generator) generateMessage(buf *bytes.Buffer, typeDef *contract.TypeDef, name string, structDefs map[string]*contract.TypeDef, generatedMsgs map[string]bool) {
	// 从 structDefs 获取类型定义
	structDef, ok := structDefs[name]

	// 生成结构体注释（如果有）
	if ok && len(structDef.Comments) > 0 {
		for _, c := range structDef.Comments {
			buf.WriteString(fmt.Sprintf("// %s\n", strings.TrimSpace(c)))
		}
	}

	buf.WriteString(fmt.Sprintf("message %s {\n", name))

	if !ok {
		// 未找到结构体定义，生成占位字段
		buf.WriteString("  // TODO: Add fields from Go struct\n")
		buf.WriteString("  string placeholder = 1;\n")
	} else {
		// 生成字段定义
		fieldNum := 1
		for _, field := range structDef.Fields {
			if field.Type == nil {
				continue
			}

			// 处理嵌套结构体
			if field.Type.Name != "" && !g.isBuiltInType(field.Type.Name) && !field.Type.IsMap {
				// 如果是自定义类型且未生成，递归生成
				if _, found := structDefs[field.Type.Name]; found && !generatedMsgs[field.Type.Name] {
					generatedMsgs[field.Type.Name] = true
					g.generateMessage(buf, field.Type, field.Type.Name, structDefs, generatedMsgs)
				}
			}

			// 生成字段定义
			protoType := g.goTypeToProtoType(field.Type)

			// 处理不同类型
			var fieldDef string
			if field.Type.IsMap {
				// Map 类型
				if field.Type.MapValue != nil && field.Type.MapValue.IsSlice {
					fieldDef = fmt.Sprintf("  %s %s = %d; // WARNING: map value slice not supported\n", protoType, field.ProtoName, fieldNum)
				} else {
					fieldDef = fmt.Sprintf("  %s %s = %d", protoType, field.ProtoName, fieldNum)
				}
			} else if field.Type.IsSlice {
				// repeated 类型
				fieldDef = fmt.Sprintf("  repeated %s %s = %d", protoType, field.ProtoName, fieldNum)
			} else {
				// 普通类型
				fieldDef = fmt.Sprintf("  %s %s = %d", protoType, field.ProtoName, fieldNum)
			}

			// 添加字段注释（优先使用 remark tag，其次是 // 注释）
			// 如果有，放在字段后面
			if field.Remark != "" {
				// 优先使用 remark tag
				fieldDef = fmt.Sprintf("%s; // %s\n", fieldDef, field.Remark)
			} else if len(field.Comments) > 0 {
				// 其次使用 // 注释
				commentText := strings.Join(field.Comments, "; ")
				fieldDef = fmt.Sprintf("%s; // %s\n", fieldDef, commentText)
			} else {
				fieldDef = fieldDef + ";\n"
			}

			buf.WriteString(fieldDef)
			fieldNum++
		}
	}

	buf.WriteString("}\n\n")
}

// goTypeToProtoType 将 Go 类型转换为 Proto 类型。
//
// 中文说明：
// - 基本类型映射：string→string, int→int32, uint64→uint64 等；
// - Map 类型生成 map<key, value> 语法；
// - 自定义类型保持原名；
// - 切片类型由调用方处理（使用 repeated）。
func (g *Generator) goTypeToProtoType(typeDef *contract.TypeDef) string {
	if typeDef == nil {
		return "string"
	}

	// 处理 Map 类型
	if typeDef.IsMap {
		keyType := g.goTypeToProtoType(typeDef.MapKey)
		valueType := g.goTypeToProtoType(typeDef.MapValue)
		return fmt.Sprintf("map<%s, %s>", keyType, valueType)
	}

	// 处理指针类型（Proto3 不区分指针）
	typeName := typeDef.Name

	// Go 基本类型到 Proto 类型映射
	typeMap := map[string]string{
		"string":        "string",
		"int":           "int32",
		"int32":         "int32",
		"int64":         "int64",
		"uint":          "uint32",
		"uint32":        "uint32",
		"uint64":        "uint64",
		"float32":       "float",
		"float64":       "double",
		"bool":          "bool",
		"byte":          "bytes",
		"[]byte":        "bytes",
		"time.Time":     "string", // 时间类型映射为 string
		"time.Duration": "int64",
		"any":           "google.protobuf.Any",
		"interface{}":   "google.protobuf.Any",
	}

	if protoType, ok := typeMap[typeName]; ok {
		return protoType
	}

	// 自定义类型，保持原名
	return typeName
}

// isBuiltInType 判断是否为内置类型。
//
// 中文说明：
// - 用于区分内置类型和自定义结构体；
// - 自定义结构体需要生成对应的 message 定义。
func (g *Generator) isBuiltInType(typeName string) bool {
	builtIns := []string{
		"string", "int", "int32", "int64",
		"uint", "uint32", "uint64",
		"float32", "float64", "bool",
		"byte", "[]byte",
		"time.Time", "time.Duration",
		"any", "interface{}",
	}
	for _, t := range builtIns {
		if typeName == t {
			return true
		}
	}
	return false
}

// generateMethod 生成 service 方法。
//
// 中文说明：
// - 生成 rpc 方法定义；
// - 包含方法注释（从 Go 代码提取）；
// - 添加 HTTP 注解（如果配置）。
func (g *Generator) generateMethod(buf *bytes.Buffer, method contract.MethodDef, opts contract.ServiceToProtoOptions) {
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
func (g *Generator) formatProtoFile(path string) {
	// 使用 buf format 格式化（如果可用）
	cmd := exec.Command("buf", "format", "-w", path)
	cmd.Run() // 忽略错误
}

// GetConfig 获取当前配置。
func (g *Generator) GetConfig() *contract.ProtoGeneratorConfig {
	return g.cfg
}