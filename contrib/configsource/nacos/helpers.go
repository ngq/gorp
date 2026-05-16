// Package nacos provides Nacos configuration parsing helpers.
// This file contains utility functions for config decoding and normalization.
//
// 本包提供 Nacos 配置解析辅助函数。
// 本文件包含配置解码和规范化的工具函数。
package nacos

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// decodeContent decodes config content from YAML or plain text format.
//
// decodeContent 从 YAML 或纯文本格式解码配置内容。
// 该函数尝试将配置内容解析为结构化数据：
//   - 如果内容是 YAML 格式，解析为 map[string]any
//   - 如果内容是纯文本，将文本作为 dataID key 的值存储
//
// 参数：
//   - content: 配置内容字符串
//   - dataID: Nacos 数据 ID，用于纯文本场景的 key
//
// 返回：
//   - 解析后的 map[string]any
//   - 解析错误（如果 YAML 解析失败）
func decodeContent(content, dataID string) (map[string]any, error) {
	// 尝试解析为结构化内容（YAML）
	value, err := decodeStructuredContent(content)
	if err != nil {
		return nil, err
	}
	// 如果解析结果是 map，直接返回
	if asMap, ok := value.(map[string]any); ok {
		return asMap, nil
	}
	// 如果解析结果是其他类型（如纯文本），将其作为 dataID key 的值
	return map[string]any{dataID: value}, nil
}

// encodeContent encodes config value to content string.
//
// encodeContent 将配置值编码为内容字符串。
// 该函数支持以下类型的编码：
//   - string: 直接返回
//   - []byte: 转换为字符串
//   - 其他类型: YAML 编码
func encodeContent(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		// 使用 YAML 编码
		data, err := yaml.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("configsource.nacos: encode config content failed: %w", err)
		}
		return string(data), nil
	}
}

// normalizeYAMLValue normalizes YAML decoded value types.
//
// normalizeYAMLValue 规范化 YAML 解码后的值类型。
// yaml.v3 解码时会将某些类型转换为非标准类型：
//   - map[any]any -> map[string]any（key 转为字符串）
//   - 保持 []any 和基本类型不变
//
// 该函数递归处理嵌套结构，确保所有 map 的 key 都是 string 类型。
func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		// 递归规范化 map[string]any 的值
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[key] = normalizeYAMLValue(nested)
		}
		return result
	case map[any]any:
		// 将 map[any]any 转换为 map[string]any
		// yaml.v3 解码整数 key 时会产生 map[any]any
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[fmt.Sprint(key)] = normalizeYAMLValue(nested)
		}
		return result
	case []any:
		// 递归规范化数组元素
		result := make([]any, 0, len(typed))
		for _, nested := range typed {
			result = append(result, normalizeYAMLValue(nested))
		}
		return result
	default:
		// 基本类型直接返回
		return typed
	}
}

// decodeStructuredContent attempts to decode content as YAML.
//
// decodeStructuredContent 尝试将内容解码为 YAML。
// 如果 YAML 解析失败，将内容作为纯文本返回。
func decodeStructuredContent(content string) (any, error) {
	var value any
	if err := yaml.Unmarshal([]byte(content), &value); err == nil {
		return normalizeYAMLValue(value), nil
	}
	// YAML 解析失败，返回纯文本
	return strings.TrimSpace(content), nil
}

// lookupNestedValue looks up a nested value by dot-separated path.
//
// lookupNestedValue 通过点分隔路径查找嵌套值。
// 该函数支持按路径查找配置值，如 "app.name" 查找 cache["app"]["name"]。
// 参数：
//   - data: 配置数据 map
//   - path: 点分隔的路径字符串，空路径返回整个 map
//
// 返回：
//   - 找到的值
//   - 是否找到（bool）
func lookupNestedValue(data map[string]any, path string) (any, bool) {
	// 空路径返回整个 map 的拷贝
	if path == "" {
		return cloneMap(data), true
	}

	// 按点分隔路径逐层查找
	current := any(data)
	for _, segment := range strings.Split(path, ".") {
		// 当前层级必须是 map
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		// 查找下一层级
		next, exists := asMap[segment]
		if !exists {
			return nil, false
		}
		current = next
	}
	return current, true
}

// cloneMap creates a deep copy of a map.
// Nested maps are recursively copied to prevent shared-reference mutations.
//
// cloneMap 创建 map 的深拷贝。
// 嵌套 map 递归拷贝，防止共享引用导致的并发数据竞争。
func cloneMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		switch v := value.(type) {
		case map[string]any:
			result[key] = cloneMap(v)
		default:
			result[key] = value
		}
	}
	return result
}
