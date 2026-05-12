// Package config_test provides unit tests for the config service.
//
// 适用场景：
// - 验证 Config Service 的 Watch、Load 和 key 变更通知行为。
package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestEnvKeyReplacer(t *testing.T) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = os.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	require.Equal(t, "127.0.0.1:6379", v.GetString("redis.addr"))
}
