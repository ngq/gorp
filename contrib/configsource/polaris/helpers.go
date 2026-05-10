// Package polaris provides Polaris configuration parsing helpers.
// This file contains utility functions for config decoding and normalization.
//
// 本包提供 Polaris 配置解析辅助函数。
// 本文件包含配置解码和规范化的工具函数。
package polaris

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

// decodeContent decodes config content from YAML or JSON format.
//
// decodeContent 从 YAML 或 JSON 格式解码配置内容。
func decodeContent(content string, fallbackKey string) (map[string]any, error) {
	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(content), &decoded); err == nil && len(decoded) > 0 {
		return normalizeMap(decoded), nil
	}

	var object map[string]any
	if err := json.Unmarshal([]byte(content), &object); err == nil && len(object) > 0 {
		return normalizeMap(object), nil
	}

	return map[string]any{fallbackKey: content}, nil
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

// isRetryablePolarisWatchError determines if watch error is retryable.
//
// isRetryablePolarisWatchError 判断 watch 错误是否可重试。
func isRetryablePolarisWatchError(err error) bool {
	return errors.Is(err, ErrSourceUnavailable)
}

// normalizePolarisRevision generates a revision hash from snapshot content.
//
// normalizePolarisRevision 从快照内容生成版本哈希。
func normalizePolarisRevision(snapshot polarisConfigSnapshot) string {
	revision := strings.TrimSpace(snapshot.Revision)
	if revision != "" {
		return revision
	}
	sum := sha1.Sum([]byte(snapshot.Content))
	return fmt.Sprintf("%x", sum[:])
}

// normalizePolarisAddresses normalizes server addresses format.
// Removes protocol prefixes and handles multiple addresses.
//
// normalizePolarisAddresses 规范化服务器地址格式。
// 移除协议前缀并处理多个地址。
func normalizePolarisAddresses(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, "://") {
			parsed, err := url.Parse(candidate)
			if err != nil {
				return nil, fmt.Errorf("polaris: invalid server address: %w", err)
			}
			if parsed.Host != "" {
				candidate = parsed.Host
			}
		}
		addresses = append(addresses, candidate)
	}
	if len(addresses) == 0 {
		return nil, ErrServerAddressRequired
	}
	return addresses, nil
}