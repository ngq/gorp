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
