package integration

type ServiceDef struct {
	Name     string
	Methods  []MethodDef
	Comments []string
}

type MethodDef struct {
	Name           string
	RequestType    *TypeDef
	ResponseType   *TypeDef
	Comments       []string
	HTTPRule       *HTTPRule
	RequestStream  bool
	ResponseStream bool
}

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

type EnumValue struct {
	Name     string
	Value    int32
	Comments []string
}

type ValidationRule struct {
	Rule    string
	Value   interface{}
	Message string
}

type RouteDef struct {
	Method       string
	Path         string
	HandlerName  string
	RequestType  *TypeDef
	ResponseType *TypeDef
	Comments     []string
	HandlerFile  string
}

type ImportDef struct {
	Path   string
	Public bool
	Weak   bool
}
