package http

// Package http API.
//
// @title           Gorp Admin API
// @version         0.1.0
// @description     Gin-based framework demo.
// @BasePath        /

import (
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// swaggerBasicAuthConfig 描述 swagger basic auth 的开关与凭证。
//
// 中文说明：
// - 用于控制是否给 /swagger/* 路由增加基础认证保护。
// - 这是一个轻量防护层，主要用于避免文档页面被随手暴露。
type swaggerBasicAuthConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// swaggerConfig 是 swagger 顶层配置结构。
//
// 中文说明：
// - Enable 控制是否整体启用 swagger 路由。
// - BasicAuth 控制是否额外开启基础认证。
type swaggerConfig struct {
	Enable    bool                   `mapstructure:"enable"`
	BasicAuth swaggerBasicAuthConfig `mapstructure:"basic_auth"`
}

// RegisterSwagger 挂载 swagger-ui 路由。
//
// 中文说明：
// - 生产环境默认不建议暴露 swagger，因此这里提供 config 开关：`swagger.enable`。
// - 为避免 swagger 被公网扫到，可以启用 basic auth：`swagger.basic_auth.enable=true`。
// - 如果开关关闭，则不注册路由（对外表现为 404）。
func RegisterSwagger(c contract.Container) error {
	// read config
	cfg, err := frameworkcontainer.MakeConfig(c)
	if err != nil {
		return err
	}

	var sw swaggerConfig
	if err := cfg.Unmarshal("swagger", &sw); err != nil {
		// 如果 swagger 配置不存在，按“关闭”处理，避免误暴露
		sw.Enable = false
	}
	if !sw.Enable {
		return nil
	}

	engine, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	h := ginSwagger.WrapHandler(swaggerFiles.Handler)

	// 可选 basic auth
	if sw.BasicAuth.Enable {
		user := sw.BasicAuth.Username
		pass := sw.BasicAuth.Password
		if user == "" {
			user = "admin"
		}
		if pass == "" {
			pass = "admin"
		}
		engine.GET("/swagger/*any", gin.BasicAuth(gin.Accounts{user: pass}), h)
		return nil
	}

	engine.GET("/swagger/*any", h)
	return nil
}
