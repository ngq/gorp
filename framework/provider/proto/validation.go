package proto

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ValidationConverter 验证规则转换器。
//
// 中文说明：
// - 将 Go 的 validate/binding tag 转换为 proto 验证规则；
// - 支持常见的验证规则映射；
// - 生成 protoc-gen-validate 兼容的注解。
type ValidationConverter struct {
	// style 验证风格：validate.rules 或 validate.rules
	style string
}

// NewValidationConverter 创建验证规则转换器。
func NewValidationConverter() *ValidationConverter {
	return &ValidationConverter{
		style: "validate.rules",
	}
}

// ConvertTag 转换验证 tag 为 proto 验证规则。
//
// 中文说明：
// - 解析 Go struct tag 中的验证规则；
// - 转换为 proto 验证注解格式；
// - 支持 binding 和 validate 两种 tag 格式。
func (c *ValidationConverter) ConvertTag(tag string, fieldType string) []integrationcontract.ValidationRule {
	var rules []integrationcontract.ValidationRule

	// 解析 binding tag
	bindingRules := c.parseBindingTag(tag)
	rules = append(rules, bindingRules...)

	// 解析 validate tag
	validateRules := c.parseValidateTag(tag)
	rules = append(rules, validateRules...)

	return rules
}

// parseBindingTag 解析 binding tag。
//
// 中文说明：
// - Gin 的 binding tag 格式：`binding:"required,min=6,max=20"`;
// - 支持常见规则：required, min, max, email, url 等。
func (c *ValidationConverter) parseBindingTag(tag string) []integrationcontract.ValidationRule {
	var rules []integrationcontract.ValidationRule

	bindingValue := c.extractTagValue(tag, "binding")
	if bindingValue == "" {
		return rules
	}

	// 分割规则
	parts := strings.Split(bindingValue, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		rule := c.parseSingleBindingRule(part)
		if rule.Rule != "" {
			rules = append(rules, rule)
		}
	}

	return rules
}

// parseSingleBindingRule 解析单个 binding 规则。
func (c *ValidationConverter) parseSingleBindingRule(part string) integrationcontract.ValidationRule {
	// 处理 key=value 格式
	if strings.Contains(part, "=") {
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "min":
			if num, err := strconv.Atoi(value); err == nil {
				return integrationcontract.ValidationRule{
					Rule:  "min_len",
					Value: num,
				}
			}
		case "max":
			if num, err := strconv.Atoi(value); err == nil {
				return integrationcontract.ValidationRule{
					Rule:  "max_len",
					Value: num,
				}
			}
		case "len":
			if num, err := strconv.Atoi(value); err == nil {
				return integrationcontract.ValidationRule{
					Rule:  "len",
					Value: num,
				}
			}
		case "eqfield":
			return integrationcontract.ValidationRule{
				Rule:  "eq_field",
				Value: value,
			}
		}
	}

	// 处理无值规则
	switch part {
	case "required":
		return integrationcontract.ValidationRule{Rule: "required", Value: true}
	case "email":
		return integrationcontract.ValidationRule{Rule: "email", Value: true}
	case "url", "uri":
		return integrationcontract.ValidationRule{Rule: "uri", Value: true}
	case "uuid":
		return integrationcontract.ValidationRule{Rule: "uuid", Value: true}
	case "alphanum":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[a-zA-Z0-9]+$"}
	case "numeric":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[0-9]+$"}
	case "alpha":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[a-zA-Z]+$"}
	}

	return integrationcontract.ValidationRule{}
}

// parseValidateTag 解析 validate tag。
//
// 中文说明：
// - 标准的 validator/v10 格式：`validate:"required,min=6,max=20"`;
// - 支持更丰富的规则。
func (c *ValidationConverter) parseValidateTag(tag string) []integrationcontract.ValidationRule {
	var rules []integrationcontract.ValidationRule

	validateValue := c.extractTagValue(tag, "validate")
	if validateValue == "" {
		return rules
	}

	parts := strings.Split(validateValue, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "omitempty" {
			continue
		}

		rule := c.parseSingleValidateRule(part)
		if rule.Rule != "" {
			rules = append(rules, rule)
		}
	}

	return rules
}

// parseSingleValidateRule 解析单个 validate 规则。
func (c *ValidationConverter) parseSingleValidateRule(part string) integrationcontract.ValidationRule {
	if strings.Contains(part, "=") {
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "min":
			if num, err := strconv.Atoi(value); err == nil {
				return integrationcontract.ValidationRule{Rule: "min_len", Value: num}
			}
		case "max":
			if num, err := strconv.Atoi(value); err == nil {
				return integrationcontract.ValidationRule{Rule: "max_len", Value: num}
			}
		case "len":
			if num, err := strconv.Atoi(value); err == nil {
				return integrationcontract.ValidationRule{Rule: "len", Value: num}
			}
		case "gte", "gtefield":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				return integrationcontract.ValidationRule{Rule: "gte", Value: num}
			}
		case "lte", "ltefield":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				return integrationcontract.ValidationRule{Rule: "lte", Value: num}
			}
		case "gt":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				return integrationcontract.ValidationRule{Rule: "gt", Value: num}
			}
		case "lt":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				return integrationcontract.ValidationRule{Rule: "lt", Value: num}
			}
		case "oneof":
			return integrationcontract.ValidationRule{Rule: "in", Value: strings.Split(value, " ")}
		case "excludesall":
			return integrationcontract.ValidationRule{Rule: "not_in", Value: strings.Split(value, " ")}
		case "unique":
			return integrationcontract.ValidationRule{Rule: "unique", Value: true}
		case "contains":
			return integrationcontract.ValidationRule{Rule: "contains", Value: value}
		case "excludes":
			return integrationcontract.ValidationRule{Rule: "not_contains", Value: value}
		case "startswith":
			return integrationcontract.ValidationRule{Rule: "prefix", Value: value}
		case "endswith":
			return integrationcontract.ValidationRule{Rule: "suffix", Value: value}
		case "datetime":
			return integrationcontract.ValidationRule{Rule: "pattern", Value: value}
		}
	}

	// 无值规则
	switch part {
	case "required":
		return integrationcontract.ValidationRule{Rule: "required", Value: true}
	case "email":
		return integrationcontract.ValidationRule{Rule: "email", Value: true}
	case "url":
		return integrationcontract.ValidationRule{Rule: "uri", Value: true}
	case "uuid":
		return integrationcontract.ValidationRule{Rule: "uuid", Value: true}
	case "uuid4":
		return integrationcontract.ValidationRule{Rule: "uuid4", Value: true}
	case "uuid5":
		return integrationcontract.ValidationRule{Rule: "uuid5", Value: true}
	case "ascii":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[\x00-\x7F]*$"}
	case "printascii":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[\x20-\x7E]*$"}
	case "alphanum":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[a-zA-Z0-9]+$"}
	case "alpha":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[a-zA-Z]+$"}
	case "numeric":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[0-9]+$"}
	case "number":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[0-9]+\\.?[0-9]*$"}
	case "hexadecimal":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[0-9a-fA-F]+$"}
	case "hexcolor":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}
	case "rgb":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^rgb\\(\\s*(\\d{1,3})\\s*,\\s*(\\d{1,3})\\s*,\\s*(\\d{1,3})\\s*\\)$"}
	case "rgba":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^rgba\\(\\s*(\\d{1,3})\\s*,\\s*(\\d{1,3})\\s*,\\s*(\\d{1,3})\\s*,\\s*([01]\\.?\\d*?)\\s*\\)$"}
	case "hsl":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^hsl\\(\\s*(\\d{1,3})\\s*,\\s*(\\d{1,3})%\\s*,\\s*(\\d{1,3})%\\s*\\)$"}
	case "hsla":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^hsla\\(\\s*(\\d{1,3})\\s*,\\s*(\\d{1,3})%\\s*,\\s*(\\d{1,3})%\\s*,\\s*([01]\\.?\\d*?)\\s*\\)$"}
	case "json":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^\\s*([\\[\\{].*[\\]\\}])\\s*$"}
	case "file":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^[^\\\\/:*?\"<>|]+$"}
	case "base64":
		return integrationcontract.ValidationRule{Rule: "pattern", Value: "^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$"}
	}

	return integrationcontract.ValidationRule{}
}

// extractTagValue 提取 tag 值。
func (c *ValidationConverter) extractTagValue(tagStr, key string) string {
	// 格式: key:"value" 或 key:"value,options"
	re := regexp.MustCompile(key + `:"([^"]*)"`)
	matches := re.FindStringSubmatch(tagStr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GenerateProtoValidation 生成 proto 验证注解。
//
// 中文说明：
// - 根据验证规则生成 protoc-gen-validate 格式的注解；
// - 例如：`[(validate.rules).string.min_len = 6]`。
func GenerateProtoValidation(fieldType string, rules []integrationcontract.ValidationRule) string {
	if len(rules) == 0 {
		return ""
	}

	var parts []string

	for _, rule := range rules {
		protoRule := c.ruleToProtoValidation(fieldType, rule)
		if protoRule != "" {
			parts = append(parts, protoRule)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("[(%s) = {%s}]", "validate.rules", strings.Join(parts, ", "))
}

// ruleToProtoValidation 将单个规则转换为 proto 验证格式。
func (c *ValidationConverter) ruleToProtoValidation(fieldType string, rule integrationcontract.ValidationRule) string {
	// 根据字段类型选择验证器
	typePrefix := ""
	switch fieldType {
	case "string":
		typePrefix = "string"
	case "int32", "int64", "uint32", "uint64", "int":
		typePrefix = "int64"
	case "float", "double", "float32", "float64":
		typePrefix = "double"
	case "bool":
		typePrefix = "bool"
	case "bytes":
		typePrefix = "bytes"
	default:
		typePrefix = "string"
	}

	switch rule.Rule {
	case "required":
		return "required: true"
	case "email":
		return fmt.Sprintf("%s.email: true", typePrefix)
	case "uri", "url":
		return fmt.Sprintf("%s.uri: true", typePrefix)
	case "uuid":
		return fmt.Sprintf("%s.uuid: true", typePrefix)
	case "min_len":
		if num, ok := rule.Value.(int); ok {
			return fmt.Sprintf("%s.min_len: %d", typePrefix, num)
		}
	case "max_len":
		if num, ok := rule.Value.(int); ok {
			return fmt.Sprintf("%s.max_len: %d", typePrefix, num)
		}
	case "len":
		if num, ok := rule.Value.(int); ok {
			return fmt.Sprintf("%s.len: %d", typePrefix, num)
		}
	case "gte":
		if num, ok := rule.Value.(float64); ok {
			return fmt.Sprintf("%s.gte: %g", typePrefix, num)
		}
	case "lte":
		if num, ok := rule.Value.(float64); ok {
			return fmt.Sprintf("%s.lte: %g", typePrefix, num)
		}
	case "gt":
		if num, ok := rule.Value.(float64); ok {
			return fmt.Sprintf("%s.gt: %g", typePrefix, num)
		}
	case "lt":
		if num, ok := rule.Value.(float64); ok {
			return fmt.Sprintf("%s.lt: %g", typePrefix, num)
		}
	case "pattern":
		if pattern, ok := rule.Value.(string); ok {
			return fmt.Sprintf("%s.pattern: \"%s\"", typePrefix, pattern)
		}
	case "prefix":
		if prefix, ok := rule.Value.(string); ok {
			return fmt.Sprintf("%s.prefix: \"%s\"", typePrefix, prefix)
		}
	case "suffix":
		if suffix, ok := rule.Value.(string); ok {
			return fmt.Sprintf("%s.suffix: \"%s\"", typePrefix, suffix)
		}
	case "contains":
		if substr, ok := rule.Value.(string); ok {
			return fmt.Sprintf("%s.contains: \"%s\"", typePrefix, substr)
		}
	case "not_contains":
		if substr, ok := rule.Value.(string); ok {
			return fmt.Sprintf("%s.not_contains: \"%s\"", typePrefix, substr)
		}
	case "in":
		if values, ok := rule.Value.([]string); ok {
			return fmt.Sprintf("%s.in: [\"%s\"]", typePrefix, strings.Join(values, "\", \""))
		}
	case "not_in":
		if values, ok := rule.Value.([]string); ok {
			return fmt.Sprintf("%s.not_in: [\"%s\"]", typePrefix, strings.Join(values, "\", \""))
		}
	}

	return ""
}

// c 是 ValidationConverter 的别名，用于在 GenerateProtoValidation 中调用方法
var c = NewValidationConverter()
