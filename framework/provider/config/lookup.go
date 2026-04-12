package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ngq/gorp/framework/contract"
)

// GetAny 返回第一个存在的配置值。
func GetAny(cfg contract.Config, keys ...string) (any, bool) {
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

// GetStringAny 返回第一个非空字符串配置。
func GetStringAny(cfg contract.Config, keys ...string) string {
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

// GetBoolAny 返回第一个可解析布尔配置及是否存在。
func GetBoolAny(cfg contract.Config, keys ...string) (bool, bool) {
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

// GetIntAny 返回第一个可解析整数配置。
func GetIntAny(cfg contract.Config, keys ...string) int {
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

// GetFloatAny 返回第一个可解析浮点配置。
func GetFloatAny(cfg contract.Config, keys ...string) float64 {
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

// GetStringSliceAny 返回第一个字符串切片配置。
func GetStringSliceAny(cfg contract.Config, keys ...string) []string {
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
				out = append(out, fmt.Sprintf("%v", item))
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

// GetStringMapAny 返回第一个字符串 map 配置。
func GetStringMapAny(cfg contract.Config, keys ...string) map[string]string {
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
