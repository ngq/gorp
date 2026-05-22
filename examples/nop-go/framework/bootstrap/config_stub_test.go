// Package bootstrap_test provides unit tests for config validation and schema checking.
//
// 适用场景：
// - 验证 ValidateCriticalConfig 在各种配置场景下的行为。
// - 有效配置通过校验；缺失必填字段报错；条件校验正确跳过或触发；错误消息格式清晰可读。
package bootstrap

import (
	"context"
	"reflect"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// ---------------------------------------------------------------------------
// 测试用 Config Stub —— 支持 Unmarshal 的 map-backed 实现
// ---------------------------------------------------------------------------

// mapConfigStub 是基于 map 的 Config 实现，用于测试配置校验。
// 支持 Get / Unmarshal 等方法，Unmarshal 使用 reflect 将 map 映射到结构体。
//
// mapConfigStub is a map-backed Config implementation for testing config validation.
// Supports Get / Unmarshal; Unmarshal uses reflect to map values into structs.
type mapConfigStub struct {
	// sections 存储顶层节，key 为节名（如 "app"），value 为字段 map
	// sections stores top-level sections, key is section name (e.g. "app"), value is field map
	sections map[string]map[string]any
	// values 是扁平化的 key-value 映射，用于 Get 方法
	// values is a flattened key-value mapping, used by Get method
	values map[string]any
}

// newMapConfigStub 创建空的 mapConfigStub。
//
// newMapConfigStub creates an empty mapConfigStub.
func newMapConfigStub() *mapConfigStub {
	return &mapConfigStub{
		sections: make(map[string]map[string]any),
		values:   make(map[string]any),
	}
}

// setSection 设置一个配置节，同时更新 values 的扁平化映射。
// 例如 setSection("app", map[string]any{"address": ":8080"}) 会同时
// 设置 values["app"] 和 values["app.address"]。
//
// setSection sets a config section and updates the flattened values map.
// E.g. setSection("app", map[string]any{"address": ":8080"}) sets both
// values["app"] and values["app.address"].
func (s *mapConfigStub) setSection(section string, fields map[string]any) {
	s.sections[section] = fields
	s.values[section] = fields
	for k, v := range fields {
		s.values[section+"."+k] = v
	}
}

func (s *mapConfigStub) Env() string        { return "test" }
func (s *mapConfigStub) Get(key string) any { return s.values[key] }
func (s *mapConfigStub) GetString(key string) string {
	v, _ := s.values[key].(string)
	return v
}
func (s *mapConfigStub) GetInt(key string) int {
	v, _ := s.values[key].(int)
	return v
}
func (s *mapConfigStub) GetBool(key string) bool {
	v, _ := s.values[key].(bool)
	return v
}
func (s *mapConfigStub) GetFloat(key string) float64 {
	v, _ := s.values[key].(float64)
	return v
}
func (s *mapConfigStub) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (s *mapConfigStub) Reload(ctx context.Context) error { return nil }

// Unmarshal 将配置节解码到目标结构体。
// 使用 reflect 基于 mapstructure tag 将 map 字段映射到结构体字段。
// 仅支持顶层节（如 "app", "log", "database"），不支持嵌套子节。
//
// Unmarshal decodes a config section into the target struct.
// Uses reflect with mapstructure tags to map values into struct fields.
// Only supports top-level sections (e.g. "app", "log", "database"), not nested sub-sections.
func (s *mapConfigStub) Unmarshal(key string, out any) error {
	section, ok := s.sections[key]
	if !ok {
		// 节不存在，out 保持零值
		// Section not found, out remains zero value
		return nil
	}
	return mapToStruct(section, out)
}

// mapToStruct 使用 reflect 将 map 值赋给结构体字段（基于 mapstructure tag）。
// 仅用于测试，不处理嵌套结构体或复杂类型。
//
// mapToStruct uses reflect to assign map values to struct fields (based on mapstructure tags).
// Only for testing; does not handle nested structs or complex types.
func mapToStruct(m map[string]any, out any) error {
	v := reflect.ValueOf(out).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" {
			continue
		}
		// 取 mapstructure tag 中逗号前的部分作为 key
		// Use the part before comma in mapstructure tag as the key
		tagKey := tag
		if idx := indexByte(tag, ','); idx >= 0 {
			tagKey = tag[:idx]
		}
		if val, ok := m[tagKey]; ok {
			f := v.Field(i)
			if f.CanSet() {
				switch sv := val.(type) {
				case string:
					f.SetString(sv)
				case int:
					f.SetInt(int64(sv))
				case bool:
					f.SetBool(sv)
				}
			}
		}
	}
	return nil
}

// indexByte 返回字符串中字节 c 的首次出现位置，不存在则返回 -1。
//
// indexByte returns the index of the first occurrence of byte c in s, or -1 if not found.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
