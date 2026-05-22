package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// LoadYAML 把指定 yaml 文件加载到目标配置对象。
func LoadYAML(path string, target any) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}
