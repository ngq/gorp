package proto

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/ngq/gorp/framework/contract"
)

// HandlerParser Handler 解析器。
//
// 中文说明：
// - 解析 Gin Handler 函数定义；
// - 提取请求/响应类型信息；
// - 支持从 Handler 所在文件解析相关类型。
type HandlerParser struct {
	// generator Generator 实例（复用类型解析方法）
	generator *Generator
}

// NewHandlerParser 创建 Handler 解析器。
func NewHandlerParser() *HandlerParser {
	return &HandlerParser{
		generator: &Generator{},
	}
}

// HandlerTypeInfo Handler 类型信息。
type HandlerTypeInfo struct {
	// HandlerName Handler 函数名
	HandlerName string

	// RequestType 请求类型
	RequestType *contract.TypeDef

	// ResponseType 响应类型
	ResponseType *contract.TypeDef

	// RequestTypeName 请求类型名称
	RequestTypeName string

	// ResponseTypeName 响应类型名称
	ResponseTypeName string
}

// ParseHandlerFile 解析 Handler 文件，提取所有 Handler 的类型信息。
//
// 中文说明：
// - 解析指定文件中的所有 Handler 函数；
// - 返回 Handler 名称到类型信息的映射。
func (p *HandlerParser) ParseHandlerFile(filePath string) (map[string]*HandlerTypeInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse handler file: %w", err)
	}

	handlers := make(map[string]*HandlerTypeInfo)

	// 遍历 AST 查找 Handler 函数
	for _, decl := range f.Decls {
		fnDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// 检查是否是 Gin Handler（参数包含 *gin.Context）
		if !p.isGinHandler(fnDecl) {
			continue
		}

		// 提取 Handler 类型信息
		info := p.extractHandlerTypeInfo(fnDecl)
		if info != nil {
			handlers[fnDecl.Name.Name] = info
		}
	}

	return handlers, nil
}

// ParseHandlers 解析指定的 Handler 函数列表。
//
// 中文说明：
// - 解析指定文件中的特定 Handler 函数；
// - 只返回指定的 Handler 类型信息。
func (p *HandlerParser) ParseHandlers(filePath string, handlerNames []string) (map[string]*HandlerTypeInfo, error) {
	allHandlers, err := p.ParseHandlerFile(filePath)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*HandlerTypeInfo)
	for _, name := range handlerNames {
		if info, exists := allHandlers[name]; exists {
			result[name] = info
		}
	}

	return result, nil
}

// isGinHandler 判断函数是否是 Gin Handler。
//
// 中文说明：
// - Gin Handler 的第一个参数是 *gin.Context；
// - 方法必须绑定到结构体（有 Receiver）。
func (p *HandlerParser) isGinHandler(fnDecl *ast.FuncDecl) bool {
	if fnDecl.Type.Params == nil || len(fnDecl.Type.Params.List) == 0 {
		return false
	}

	// 检查第一个参数是否是 *gin.Context
	firstParam := fnDecl.Type.Params.List[0]
	return p.isGinContext(firstParam.Type)
}

// isGinContext 判断类型是否是 *gin.Context。
func (p *HandlerParser) isGinContext(expr ast.Expr) bool {
	// 处理 *gin.Context
	starExpr, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}

	selExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	return selExpr.Sel.Name == "Context"
}

// extractHandlerTypeInfo 提取 Handler 的类型信息。
//
// 中文说明：
// - 从 Handler 函数签名提取请求/响应类型；
// - 请求类型通常是 gin.Context 之后的参数；
// - 响应类型通常从函数体内推断（分析 c.JSON 调用）。
func (p *HandlerParser) extractHandlerTypeInfo(fnDecl *ast.FuncDecl) *HandlerTypeInfo {
	info := &HandlerTypeInfo{
		HandlerName: fnDecl.Name.Name,
	}

	// 提取请求类型
	info.RequestType, info.RequestTypeName = p.extractRequestType(fnDecl)

	// 提取响应类型
	info.ResponseType, info.ResponseTypeName = p.extractResponseType(fnDecl)

	return info
}

// extractRequestType 提取请求类型。
//
// 中文说明：
// - Gin Handler 的请求通常通过以下方式获取：
//   1. ShouldBindJSON / BindJSON 等方法绑定到结构体
//   2. Query / Param 等方法获取参数
// - 这里主要分析 ShouldBindJSON 的目标类型。
func (p *HandlerParser) extractRequestType(fnDecl *ast.FuncDecl) (*contract.TypeDef, string) {
	var requestType *contract.TypeDef
	var requestTypeName string

	// 分析函数体，查找 ShouldBindJSON/Bind 等调用
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 检查是否是 c.ShouldBindJSON(&req) 或类似调用
		methodName := p.getMethodName(callExpr)
		if p.isBindMethod(methodName) {
			// 提取绑定的变量类型
			if len(callExpr.Args) > 0 {
				typeName := p.extractVarType(callExpr.Args[0])
				if typeName != "" {
					requestTypeName = typeName
					requestType = &contract.TypeDef{Name: typeName}
					return false
				}
			}
		}

		return true
	})

	return requestType, requestTypeName
}

// extractResponseType 提取响应类型。
//
// 中文说明：
// - 分析 c.JSON / c.XML 等响应调用；
// - 提取响应数据的类型。
func (p *HandlerParser) extractResponseType(fnDecl *ast.FuncDecl) (*contract.TypeDef, string) {
	var responseType *contract.TypeDef
	var responseTypeName string

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 检查是否是 c.JSON(statusCode, data) 调用
		methodName := p.getMethodName(callExpr)
		if methodName == "JSON" || methodName == "XML" {
			// 第二个参数是响应数据
			if len(callExpr.Args) >= 2 {
				typeName := p.extractResponseTypeFromArg(callExpr.Args[1])
				if typeName != "" {
					responseTypeName = typeName
					responseType = &contract.TypeDef{Name: typeName}
					return false
				}
			}
		}

		return true
	})

	return responseType, responseTypeName
}

// getMethodName 获取方法调用名称。
func (p *HandlerParser) getMethodName(callExpr *ast.CallExpr) string {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	return selExpr.Sel.Name
}

// isBindMethod 判断是否是绑定方法。
func (p *HandlerParser) isBindMethod(methodName string) bool {
	bindMethods := []string{
		"ShouldBindJSON", "BindJSON",
		"ShouldBind", "Bind",
		"ShouldBindXML", "BindXML",
		"ShouldBindQuery", "BindQuery",
		"ShouldBindUri", "BindUri",
		"ShouldBindHeader", "BindHeader",
	}
	for _, m := range bindMethods {
		if methodName == m {
			return true
		}
	}
	return false
}

// extractVarType 从变量表达式中提取类型名称。
//
// 中文说明：
// - 处理 &req 形式的表达式；
// - 返回 req 的类型名称。
func (p *HandlerParser) extractVarType(expr ast.Expr) string {
	// 处理 &req
	unaryExpr, ok := expr.(*ast.UnaryExpr)
	if ok && unaryExpr.Op == '&' {
		// 获取变量名，然后在函数中查找变量声明
		if ident, ok := unaryExpr.X.(*ast.Ident); ok {
			// 这里简化处理，返回变量名（实际应该查找变量声明）
			return ident.Name + "Request"
		}
	}
	return ""
}

// extractResponseTypeFromArg 从响应参数中提取类型名称。
func (p *HandlerParser) extractResponseTypeFromArg(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// 单个变量，如 c.JSON(200, user)
		return e.Name + "Response"
	case *ast.CompositeLit:
		// 复合字面量，如 c.JSON(200, UserResponse{...})
		return p.extractTypeNameFromComposite(e)
	case *ast.CallExpr:
		// 函数调用，如 c.JSON(200, ToResponse(user))
		if selExpr, ok := e.Fun.(*ast.SelectorExpr); ok {
			// 返回函数名作为类型提示
			return selExpr.Sel.Name + "Result"
		}
	}
	return ""
}

// extractTypeNameFromComposite 从复合字面量中提取类型名称。
func (p *HandlerParser) extractTypeNameFromComposite(lit *ast.CompositeLit) string {
	switch t := lit.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

// InferRequestTypeFromHandler 从 Handler 函数推断请求类型。
//
// 中文说明：
// - 分析 Handler 函数体；
// - 查找 ShouldBindJSON 等调用；
// - 提取绑定的结构体类型。
func (p *HandlerParser) InferRequestTypeFromHandler(fnDecl *ast.FuncDecl, structDefs map[string][]contract.FieldDef) *contract.TypeDef {
	info := &HandlerTypeInfo{}
	requestType, typeName := p.extractRequestType(fnDecl)
	info.RequestType = requestType
	info.RequestTypeName = typeName

	if typeName == "" {
		return nil
	}

	// 如果类型在 structDefs 中，返回完整定义
	if fields, exists := structDefs[typeName]; exists {
		return &contract.TypeDef{
			Name:   typeName,
			Fields: fields,
		}
	}

	return requestType
}

// ExtractPathParams 从路径中提取路径参数。
//
// 中文说明：
// - 解析 Gin 路径参数格式（如 /users/:id）；
// - 返回参数名列表。
func ExtractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			// 处理正则约束（如 :id<int>）
			if idx := strings.Index(paramName, "<"); idx > 0 {
				paramName = paramName[:idx]
			}
			// 处理 * 通配符
			paramName = strings.TrimSuffix(paramName, "*")
			params = append(params, paramName)
		}
	}

	return params
}

// GenerateRequestMessageFromHandler 根据 Handler 信息生成请求消息。
//
// 中文说明：
// - 从路径参数和 Handler 分析结果生成请求 message；
// - 包含路径参数和请求体字段。
func GenerateRequestMessageFromHandler(handlerName string, pathParams []string, requestType *contract.TypeDef) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("message %sRequest {\n", handlerName))

	fieldNum := 1

	// 添加路径参数字段
	for _, param := range pathParams {
		buf.WriteString(fmt.Sprintf("  string %s = %d;\n", param, fieldNum))
		fieldNum++
	}

	// 添加请求体字段
	if requestType != nil && len(requestType.Fields) > 0 {
		for _, field := range requestType.Fields {
			buf.WriteString(fmt.Sprintf("  string %s = %d; // from request body\n", field.ProtoName, fieldNum))
			fieldNum++
		}
	}

	buf.WriteString("}\n\n")
	return buf.String()
}