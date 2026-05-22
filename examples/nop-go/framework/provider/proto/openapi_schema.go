// Package proto provides OpenAPI Schema generation from proto messages.
// This file implements proto message parsing and schema conversion.
// Supports nested messages, repeated fields, enums, and special types.
//
// Proto 包提供从 proto message 生成 OpenAPI Schema 的实现。
// 本文件实现 proto message 解析和 schema 转换。
// 支持嵌套消息、repeated 字段、枚举和特殊类型。
package proto

import (
	"os"
	"regexp"
	"strings"
)

// ProtoMessage 表示从 proto 文件解析出的 message 定义。
//
// 中文说明：
// - 包含 message 名称和字段列表；
// - 用于生成 OpenAPI Components Schemas。
type ProtoMessage struct {
	// Name message 名称
	Name string

	// Fields 字段列表
	Fields []ProtoField

	// Description 注释描述
	Description string
}

// ProtoField 表示 proto message 的字段定义。
//
// 中文说明：
// - 包含字段名、类型、编号和修饰符；
// - 支持 optional/repeated/map 等修饰符。
type ProtoField struct {
	// Name 字段名
	Name string

	// Type 字段类型（string/int32/bool/MessageName等）
	Type string

	// Number 字段编号
	Number int

	// Repeated 是否是 repeated 字段（数组）
	Repeated bool

	// Optional 是否是 optional 字段
	Optional bool

	// MapKey map 类型的键类型（如果字段是 map）
	MapKey string

	// MapValue map 类型的值类型（如果字段是 map）
	MapValue string

	// Description 注释描述
	Description string
}

// ProtoEnum 表示从 proto 文件解析出的 enum 定义。
//
// 中文说明：
// - 包含 enum 名称和值列表；
// - 用于生成 OpenAPI enum schema。
type ProtoEnum struct {
	// Name enum 名称
	Name string

	// Values enum 值列表
	Values []ProtoEnumValue

	// Description 注释描述
	Description string
}

// ProtoEnumValue 表示 enum 的一个值。
type ProtoEnumValue struct {
	// Name 值名称
	Name string

	// Number 值编号
	Number int
}

// ProtoFileContent 表示解析后的 proto 文件内容。
//
// 中文说明：
// - 包含所有 message、enum 和 service 定义；
// - 用于生成完整的 OpenAPI 文档。
type ProtoFileContent struct {
	// Package proto 包名
	Package string

	// Messages message 定义列表
	Messages []ProtoMessage

	// Enums enum 定义列表
	Enums []ProtoEnum

	// Services service 定义列表
	Services []ProtoService
}

// parseProtoFileContent 解析 proto 文件内容，提取 message、enum 和 service。
//
// 中文说明：
// - 使用正则表达式解析 proto 文件；
// - 提取 message、enum 和 service 定义；
// - 返回完整的 ProtoFileContent 结构。
func parseProtoFileContent(protoFile string) (*ProtoFileContent, error) {
	content, err := readFileContent(protoFile)
	if err != nil {
		return nil, err
	}

	result := &ProtoFileContent{
		Messages: []ProtoMessage{},
		Enums:    []ProtoEnum{},
		Services: []ProtoService{},
	}

	contentStr := string(content)

	// 解析 package 声明
	result.Package = parseProtoPackage(contentStr)

	// 解析 message 定义
	result.Messages = parseProtoMessages(contentStr)

	// 解析 enum 定义
	result.Enums = parseProtoEnums(contentStr)

	// 解析 service 定义（复用现有逻辑）
	services, err := parseProtoFile(protoFile)
	if err != nil {
		return nil, err
	}
	result.Services = services

	return result, nil
}

// parseProtoPackage 解析 proto package 声明。
func parseProtoPackage(content string) string {
	packagePattern := regexp.MustCompile(`package\s+(\w+(?:\.\w+)*);`)
	match := packagePattern.FindStringSubmatch(content)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// parseProtoMessages 解析 proto 文件中的所有 message 定义。
//
// 中文说明：
// - 匹配 message 块并提取字段；
// - 支持嵌套 message；
// - 提取字段注释作为 description。
func parseProtoMessages(content string) []ProtoMessage {
	messages := []ProtoMessage{}

	// 匹配 message 块（支持嵌套）
	// 使用非贪婪匹配处理简单 message
	messagePattern := regexp.MustCompile(`message\s+(\w+)\s*\{([\s\S]*?)\n\}`)
	messageMatches := messagePattern.FindAllStringSubmatch(content, -1)

	for _, msgMatch := range messageMatches {
		messageName := msgMatch[1]
		messageBody := msgMatch[2]

		// 提取 message 注释（在 message 声明之前的注释）
		description := extractDescriptionBefore(content, "message "+messageName)

		fields := parseMessageFields(messageBody)

		messages = append(messages, ProtoMessage{
			Name:        messageName,
			Fields:      fields,
			Description: description,
		})
	}

	return messages
}

// parseMessageFields 解析 message body 中的字段定义。
//
// 中文说明：
// - 匯配字段行并提取类型、名称、编号；
// - 支持 repeated/optional/map 修饰符；
// - 提取字段注释作为 description。
func parseMessageFields(messageBody string) []ProtoField {
	fields := []ProtoField{}

	// 匹配字段定义
	// 格式：[optional] [repeated] type name = number;
	// 或：map<KeyType, ValueType> name = number;
	fieldPattern := regexp.MustCompile(`(//.*\n)?\s*(optional\s+)?(repeated\s+)?(?:map<(\w+),\s*(\w+)>)?(\w+)\s+(\w+)\s*=\s*(\d+);`)
	fieldMatches := fieldPattern.FindAllStringSubmatch(messageBody, -1)

	for _, fieldMatch := range fieldMatches {
		comment := strings.TrimSpace(fieldMatch[1])
		optional := strings.TrimSpace(fieldMatch[2]) != ""
		repeated := strings.TrimSpace(fieldMatch[3]) != ""
		mapKey := strings.TrimSpace(fieldMatch[4])
		mapValue := strings.TrimSpace(fieldMatch[5])
		fieldType := strings.TrimSpace(fieldMatch[6])
		fieldName := strings.TrimSpace(fieldMatch[7])
		fieldNumber := parseInt(fieldMatch[8])

		// 提取注释内容作为 description
		description := ""
		if comment != "" {
			description = strings.TrimPrefix(comment, "//")
			description = strings.TrimSpace(description)
		}

		// 处理 map 类型
		if mapKey != "" && mapValue != "" {
			fields = append(fields, ProtoField{
				Name:        fieldName,
				Type:        "map",
				Number:      fieldNumber,
				MapKey:      mapKey,
				MapValue:    mapValue,
				Description: description,
			})
			continue
		}

		fields = append(fields, ProtoField{
			Name:        fieldName,
			Type:        fieldType,
			Number:      fieldNumber,
			Repeated:    repeated,
			Optional:    optional,
			Description: description,
		})
	}

	return fields
}

// parseProtoEnums 解析 proto 文件中的所有 enum 定义。
//
// 中文说明：
// - 匹配 enum 块并提取值；
// - 提取 enum 注释作为 description。
func parseProtoEnums(content string) []ProtoEnum {
	enums := []ProtoEnum{}

	enumPattern := regexp.MustCompile(`enum\s+(\w+)\s*\{([\s\S]*?)\n\}`)
	enumMatches := enumPattern.FindAllStringSubmatch(content, -1)

	for _, enumMatch := range enumMatches {
		enumName := enumMatch[1]
		enumBody := enumMatch[2]

		description := extractDescriptionBefore(content, "enum "+enumName)

		values := parseEnumValues(enumBody)

		enums = append(enums, ProtoEnum{
			Name:        enumName,
			Values:      values,
			Description: description,
		})
	}

	return enums
}

// parseEnumValues 解析 enum body 中的值定义。
func parseEnumValues(enumBody string) []ProtoEnumValue {
	values := []ProtoEnumValue{}

	valuePattern := regexp.MustCompile(`(\w+)\s*=\s*(\d+);`)
	valueMatches := valuePattern.FindAllStringSubmatch(enumBody, -1)

	for _, valueMatch := range valueMatches {
		values = append(values, ProtoEnumValue{
			Name:   strings.TrimSpace(valueMatch[1]),
			Number: parseInt(valueMatch[2]),
		})
	}

	return values
}

// extractDescriptionBefore 提取声明之前的注释作为 description。
func extractDescriptionBefore(content, declaration string) string {
	// 查找声明位置
	declIndex := strings.Index(content, declaration)
	if declIndex == -1 {
		return ""
	}

	// 从声明位置向前查找注释
	beforeContent := content[:declIndex]

	// 查找最后一个注释块
	lines := strings.Split(beforeContent, "\n")
	var commentLines []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "//") {
			commentLines = append([]string{strings.TrimPrefix(line, "//")}, commentLines...)
		} else {
			break
		}
	}

	if len(commentLines) > 0 {
		return strings.TrimSpace(strings.Join(commentLines, " "))
	}

	return ""
}

// protoTypeToOpenAPIType 将 proto 类型映射到 OpenAPI 类型。
//
// 中文说明：
// - 基本类型映射（string/int32/bool 等）；
// - 特殊类型映射（google.protobuf.Timestamp 等）；
// - 返回 OpenAPI type 和 format。
func protoTypeToOpenAPIType(protoType string) (openapiType, format string) {
	// 基本类型映射
	typeMap := map[string][2]string{
		"string":   {"string", ""},
		"bool":     {"boolean", ""},
		"int32":    {"integer", "int32"},
		"int64":    {"integer", "int64"},
		"uint32":   {"integer", "int32"},
		"uint64":   {"integer", "int64"},
		"sint32":   {"integer", "int32"},
		"sint64":   {"integer", "int64"},
		"fixed32":  {"integer", "int32"},
		"fixed64":  {"integer", "int64"},
		"sfixed32": {"integer", "int32"},
		"sfixed64": {"integer", "int64"},
		"float":    {"number", "float"},
		"double":   {"number", "double"},
		"bytes":    {"string", "base64"},
	}

	// 特殊类型映射（google.protobuf）
	specialTypeMap := map[string][2]string{
		"google.protobuf.Timestamp":   {"string", "date-time"},
		"google.protobuf.Duration":    {"string", "duration"},
		"google.protobuf.Empty":       {"object", ""},
		"google.protobuf.Any":         {"object", ""},
		"google.protobuf.StringValue": {"string", ""},
		"google.protobuf.Int32Value":  {"integer", "int32"},
		"google.protobuf.Int64Value":  {"integer", "int64"},
		"google.protobuf.BoolValue":   {"boolean", ""},
		"google.protobuf.FloatValue":  {"number", "float"},
		"google.protobuf.DoubleValue": {"number", "double"},
		"google.protobuf.BytesValue":  {"string", "base64"},
	}

	// 检查基本类型
	if mapping, ok := typeMap[protoType]; ok {
		return mapping[0], mapping[1]
	}

	// 检查特殊类型
	if mapping, ok := specialTypeMap[protoType]; ok {
		return mapping[0], mapping[1]
	}

	// 自定义 message 类型，返回 $ref
	return "object", ""
}

// messageToSchema 将 ProtoMessage 转换为 OpenAPI Schema。
//
// 中文说明：
// - 遍历 message 字段生成 properties；
// - 处理 repeated 字段为数组；
// - 处理嵌套 message 为 $ref；
// - 处理 enum 为 enum 类型。
func messageToSchema(msg ProtoMessage, allMessages []ProtoMessage, allEnums []ProtoEnum) Schema {
	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
	}

	if msg.Description != "" {
		// Description 不在 Schema 结构中，但可以在生成时添加
	}

	for _, field := range msg.Fields {
		fieldSchema := fieldToSchema(field, allMessages, allEnums)
		schema.Properties[field.Name] = fieldSchema
	}

	return schema
}

// fieldToSchema 将 ProtoField 转换为 OpenAPI Schema。
//
// 中文说明：
// - 根据 field 类型生成对应的 schema；
// - repeated 字段包装为数组；
// - map 字段转换为 object with additionalProperties；
// - 自定义类型转换为 $ref。
func fieldToSchema(field ProtoField, allMessages []ProtoMessage, allEnums []ProtoEnum) Schema {
	var schema Schema

	// 处理 map 类型
	if field.Type == "map" && field.MapKey != "" && field.MapValue != "" {
		schema = Schema{
			Type: "object",
			AdditionalProperties: &Schema{
				Type: protoTypeToOpenAPISimple(field.MapValue, allMessages, allEnums),
			},
		}
		if field.MapValue != "" {
			// 如果 map value 是自定义类型，使用 $ref
			if isCustomType(field.MapValue, allMessages, allEnums) {
				schema.AdditionalProperties = &Schema{
					Ref: "#/components/schemas/" + field.MapValue,
				}
			}
		}
	} else {
		// 普通字段
		openapiType, format := protoTypeToOpenAPIType(field.Type)

		if openapiType == "object" && format == "" && isCustomType(field.Type, allMessages, allEnums) {
			// 自定义类型，使用 $ref
			schema = Schema{
				Ref: "#/components/schemas/" + field.Type,
			}
		} else if isEnumType(field.Type, allEnums) {
			// enum 类型
			enumSchema := enumToSchema(field.Type, allEnums)
			schema = enumSchema
		} else {
			// 基本类型
			schema = Schema{
				Type:   openapiType,
				Format: format,
			}
		}
	}

	// 处理 repeated（数组）
	if field.Repeated {
		schema = Schema{
			Type:  "array",
			Items: &schema,
		}
	}

	return schema
}

// protoTypeToOpenAPISimple 将 proto 类型转换为简单的 OpenAPI type 名称。
func protoTypeToOpenAPISimple(protoType string, allMessages []ProtoMessage, allEnums []ProtoEnum) string {
	openapiType, _ := protoTypeToOpenAPIType(protoType)
	return openapiType
}

// isCustomType 判断类型是否是自定义 message 类型。
func isCustomType(typeName string, allMessages []ProtoMessage, allEnums []ProtoEnum) bool {
	// 检查是否是已知 message
	for _, msg := range allMessages {
		if msg.Name == typeName {
			return true
		}
	}
	// 检查是否是已知 enum
	for _, enum := range allEnums {
		if enum.Name == typeName {
			return true
		}
	}
	// 检查是否是 google.protobuf 特殊类型
	if strings.HasPrefix(typeName, "google.protobuf.") {
		return false // 特殊类型已内置处理
	}
	// 基本类型
	basicTypes := []string{
		"string", "bool", "int32", "int64", "uint32", "uint64",
		"sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64",
		"float", "double", "bytes",
	}
	for _, t := range basicTypes {
		if t == typeName {
			return false
		}
	}
	// 未知类型，可能是外部引用，当作自定义类型处理
	return true
}

// isEnumType 判断类型是否是 enum 类型。
func isEnumType(typeName string, allEnums []ProtoEnum) bool {
	for _, enum := range allEnums {
		if enum.Name == typeName {
			return true
		}
	}
	return false
}

// enumToSchema 将 ProtoEnum 转换为 OpenAPI Schema。
func enumToSchema(enumName string, allEnums []ProtoEnum) Schema {
	for _, enum := range allEnums {
		if enum.Name == enumName {
			values := []string{}
			for _, v := range enum.Values {
				values = append(values, v.Name)
			}
			return Schema{
				Type: "string",
				Enum: values,
			}
		}
	}
	return Schema{Type: "string"}
}

// generateExampleFromSchema 从 Schema 生成示例值。
//
// 中文说明：
// - 根据类型生成合理的示例值；
// - string -> "example"；
// - integer -> 0；
// - boolean -> false；
// - object -> 递归生成；
// - array -> 空数组。
func generateExampleFromSchema(schema Schema) interface{} {
	switch schema.Type {
	case "string":
		if len(schema.Enum) > 0 {
			return schema.Enum[0]
		}
		if schema.Format == "date-time" {
			return "2026-01-01T00:00:00Z"
		}
		if schema.Format == "duration" {
			return "1s"
		}
		return "example"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return false
	case "array":
		if schema.Items != nil {
			return []interface{}{generateExampleFromSchema(*schema.Items)}
		}
		return []interface{}{}
	case "object":
		if schema.Ref != "" {
			// $ref 类型，返回空对象
			return map[string]interface{}{}
		}
		if schema.Properties != nil {
			example := map[string]interface{}{}
			for name, prop := range schema.Properties {
				example[name] = generateExampleFromSchema(prop)
			}
			return example
		}
		if schema.AdditionalProperties != nil {
			// map 类型
			return map[string]interface{}{
				"key": generateExampleFromSchema(*schema.AdditionalProperties),
			}
		}
		return map[string]interface{}{}
	default:
		return nil
	}
}

// parseInt 解析整数字符串。
func parseInt(s string) int {
	var result int
	for _, c := range strings.TrimSpace(s) {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// readFileContent 读取文件内容。
func readFileContent(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
