// Package proto provides type resolver for cross-file/cross-package type resolution.
// Parses Go AST to resolve type dependencies across imports.
// Caches resolved types for performance.
//
// Proto 包提供类型解析器，用于跨文件/跨包类型解析。
// 解析 Go AST 以解决跨 import 的类型依赖。
// 缓存已解析的类型以提高性能。
package proto

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// TypeResolver 类型解析器。
//
// 中文说明：
// - 解析跨文件、跨包的类型定义；
// - 支持递归解析嵌套类型的依赖；
// - 缓存已解析的目录与文件以提高性能。
type TypeResolver struct {
	// importPaths import 搜索路径
	importPaths []string

	// cache 目录级类型缓存（dir -> typeName -> TypeDef）
	cache map[string]map[string]*integrationcontract.TypeDef
}

// NewTypeResolver 创建类型解析器。
//
// 中文说明：
// - importPaths 指定 import 的搜索路径；
// - 用于查找 import 的类型定义目录。
func NewTypeResolver(importPaths []string) *TypeResolver {
	return &TypeResolver{
		importPaths: importPaths,
		cache:       make(map[string]map[string]*integrationcontract.TypeDef),
	}
}

// ResolveImports 解析文件的 import 语句并返回“代码里使用的包名 -> 对应目录”。
//
// 中文说明：
// - 如果使用 import alias，则用 alias 作为 key；
// - 否则使用 import path 的最后一个目录名作为包名；
// - 返回值用于解析 SelectorExpr（如 time.Time / dto.Profile）。
func (r *TypeResolver) ResolveImports(f *ast.File, baseDir string) map[string]string {
	imports := make(map[string]string)

	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		dir := r.findImportDir(importPath, baseDir)
		if dir == "" {
			continue
		}

		packageName := filepath.Base(importPath)
		if imp.Name != nil && imp.Name.Name != "" && imp.Name.Name != "." && imp.Name.Name != "_" {
			packageName = imp.Name.Name
		}
		imports[packageName] = dir
	}

	return imports
}

// findImportDir 查找 import 对应的目录路径。
//
// 中文说明：
// - 在 baseDir 和 importPaths 下搜索对应目录；
// - 命中后返回目录路径，用于扫描其中的 Go 文件。
func (r *TypeResolver) findImportDir(importPath, baseDir string) string {
	candidates := make([]string, 0, len(r.importPaths)+1)
	if baseDir != "" {
		candidates = append(candidates, filepath.Join(baseDir, importPath))
	}
	for _, ip := range r.importPaths {
		candidates = append(candidates, filepath.Join(ip, importPath))
	}

	for _, dir := range candidates {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil || len(files) == 0 {
			continue
		}
		return dir
	}

	return ""
}

// ResolveType 解析目录中的单个类型定义。
//
// 中文说明：
// - 目录作为最小缓存单元，先扫描整个目录再取目标类型；
// - 返回该类型以及它依赖的自定义类型列表。
func (r *TypeResolver) ResolveType(typeName, dir string) (*integrationcontract.TypeDef, []string, error) {
	structDefs, err := r.loadStructDefsFromDir(dir)
	if err != nil {
		return nil, nil, err
	}

	typeDef, exists := structDefs[typeName]
	if !exists {
		return nil, nil, fmt.Errorf("type %s not found in %s", typeName, dir)
	}

	dependencies := make([]string, 0)
	for _, field := range typeDef.Fields {
		dependencies = append(dependencies, r.collectCustomTypeRefs(field.Type)...)
	}
	return typeDef, uniqueStrings(dependencies), nil
}

// ResolveTypesFromImport 从 import 目录解析多个类型定义。
//
// 中文说明：
// - 入参 importPath 可以是 import 路径，也可以直接是目录路径；
// - 返回目录中命中的类型定义集合。
func (r *TypeResolver) ResolveTypesFromImport(importPath string, typeNames []string, baseDir string) (map[string]*integrationcontract.TypeDef, error) {
	result := make(map[string]*integrationcontract.TypeDef)

	dir := importPath
	if !filepath.IsAbs(dir) {
		dir = r.findImportDir(importPath, baseDir)
	}
	if dir == "" {
		return nil, fmt.Errorf("import path not found: %s", importPath)
	}

	structDefs, err := r.loadStructDefsFromDir(dir)
	if err != nil {
		return nil, err
	}

	for _, typeName := range typeNames {
		if typeDef, exists := structDefs[typeName]; exists {
			result[typeName] = typeDef
		}
	}

	return result, nil
}

// ResolveAllTypes 从主文件出发解析所有初始类型的完整递归闭包。
//
// 中文说明：
// - 主文件同包类型由 generator 主链路负责收集；
// - 这里主要负责 import 引入的外部类型递归补全；
// - 会根据 TypeDef.Package 优先在对应 import 目录中查找。
func (r *TypeResolver) ResolveAllTypes(mainFile string, initialTypes []string) (map[string]*integrationcontract.TypeDef, error) {
	allTypes := make(map[string]*integrationcontract.TypeDef)
	pending := make([]string, 0, len(initialTypes))
	seen := make(map[string]bool)
	for _, name := range initialTypes {
		if strings.TrimSpace(name) == "" {
			continue
		}
		pending = append(pending, name)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, mainFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse main file: %w", err)
	}

	baseDir := filepath.Dir(mainFile)
	imports := r.ResolveImports(f, baseDir)

	for len(pending) > 0 {
		current := pending[0]
		pending = pending[1:]
		if seen[current] {
			continue
		}
		seen[current] = true

		packageName, shortName := splitQualifiedTypeName(current)
		if shortName == "" || r.isBuiltInType(current) || r.isBuiltInType(shortName) {
			continue
		}
		if _, exists := allTypes[shortName]; exists {
			continue
		}

		var resolved *integrationcontract.TypeDef

		// 优先按 selector 包名精确命中 import 目录。
		if packageName != "" {
			if dir, ok := imports[packageName]; ok {
				if typeDef, _, err := r.ResolveType(shortName, dir); err == nil {
					resolved = typeDef
				}
			}
		}

		// 非 selector 类型：遍历所有 import 目录兜底查找。
		if resolved == nil {
			for _, dir := range imports {
				typeDef, _, err := r.ResolveType(shortName, dir)
				if err == nil {
					resolved = typeDef
					break
				}
			}
		}

		if resolved == nil {
			continue
		}

		allTypes[shortName] = resolved
		for _, field := range resolved.Fields {
			for _, ref := range r.collectCustomTypeRefs(field.Type) {
				if !seen[ref] {
					pending = append(pending, ref)
				}
			}
		}
	}

	return allTypes, nil
}

// loadStructDefsFromDir 从目录缓存或加载所有结构体定义。
func (r *TypeResolver) loadStructDefsFromDir(dir string) (map[string]*integrationcontract.TypeDef, error) {
	if cached, ok := r.cache[dir]; ok {
		return cached, nil
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}

	structDefs := make(map[string]*integrationcontract.TypeDef)
	g := &Generator{}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		for typeName, typeDef := range g.extractStructDefs(f) {
			structDefs[typeName] = typeDef
		}
	}

	r.cache[dir] = structDefs
	return structDefs, nil
}

// collectCustomTypeRefs 收集 TypeDef 中所有需要继续解析的自定义类型引用。
func (r *TypeResolver) collectCustomTypeRefs(typeDef *integrationcontract.TypeDef) []string {
	if typeDef == nil {
		return nil
	}

	if typeDef.IsMap {
		refs := r.collectCustomTypeRefs(typeDef.MapKey)
		refs = append(refs, r.collectCustomTypeRefs(typeDef.MapValue)...)
		return uniqueStrings(refs)
	}

	fullName := fullTypeName(typeDef)
	if typeDef.Name != "" && !r.isBuiltInType(fullName) && !r.isBuiltInType(typeDef.Name) {
		return []string{fullName}
	}
	return nil
}

// isBuiltInType 判断是否为内置类型。
func (r *TypeResolver) isBuiltInType(typeName string) bool {
	_, ok := protoBuiltinType(typeName)
	return ok
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

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func splitQualifiedTypeName(typeName string) (string, string) {
	parts := strings.Split(typeName, ".")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", typeName
}

func fullTypeName(typeDef *integrationcontract.TypeDef) string {
	if typeDef == nil {
		return ""
	}
	if typeDef.Package != "" && typeDef.Name != "" {
		return typeDef.Package + "." + typeDef.Name
	}
	return typeDef.Name
}

func protoBuiltinType(typeName string) (string, bool) {
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
		"time.Time":     "string",
		"time.Duration": "int64",
		"any":           "google.protobuf.Any",
		"interface{}":   "google.protobuf.Any",
	}
	value, ok := typeMap[typeName]
	return value, ok
}
