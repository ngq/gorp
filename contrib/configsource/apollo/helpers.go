// Package apollo provides Apollo configuration parsing helpers.
// This file contains utility functions for config decoding and normalization.
//
// 本包提供 Apollo 配置解析辅助函数。
// 本文件包含配置解码和规范化的工具函数。
package apollo

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// decodeContent decodes config content from YAML or JSON format.
//
// decodeContent 从 YAML 或 JSON 格式解码配置内容。
func decodeContent(content string, namespace string) (map[string]any, error) {
	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(content), &decoded); err == nil && len(decoded) > 0 {
		return normalizeMap(decoded), nil
	}

	var object map[string]any
	if err := json.Unmarshal([]byte(content), &object); err == nil && len(object) > 0 {
		return normalizeMap(object), nil
	}

	return map[string]any{
		namespace: content,
	}, nil
}

// cloneMap creates a deep copy of a map.
//
// cloneMap 创建 map 的深拷贝。
func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for k, v := range input {
		if nested, ok := v.(map[string]any); ok {
			cloned[k] = cloneMap(nested)
			continue
		}
		cloned[k] = v
	}
	return cloned
}

// lookupNestedValue looks up a nested value by dot-separated path.
//
// lookupNestedValue 通过点分隔路径查找嵌套值。
func lookupNestedValue(data map[string]any, path string) (any, bool) {
	if path == "" {
		return data, true
	}

	current := any(data)
	for _, part := range strings.Split(path, ".") {
		nested, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = nested[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// normalizeMap normalizes map keys and nested structures.
//
// normalizeMap 规范化 map key 和嵌套结构。
func normalizeMap(input map[string]any) map[string]any {
	normalized := make(map[string]any, len(input))
	for k, v := range input {
		switch typed := v.(type) {
		case map[string]any:
			normalized[k] = normalizeMap(typed)
		case map[any]any:
			converted := make(map[string]any, len(typed))
			for nestedKey, nestedValue := range typed {
				keyString, ok := nestedKey.(string)
				if !ok {
					continue
				}
				converted[keyString] = nestedValue
			}
			normalized[k] = normalizeMap(converted)
		default:
			normalized[k] = v
		}
	}
	return normalized
}

// assignNestedValue assigns a value to a nested path in the map.
//
// assignNestedValue 将值分配给 map 中的嵌套路径。
func assignNestedValue(target map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := target
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}
}

// isRetryableApolloWatchError determines if watch error is retryable.
//
// isRetryableApolloWatchError 判断 watch 错误是否可重试。
func isRetryableApolloWatchError(err error) bool {
	return errors.Is(err, ErrSourceUnavailable)
}

// normalizeApolloRevision generates a revision hash from snapshot content.
//
// normalizeApolloRevision 从快照内容生成版本哈希。
func normalizeApolloRevision(snapshot apolloConfigSnapshot) string {
	revision := strings.TrimSpace(snapshot.Revision)
	if revision != "" {
		return revision
	}
	sum := sha1.Sum([]byte(snapshot.Content))
	return fmt.Sprintf("%x", sum[:])
}