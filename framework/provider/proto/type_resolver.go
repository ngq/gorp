package proto

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/ngq/gorp/framework/contract"
)

// TypeResolver 类型解析器。
//
// 中文说明：
// - 解析跨文件、跨包的类型定义；
// - 支持递归解析嵌套类型的依赖；
// - 缓存已解析的类型以提高性能。
type TypeResolver struct {
	// importPaths import 搜索路径
	importPaths []string

	// cache 类型定义缓存（filePath -> typeName -> TypeDef）
	cache map[string]map[string]*contract.TypeDef

	// resolvedFiles 已解析的文件路径
	resolvedFiles map[string]bool
}

// NewTypeResolver 创建类型解析器。
//
// 中文说明：
// - importPaths 指定 import 的搜索路径；
// - 用于查找 import 的类型定义文件。
func NewTypeResolver(importPaths []string) *TypeResolver {
	return &TypeResolver{
		importPaths:    importPaths,
		cache:          make(map[string]map[string]*contract.TypeDef),
		resolvedFiles:  make(map[string]bool),
	}
}

// ResolveImports 解析文件的 import 语句。
//
// 中文说明：
// - 返回 import 路径到可能文件路径的映射；
// - 用于后续解析跨包类型。
func (r *TypeResolver) ResolveImports(f *ast.File, baseDir string) map[string]string {
	imports := make(map[string]string)

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		// 尝试在 importPaths 中查找文件
		filePath := r.findImportFile(path, baseDir)
		if filePath != "" {
			imports[path] = filePath
		}
	}

	return imports
}

// findImportFile 查找 import 对应的文件路径。
//
// 中文说明：
// - 在 importPaths 中搜索对应的 Go 文件；
// - 支持相对路径和绝对路径。
func (r *TypeResolver) findImportFile(importPath, baseDir string) string {
	// 构建可能的文件路径
	candidates := []string{}

	// 1. 相对于 baseDir
	if baseDir != "" {
		candidates = append(candidates, filepath.Join(baseDir, importPath))
	}

	// 2. 在 importPaths 中搜索
	for _, ip := range r.importPaths {
		candidates = append(candidates, filepath.Join(ip, importPath))
	}

	// 检查每个候选路径
	for _, dir := range candidates {
		// 尝试查找 Go 文件
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err == nil && len(files) > 0 {
			return dir
		}
	}

	return ""
}

// ResolveType 解析类型定义。
//
// 中文说明：
// - 从指定文件中解析类型定义；
// - 支持递归解析嵌套类型；
// - 返回类型定义和所有依赖的类型。
func (r *TypeResolver) ResolveType(typeName, filePath string) (*contract.TypeDef, []string, error) {
	// 检查缓存
	if cached, ok := r.cache[filePath]; ok {
		if typeDef, exists := cached[typeName]; exists {
			return typeDef, nil, nil
		}
	}

	// 解析文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parse file %s: %w", filePath, err)
	}

	// 提取结构体定义
	structDefs := r.extractStructDefs(f)

	// 更新缓存
	r.cache[filePath] = structDefs

	// 查找目标类型
	typeDef, exists := structDefs[typeName]
	if !exists {
		return nil, nil, fmt.Errorf("type %s not found in %s", typeName, filePath)
	}

	// 收集依赖的类型
	var dependencies []string
	for _, field := range typeDef.Fields {
		if field.Type != nil && field.Type.Name != "" && !r.isBuiltInType(field.Type.Name) {
			dependencies = append(dependencies, field.Type.Name)
		}
	}

	return typeDef, dependencies, nil
}

// ResolveTypesFromImport 从 import 解析多个类型定义。
//
// 中文说明：
// - 解析指定 import 路径下的类型；
// - 返回所有解析到的类型定义。
func (r *TypeResolver) ResolveTypesFromImport(importPath string, typeNames []string, baseDir string) (map[string]*contract.TypeDef, error) {
	result := make(map[string]*contract.TypeDef)

	// 查找 import 对应的目录
	dir := r.findImportFile(importPath, baseDir)
	if dir == "" {
		return nil, fmt.Errorf("import path not found: %s", importPath)
	}

	// 遍历目录下的 Go 文件
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("glob files: %w", err)
	}

	for _, file := range files {
		// 跳过测试文件
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		// 解析文件
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// 提取结构体定义
		structDefs := r.extractStructDefs(f)

		// 收集所需的类型
		for _, typeName := range typeNames {
			if typeDef, exists := structDefs[typeName]; exists {
				result[typeName] = typeDef
			}
		}

		// 如果所有类型都已找到，提前返回
		if len(result) == len(typeNames) {
			break
		}
	}

	return result, nil
}

// extractStructDefs 从文件提取结构体定义。
//
// 中文说明：
// - 遍历 AST 提取所有结构体定义；
// - 返回类型名到类型定义的映射（包含字段和注释）。
func (r *TypeResolver) extractStructDefs(f *ast.File) map[string]*contract.TypeDef {
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

			fields := r.extractStructFields(structType)
			structDefs[typeSpec.Name.Name] = &contract.TypeDef{
				Name:     typeSpec.Name.Name,
				Fields:   fields,
				Comments: comments,
			}
		}
	}

	return structDefs
}

// extractStructFields 提取结构体字段。
func (r *TypeResolver) extractStructFields(structType *ast.StructType) []contract.FieldDef {
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
			// 嵌入字段，跳过
			continue
		}

		// 提取字段类型
		g := &Generator{}
		typeDef := g.extractType(field.Type)

		// 提取 JSON tag
		var jsonName, protoName string
		if field.Tag != nil {
			tagValue := strings.Trim(field.Tag.Value, "`")
			jsonName = g.extractTagValue(tagValue, "json")
			protoName = g.extractTagValue(tagValue, "proto")
		}

		if jsonName == "" {
			jsonName = fieldName
		}
		if protoName == "" {
			protoName = g.toSnakeCase(jsonName)
		}

		// 提取注释
		var comments []string
		if field.Comment != nil {
			for _, c := range field.Comment.List {
				comments = append(comments, strings.TrimPrefix(c.Text, "//"))
			}
		}

		fields = append(fields, contract.FieldDef{
			Name:      fieldName,
			JSONName:  jsonName,
			ProtoName: protoName,
			Type:      typeDef,
			Comments:  comments,
		})
	}

	return fields
}

// isBuiltInType 判断是否为内置类型。
func (r *TypeResolver) isBuiltInType(typeName string) bool {
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

// ResolveAllTypes 解析文件中引用的所有类型。
//
// 中文说明：
// - 从主文件开始，递归解析所有引用的类型；
// - 返回所有解析到的类型定义；
// - 支持跨 import 的类型解析。
func (r *TypeResolver) ResolveAllTypes(mainFile string, initialTypes []string) (map[string]*contract.TypeDef, error) {
	allTypes := make(map[string]*contract.TypeDef)
	toResolve := make(map[string]bool)

	// 初始化待解析类型
	for _, t := range initialTypes {
		toResolve[t] = true
	}

	// 解析主文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, mainFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse main file: %w", err)
	}

	// 提取主文件的结构体定义
	structDefs := r.extractStructDefs(f)
	for typeName, typeDef := range structDefs {
		allTypes[typeName] = typeDef
		delete(toResolve, typeName)
	}

	// 解析 import 语句
	baseDir := filepath.Dir(mainFile)
	imports := r.ResolveImports(f, baseDir)

	// 递归解析未找到的类型
	for len(toResolve) > 0 {
		found := false
		for typeName := range toResolve {
			// 检查是否已解析
			if _, exists := allTypes[typeName]; exists {
				delete(toResolve, typeName)
				found = true
				continue
			}

			// 尝试从 import 中解析
			for importPath, dir := range imports {
				types, err := r.ResolveTypesFromImport(importPath, []string{typeName}, filepath.Dir(dir))
				if err == nil {
					for t, typeDef := range types {
						allTypes[t] = typeDef
						delete(toResolve, t)
						found = true
					}
				}
			}
		}

		// 如果没有找到任何类型，退出循环
		if !found {
			break
		}
	}

	return allTypes, nil
}

// FindGoFileInDir 在目录中查找包含指定类型的 Go 文件。
//
// 中文说明：
// - 遍历目录下的 Go 文件；
// - 查找包含指定类型定义的文件。
func FindGoFileInDir(dir, typeName string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}

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

				if typeSpec.Name.Name == typeName {
					return file, nil
				}
			}
		}
	}

	return "", os.ErrNotExist
}