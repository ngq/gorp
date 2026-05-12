// Package bootstrap_test provides unit tests for config validation and schema checking.
//
// 适用场景：
// - 验证 ValidateCriticalConfig 在各种配置场景下的行为。
// - 有效配置通过校验；缺失必填字段报错；条件校验正确跳过或触发；错误消息格式清晰可读。
package bootstrap

import (
	"context"
	"reflect"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
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

func (s *mapConfigStub) Env() string    { return "test" }
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

// ---------------------------------------------------------------------------
// 有效配置测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigPassesWithValidConfig 验证完整有效的配置能通过校验。
//
// TestValidateCriticalConfigPassesWithValidConfig verifies that a complete valid config passes validation.
func TestValidateCriticalConfigPassesWithValidConfig(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigPassesWithAllSections 验证所有节都存在且有效时通过校验。
//
// TestValidateCriticalConfigPassesWithAllSections verifies validation passes when all sections are present and valid.
func TestValidateCriticalConfigPassesWithAllSections(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("server", map[string]any{}) // server.http 通过嵌套需要特殊处理
	cfg.setSection("log", map[string]any{"level": "debug", "format": "json"})
	cfg.setSection("database", map[string]any{"driver": "mysql", "dsn": "user:pass@tcp(localhost:3306)/db"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigPassesWithJsonFormat 验证 json 日志格式通过校验。
//
// TestValidateCriticalConfigPassesWithJsonFormat verifies json log format passes validation.
func TestValidateCriticalConfigPassesWithJsonFormat(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":9090"})
	cfg.setSection("log", map[string]any{"level": "warn", "format": "json"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// 缺失必填字段测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigFailsWithoutAppAddress 验证缺失 app.address 时报错。
//
// TestValidateCriticalConfigFailsWithoutAppAddress verifies error when app.address is missing.
func TestValidateCriticalConfigFailsWithoutAppAddress(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{}) // address 为空
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: app.address is required")
}

// TestValidateCriticalConfigFailsWithoutLogLevel 验证缺失 log.level 时报错。
//
// TestValidateCriticalConfigFailsWithoutLogLevel verifies error when log.level is missing.
func TestValidateCriticalConfigFailsWithoutLogLevel(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"format": "console"}) // level 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.level is required")
}

// TestValidateCriticalConfigFailsWithoutLogFormat 验证缺失 log.format 时报错。
//
// TestValidateCriticalConfigFailsWithoutLogFormat verifies error when log.format is missing.
func TestValidateCriticalConfigFailsWithoutLogFormat(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info"}) // format 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.format is required")
}

// TestValidateCriticalConfigFailsWithInvalidLogLevel 验证无效的 log.level 值报错。
//
// TestValidateCriticalConfigFailsWithInvalidLogLevel verifies error with invalid log.level value.
func TestValidateCriticalConfigFailsWithInvalidLogLevel(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "verbose", "format": "console"}) // "verbose" 不在允许列表

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.level must be one of [debug info warn error]")
}

// TestValidateCriticalConfigFailsWithInvalidLogFormat 验证无效的 log.format 值报错。
//
// TestValidateCriticalConfigFailsWithInvalidLogFormat verifies error with invalid log.format value.
func TestValidateCriticalConfigFailsWithInvalidLogFormat(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "yaml"}) // "yaml" 不在允许列表

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.format must be one of [console json]")
}

// ---------------------------------------------------------------------------
// 条件校验测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigSkipsDatabaseWhenAbsent 验证 database 节缺失时跳过校验。
//
// TestValidateCriticalConfigSkipsDatabaseWhenAbsent verifies database validation is skipped when section is absent.
func TestValidateCriticalConfigSkipsDatabaseWhenAbsent(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	// 不设置 database 节

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigSkipsDatabaseWhenEmpty 验证 database 节存在但 driver/dsn 都为空时跳过校验。
//
// TestValidateCriticalConfigSkipsDatabaseWhenEmpty verifies database validation is skipped
// when section exists but both driver and dsn are empty.
func TestValidateCriticalConfigSkipsDatabaseWhenEmpty(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{}) // driver 和 dsn 都为空

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigFailsWithDatabaseDriverButNoDSN 验证 database.driver 存在但 dsn 缺失时报错。
//
// TestValidateCriticalConfigFailsWithDatabaseDriverButNoDSN verifies error when database.driver
// is set but database.dsn is missing.
func TestValidateCriticalConfigFailsWithDatabaseDriverButNoDSN(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"driver": "mysql"}) // dsn 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: database.dsn is required")
}

// TestValidateCriticalConfigFailsWithDatabaseDSNButNoDriver 验证 database.dsn 存在但 driver 缺失时报错。
//
// TestValidateCriticalConfigFailsWithDatabaseDSNButNoDriver verifies error when database.dsn
// is set but database.driver is missing.
func TestValidateCriticalConfigFailsWithDatabaseDSNButNoDriver(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"dsn": "user:pass@tcp(localhost:3306)/db"}) // driver 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: database.driver is required")
}

// TestValidateCriticalConfigFailsWithInvalidDatabaseDriver 验证无效的 database.driver 值报错。
//
// TestValidateCriticalConfigFailsWithInvalidDatabaseDriver verifies error with invalid database.driver value.
func TestValidateCriticalConfigFailsWithInvalidDatabaseDriver(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"driver": "oracle", "dsn": "user:pass@localhost/db"}) // "oracle" 不在允许列表

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: database.driver must be one of [sqlite mysql postgres pgx]")
}

// TestValidateCriticalConfigPassesWithValidDatabase 验证有效的 database 配置通过校验。
//
// TestValidateCriticalConfigPassesWithValidDatabase verifies valid database config passes validation.
func TestValidateCriticalConfigPassesWithValidDatabase(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"driver": "sqlite", "dsn": "file:test.db"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// server.http 条件校验测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigSkipsServerHTTPWhenAbsent 验证 server.http 节缺失时跳过校验。
//
// TestValidateCriticalConfigSkipsServerHTTPWhenAbsent verifies server.http validation is skipped when absent.
func TestValidateCriticalConfigSkipsServerHTTPWhenAbsent(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	// 不设置 server.http

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// 多重错误聚合测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigAggregatesMultipleErrors 验证多个校验错误被聚合到一条错误消息中。
//
// TestValidateCriticalConfigAggregatesMultipleErrors verifies multiple validation errors
// are aggregated into a single error message.
func TestValidateCriticalConfigAggregatesMultipleErrors(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{})                                    // address 缺失
	cfg.setSection("log", map[string]any{"level": "verbose", "format": "yaml"}) // level 和 format 都无效

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: app.address is required")
	require.Contains(t, err.Error(), "config: log.level must be one of [debug info warn error]")
	require.Contains(t, err.Error(), "config: log.format must be one of [console json]")
}

// ---------------------------------------------------------------------------
// 各种日志级别和格式组合测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigAcceptsAllLogLevels 验证所有合法日志级别均通过校验。
//
// TestValidateCriticalConfigAcceptsAllLogLevels verifies all valid log levels pass validation.
func TestValidateCriticalConfigAcceptsAllLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg := newMapConfigStub()
		cfg.setSection("app", map[string]any{"address": ":8080"})
		cfg.setSection("log", map[string]any{"level": level, "format": "console"})

		err := ValidateCriticalConfig(cfg)
		require.NoError(t, err, "log.level=%s should be valid", level)
	}
}

// TestValidateCriticalConfigAcceptsAllDatabaseDrivers 验证所有合法数据库驱动均通过校验。
//
// TestValidateCriticalConfigAcceptsAllDatabaseDrivers verifies all valid database drivers pass validation.
func TestValidateCriticalConfigAcceptsAllDatabaseDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "mysql", "postgres", "pgx"} {
		cfg := newMapConfigStub()
		cfg.setSection("app", map[string]any{"address": ":8080"})
		cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
		cfg.setSection("database", map[string]any{"driver": driver, "dsn": "test-dsn"})

		err := ValidateCriticalConfig(cfg)
		require.NoError(t, err, "database.driver=%s should be valid", driver)
	}
}

// ---------------------------------------------------------------------------
// 完全空配置测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigFailsWithEmptyConfig 验证完全空的配置会报出必填字段缺失错误。
//
// TestValidateCriticalConfigFailsWithEmptyConfig verifies that an entirely empty config
// produces errors about missing required fields.
func TestValidateCriticalConfigFailsWithEmptyConfig(t *testing.T) {
	cfg := newMapConfigStub()

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: app.address is required")
	require.Contains(t, err.Error(), "config: log.level is required")
	require.Contains(t, err.Error(), "config: log.format is required")
}
