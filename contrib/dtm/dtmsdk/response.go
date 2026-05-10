// Package dtmsdk provides response parsing utilities for DTM API.
// This file handles transaction info parsing from DTM server responses.
//
// 本包提供 DTM API 响应解析工具。
// 本文件处理 DTM 服务器响应的事务信息解析。
package dtmsdk

import (
	"encoding/json"
	"fmt"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// parseTransactionInfo extracts TransactionInfo from raw API response.
//
// parseTransactionInfo 从原始 API 响应提取 TransactionInfo。
func parseTransactionInfo(raw map[string]any, fallbackGID string) *integrationcontract.TransactionInfo {
	payload := unwrapTransactionPayload(raw)
	info := &integrationcontract.TransactionInfo{
		GID:             fallbackGID,
		Status:          stringFromMap(payload, "status"),
		TransactionType: stringFromMap(payload, "trans_type", "transaction_type"),
		CreateTime:      int64FromMap(payload, "create_time", "createTime"),
		UpdateTime:      int64FromMap(payload, "update_time", "updateTime"),
		Steps:           parseTransactionSteps(payload),
	}
	if gid := stringFromMap(payload, "gid"); gid != "" {
		info.GID = gid
	}
	if info.Status == "" {
		info.Status = "unknown"
	}
	if info.TransactionType == "" {
		info.TransactionType = "unknown"
	}
	return info
}

// unwrapTransactionPayload unwraps nested payload from raw response.
//
// unwrapTransactionPayload 从原始响应解开嵌套 payload。
func unwrapTransactionPayload(raw map[string]any) map[string]any {
	for _, key := range []string{"transaction", "data", "result"} {
		if nested, ok := raw[key].(map[string]any); ok && len(nested) > 0 {
			return nested
		}
	}
	return raw
}

// parseTransactionSteps extracts TransactionStep slice from payload.
//
// parseTransactionSteps 从 payload 提取 TransactionStep 列表。
func parseTransactionSteps(raw map[string]any) []integrationcontract.TransactionStep {
	values, ok := raw["steps"].([]any)
	if !ok {
		return nil
	}
	steps := make([]integrationcontract.TransactionStep, 0, len(values))
	for _, value := range values {
		stepMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		steps = append(steps, integrationcontract.TransactionStep{
			BranchID: stringFromMap(stepMap, "branch_id", "branchID"),
			Status:   stringFromMap(stepMap, "status"),
			Op:       stringFromMap(stepMap, "op"),
			URL:      stringFromMap(stepMap, "url", "action", "try", "confirm", "cancel"),
		})
	}
	return steps
}

// stringFromMap extracts string value from map with multiple key fallbacks.
//
// stringFromMap 从 map 提取字符串值，支持多个 key 回退。
func stringFromMap(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if s := stringFromAny(value); s != "" {
				return s
			}
		}
	}
	return ""
}

// int64FromMap extracts int64 value from map with multiple key fallbacks.
//
// int64FromMap 从 map 提取 int64 值，支持多个 key 回退。
func int64FromMap(values map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if n, ok := int64FromAny(value); ok {
				return n
			}
		}
	}
	return 0
}

// stringFromAny converts any value to string.
//
// stringFromAny 将任意值转换为字符串。
func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}

// int64FromAny converts any value to int64.
//
// int64FromAny 将任意值转换为 int64。
func int64FromAny(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	case string:
		if v == "" {
			return 0, false
		}
		parsed, err := json.Number(v).Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}