package proto

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/ngq/gorp/framework/contract"
)

// EnumParser 枚举类型解析器。
//
// 中文说明：
// - 解析 Go 的 const/iota 枚举定义；
// - 支持识别枚举类型和值；
// - 生成 proto enum 定义。
type EnumParser struct {
	// enumPrefix 枚举类型前缀（用于识别枚举常量）
	enumPrefix string
}

// NewEnumParser 创建枚举解析器。
func NewEnumParser() *EnumParser {
	return &EnumParser{}
}

// EnumInfo 枚举信息。
type EnumInfo struct {
	// TypeName 枚举类型名称
	TypeName string

	// Values 枚举值列表
	Values []contract.EnumValue

	// Comments 注释
	Comments []string
}

// ParseFile 解析文件中的枚举定义。
//
// 中文说明：
// - 解析 Go 文件中的 const 块；
// - 识别 iota 枚举模式；
// - 返回枚举信息列表。
func (p *EnumParser) ParseFile(filePath string) ([]EnumInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	return p.ParseAST(f), nil
}

// ParseAST 从 AST 解析枚举定义。
func (p *EnumParser) ParseAST(f *ast.File) []EnumInfo {
	var enums []EnumInfo
	enumMap := make(map[string]*EnumInfo)

	// 遍历声明
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		// 检查是否有 iota
		if !p.hasIota(genDecl) {
			continue
		}

		// 解析 const 块
		p.parseConstBlock(genDecl, enumMap)
	}

	// 转换为列表
	for _, enum := range enumMap {
		enums = append(enums, *enum)
	}

	return enums
}

// hasIota 检查 const 块是否包含 iota。
func (p *EnumParser) hasIota(decl *ast.GenDecl) bool {
	for _, spec := range decl.Specs {
		vspec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for _, val := range vspec.Values {
			if p.containsIota(val) {
				return true
			}
		}
	}
	return false
}

// containsIota 检查表达式是否包含 iota。
func (p *EnumParser) containsIota(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "iota"
	case *ast.BinaryExpr:
		return p.containsIota(e.X) || p.containsIota(e.Y)
	case *ast.CallExpr:
		for _, arg := range e.Args {
			if p.containsIota(arg) {
				return true
			}
		}
	case *ast.ParenExpr:
		return p.containsIota(e.X)
	}
	return false
}

// parseConstBlock 解析 const 块。
//
// 中文说明：
// - 从 const 块中提取枚举值；
// - 根据常量名推断枚举类型；
// - 处理 iota 表达式计算枚举值。
func (p *EnumParser) parseConstBlock(decl *ast.GenDecl, enumMap map[string]*EnumInfo) {
	iotaValue := 0

	for _, spec := range decl.Specs {
		vspec, ok := spec.(*ast.ValueSpec)
		if !ok || len(vspec.Names) == 0 {
			continue
		}

		// 获取类型名
		typeName := p.extractTypeName(vspec)
		if typeName == "" {
			// 尝试从常量名推断类型
			typeName = p.inferTypeName(vspec.Names[0].Name)
		}

		if typeName == "" {
			iotaValue++
			continue
		}

		// 创建或获取枚举
		enum, exists := enumMap[typeName]
		if !exists {
			enum = &EnumInfo{
				TypeName: typeName,
			}
			enumMap[typeName] = enum
		}

		// 提取枚举值
		for i, name := range vspec.Names {
			var value int32
			if i < len(vspec.Values) {
				value = p.evalIotaExpr(vspec.Values[i], iotaValue)
			} else {
				value = int32(iotaValue)
			}

			// 提取注释
			var comments []string
			if vspec.Comment != nil {
				for _, c := range vspec.Comment.List {
					comments = append(comments, strings.TrimPrefix(c.Text, "//"))
				}
			}

			enum.Values = append(enum.Values, contract.EnumValue{
				Name:     p.toProtoEnumName(name.Name, typeName),
				Value:    value,
				Comments: comments,
			})
		}

		iotaValue++
	}
}

// extractTypeName 提取类型名。
func (p *EnumParser) extractTypeName(vspec *ast.ValueSpec) string {
	if vspec.Type == nil {
		return ""
	}

	switch t := vspec.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

// inferTypeName 从常量名推断类型名。
//
// 中文说明：
// - 例如 StatusActive -> Status；
// - 例如 UserRoleAdmin -> UserRole。
func (p *EnumParser) inferTypeName(constName string) string {
	// 尝试找到公共前缀
	// 例如: StatusActive, StatusInactive -> Status
	parts := strings.Split(constName, "_")
	if len(parts) >= 2 {
		// 假设第一部分是类型名
		return parts[0]
	}

	// 驼峰命名
	re := regexp.MustCompile(`([A-Z][a-z]+)`)
	matches := re.FindAllString(constName, -1)
	if len(matches) >= 2 {
		// 假设第一个词是类型名
		return matches[0]
	}

	return ""
}

// evalIotaExpr 计算 iota 表达式的值。
//
// 中文说明：
// - 支持基本的 iota 算术表达式；
// - 如 iota + 1, 1 << iota 等。
func (p *EnumParser) evalIotaExpr(expr ast.Expr, iotaValue int) int32 {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "iota" {
			return int32(iotaValue)
		}
		return 0
	case *ast.BasicLit:
		val, _ := strconv.Atoi(e.Value)
		return int32(val)
	case *ast.BinaryExpr:
		x := p.evalIotaExpr(e.X, iotaValue)
		y := p.evalIotaExpr(e.Y, iotaValue)
		switch e.Op {
		case token.ADD:
			return x + y
		case token.SUB:
			return x - y
		case token.MUL:
			return x * y
		case token.QUO:
			if y != 0 {
				return x / y
			}
		case token.REM:
			if y != 0 {
				return x % y
			}
		case token.SHL:
			return x << y
		case token.SHR:
			return x >> y
		case token.AND:
			return x & y
		case token.OR:
			return x | y
		case token.XOR:
			return x ^ y
		}
	case *ast.ParenExpr:
		return p.evalIotaExpr(e.X, iotaValue)
	}
	return int32(iotaValue)
}

// toProtoEnumName 将 Go 常量名转换为 proto 枚举名。
//
// 中文说明：
// - 移除类型前缀；
// - 转换为大写蛇形命名。
// 例如: StatusActive -> STATUS_ACTIVE 或 ACTIVE
func (p *EnumParser) toProtoEnumName(constName, typeName string) string {
	// 移除类型前缀
	name := constName

	// 尝试移除前缀（驼峰）
	if strings.HasPrefix(constName, typeName) {
		name = strings.TrimPrefix(constName, typeName)
	}

	// 尝试移除前缀（蛇形）
	prefix := strings.ToUpper(typeName) + "_"
	if strings.HasPrefix(strings.ToUpper(constName), prefix) {
		name = constName[len(prefix):]
	}

	// 转换为大写
	return strings.ToUpper(name)
}

// GenerateProtoEnum 生成 proto enum 定义。
//
// 中文说明：
// - 根据枚举信息生成 proto enum 语法；
// - 包含枚举值和注释。
func GenerateProtoEnum(enum EnumInfo) string {
	var buf strings.Builder

	// 写入注释
	if len(enum.Comments) > 0 {
		for _, c := range enum.Comments {
			buf.WriteString(fmt.Sprintf("// %s\n", strings.TrimSpace(c)))
		}
	}

	buf.WriteString(fmt.Sprintf("enum %s {\n", enum.TypeName))

	// 添加 UNKNOWN = 0 作为默认值
	buf.WriteString(fmt.Sprintf("  %s_UNKNOWN = 0;\n", enum.TypeName))

	// 添加枚举值
	for _, val := range enum.Values {
		if len(val.Comments) > 0 {
			for _, c := range val.Comments {
				buf.WriteString(fmt.Sprintf("  // %s\n", strings.TrimSpace(c)))
			}
		}
		buf.WriteString(fmt.Sprintf("  %s = %d;\n", val.Name, val.Value))
	}

	buf.WriteString("}\n\n")

	return buf.String()
}

// DetectEnumType 检测字段类型是否是枚举类型。
//
// 中文说明：
// - 根据已解析的枚举列表检测类型；
// - 返回枚举类型名或空字符串。
func DetectEnumType(typeName string, enums []EnumInfo) string {
	for _, enum := range enums {
		if enum.TypeName == typeName {
			return typeName
		}
		// 检查是否是枚举类型的别名
		if strings.HasSuffix(typeName, enum.TypeName) {
			return enum.TypeName
		}
	}
	return ""
}