// Package config provides multi-key configuration lookup utilities.
// Supports fallback across multiple config keys, useful for backward compatibility.
//
// 配置包提供多键配置查找工具。
// 支持在多个配置键之间回退查找，适用于向后兼容场景。
package config

import (
	"fmt"
	"strconv"
	"strings"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// GetAny returns the first existing configuration value from multiple keys.
// Iterates through keys in order, returns first non-nil value.
// Core logic: Validate config, iterate keys, return first match.
//
// GetAny 从多个键中返回第一个存在的配置值。
// 按顺序遍历键，返回第一个非 nil 的值。
// 核心逻辑：验证配置对象、遍历键、返回首个匹配。
func GetAny(cfg datacontract.Config, keys ...string) (any, bool) {
	if cfg == nil {
		return nil, false
	}
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if v := cfg.Get(key); v != nil {
			return v, true
		}
	}
	return nil, false
}

// GetStringAny returns the first non-empty string configuration from multiple keys.
// Handles string conversion for various types.
// Core logic: Iterate keys, trim whitespace, return first non-empty.
//
// GetStringAny 从多个键中返回第一个非空字符串配置。
// 处理各种类型的字符串转换。
// 核心逻辑：遍历键、去除空白、返回首个非空值。
func GetStringAny(cfg datacontract.Config, keys ...string) string {
	if cfg == nil {
		return ""
	}
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if s := strings.TrimSpace(cfg.GetString(key)); s != "" {
			return s
		}
		if v, ok := GetAny(cfg, key); ok {
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}

// GetBoolAny returns the first parseable boolean configuration and existence flag.
// Supports bool, string, int, int64, float64 types.
// Core logic: Iterate keys, parse to bool, return value and existence.
//
// GetBoolAny 返回第一个可解析布尔配置及是否存在标志。
// 支持 bool、string、int、int64、float64 类型。
// 核心逻辑：遍历键、解析为布尔值、返回值和存在标志。
func GetBoolAny(cfg datacontract.Config, keys ...string) (bool, bool) {
	for _, key := range keys {
		v, ok := GetAny(cfg, key)
		if !ok {
			continue
		}
		switch b := v.(type) {
		case bool:
			return b, true
		case string:
			parsed, err := strconv.ParseBool(strings.TrimSpace(b))
			if err == nil {
				return parsed, true
			}
		case int:
			return b != 0, true
		case int64:
			return b != 0, true
		case float64:
			return b != 0, true
		}
	}
	return false, false
}

// GetIntAny returns the first parseable integer configuration from multiple keys.
// Handles int, int64, float64, string types.
// Core logic: Iterate keys, parse to int, return first valid.
//
// GetIntAny 从多个键中返回第一个可解析整数配置。
// 处理 int、int64、float64、string 类型。
// 核心逻辑：遍历键、解析为整数、返回首个有效值。
func GetIntAny(cfg datacontract.Config, keys ...string) int {
	for _, key := range keys {
		v, ok := GetAny(cfg, key)
		if !ok {
			continue
		}
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case uint:
			return int(n)
		case uint64:
			return int(n)
		case float64:
			return int(n)
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(n))
			if err == nil {
				return parsed
			}
		}
	}
	return 0
}

// GetFloatAny returns the first parseable float configuration from multiple keys.
// Handles float64, float32, int, int64, string types.
// Core logic: Iterate keys, parse to float, return first valid.
//
// GetFloatAny 从多个键中返回第一个可解析浮点配置。
// 处理 float64、float32、int、int64、string 类型。
// 核心逻辑：遍历键、解析为浮点数、返回首个有效值。
func GetFloatAny(cfg datacontract.Config, keys ...string) float64 {
	for _, key := range keys {
		v, ok := GetAny(cfg, key)
		if !ok {
			continue
		}
		switch n := v.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
			if err == nil {
				return parsed
			}
		}
	}
	return 0
}

// GetStringSliceAny returns the first string slice configuration from multiple keys.
// Handles []string, []any, string types with proper conversion.
// Core logic: Iterate keys, convert to []string, return first non-empty.
//
// GetStringSliceAny 从多个键中返回第一个字符串切片配置。
// 处理 []string、[]any、string 类型并进行适当转换。
// 核心逻辑：遍历键、转换为 []string、返回首个非空。
func GetStringSliceAny(cfg datacontract.Config, keys ...string) []string {
	for _, key := range keys {
		v, ok := GetAny(cfg, key)
		if !ok {
			continue
		}
		switch arr := v.(type) {
		case []string:
			if len(arr) > 0 {
				return arr
			}
		case []any:
			out := make([]string, 0, len(arr))
			for _, item := range arr {
				switch v := item.(type) {
				case string:
					if v != "" {
						out = append(out, v)
					}
				case fmt.Stringer:
					s := v.String()
					if s != "" {
						out = append(out, s)
					}
				default:
					s := fmt.Sprintf("%v", item)
					if s != "" && s != "<nil>" && s != "map[]" && s != "[]" {
						out = append(out, s)
					}
				}
			}
			if len(out) > 0 {
				return out
			}
		case string:
			s := strings.TrimSpace(arr)
			if s != "" {
				return []string{s}
			}
		}
	}
	return nil
}

// GetStringMapAny returns the first string map configuration from multiple keys.
// Handles map[string]string, map[string]any with proper conversion.
// Core logic: Iterate keys, convert to map[string]string, return first found.
//
// GetStringMapAny 从多个键中返回第一个字符串 map 配置。
// 处理 map[string]string、map[string]any 类型并进行适当转换。
// 核心逻辑：遍历键、转换为 map[string]string、返回首个找到的。
func GetStringMapAny(cfg datacontract.Config, keys ...string) map[string]string {
	for _, key := range keys {
		v, ok := GetAny(cfg, key)
		if !ok {
			continue
		}
		switch m := v.(type) {
		case map[string]string:
			return m
		case map[string]any:
			out := make(map[string]string, len(m))
			for mk, mv := range m {
				out[mk] = fmt.Sprintf("%v", mv)
			}
			return out
		}
	}
	return nil
}
