// Application scenarios:
// - Hold the shared AST-like models used during proto generation.
// - Standardize service, method, type, field, enum, route, and import descriptions.
// - Keep generator implementations and tooling aligned on one reusable model layer.
//
// 适用场景：
// - 承载 proto 生成过程中共享的类 AST 模型。
// - 统一 service、method、type、field、enum、route 和 import 描述。
// - 让生成器实现与工具链共享同一层可复用模型。
package integration

// ServiceDef describes one service definition.
//
// ServiceDef 描述一个服务定义。
type ServiceDef struct {
	Name     string
	Methods  []MethodDef
	Comments []string
}

// MethodDef describes one service method definition.
//
// MethodDef 描述一个服务方法定义。
type MethodDef struct {
	Name           string
	RequestType    *TypeDef
	ResponseType   *TypeDef
	Comments       []string
	HTTPRule       *HTTPRule
	RequestStream  bool
	ResponseStream bool
}

// TypeDef describes one type definition used in proto generation.
//
// TypeDef 描述 proto 生成过程中使用的类型定义。
type TypeDef struct {
	Name       string
	Package    string
	IsPointer  bool
	IsSlice    bool
	IsMap      bool
	MapKey     *TypeDef
	MapValue   *TypeDef
	Fields     []FieldDef
	Comments   []string
	IsEnum     bool
	EnumValues []EnumValue
}

// FieldDef describes one field definition.
//
// FieldDef 描述一个字段定义。
type FieldDef struct {
	Name            string
	JSONName        string
	ProtoName       string
	Type            *TypeDef
	Tag             string
	Remark          string
	Comments        []string
	ProtoNumber     int
	ValidationRules []ValidationRule
	DefaultValue    string
	IsOptional      bool
}

// EnumValue describes one enum value definition.
//
// EnumValue 描述一个枚举值定义。
type EnumValue struct {
	Name     string
	Value    int32
	Comments []string
}

// ValidationRule describes one generated validation rule.
//
// ValidationRule 描述一条生成出的校验规则。
type ValidationRule struct {
	Rule    string
	Value   interface{}
	Message string
}

// RouteDef describes one route-to-proto mapping definition.
//
// RouteDef 描述一条路由到 proto 的映射定义。
type RouteDef struct {
	Method       string
	Path         string
	HandlerName  string
	RequestType  *TypeDef
	ResponseType *TypeDef
	Comments     []string
	HandlerFile  string
}

// ImportDef describes one proto import declaration.
//
// ImportDef 描述一条 proto import 声明。
type ImportDef struct {
	Path   string
	Public bool
	Weak   bool
}
