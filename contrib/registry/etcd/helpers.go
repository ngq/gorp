// Package etcd provides etcd registry helper functions.
// This file contains utility functions for address parsing, ID generation, and data extraction.
//
// 本包提供 etcd 注册中心辅助函数。
// 本文件包含地址解析、ID 生成和数据提取的工具函数。
package etcd

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// parseAddr parses address into host and port.
//
// parseAddr 将地址解析为 host 和 port。
func parseAddr(addr string) (host string, port int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		host = parts[0]
		port, _ = strconv.Atoi(parts[1])
	} else {
		host = addr
	}
	return host, port
}

// generateServiceID generates a unique service instance ID.
//
// generateServiceID 生成唯一的服务实例 ID。
func generateServiceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}

// getString extracts a string value from a map.
//
// getString 从 map 中提取字符串值。
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// getInt extracts an integer value from a map.
//
// getInt 从 map 中提取整数值。
func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

// getBool extracts a boolean value from a map.
//
// getBool 从 map 中提取布尔值。
func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return true
}

// getMap extracts a string map from a map.
//
// getMap 从 map 中提取字符串 map。
func getMap(m map[string]any, key string) map[string]string {
	result := make(map[string]string)
	if v, ok := m[key]; ok {
		if meta, ok := v.(map[string]any); ok {
			for k, val := range meta {
				result[k] = fmt.Sprintf("%v", val)
			}
		}
	}
	return result
}

// cloneStringMap creates a copy of a string map.
//
// cloneStringMap 创建字符串 map 的副本。
func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

// randShuffle shuffles a slice using math/rand.
//
// randShuffle 使用 math/rand 随机打乱切片。
func randShuffle(n int, swap func(i, j int)) {
	rand.Shuffle(n, swap)
}

// timeNow returns current time.
//
// timeNow 返回当前时间。
func timeNow() time.Time {
	return time.Now()
}

// timeSecond returns 1 second duration.
//
// timeSecond 返回 1 秒时长。
func timeSecond() time.Duration {
	return time.Second
}