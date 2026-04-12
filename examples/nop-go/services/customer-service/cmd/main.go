// Package main 客户服务入口
//
// 客户服务 API 文档
//
// @title           客户服务 API
// @version         1.0
// @description     客户服务 - 用户注册、登录、地址管理、GDPR合规
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@nop-go.local

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8002
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT认证令牌，格式: Bearer {token}
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/customer-service/internal/models"

	_ "nop-go/services/customer-service/docs" // swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("customer-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	customerService, err := wireCustomerService(rt.DB, bootstrap.MustMakeJWTService(rt.Container))
	if err != nil {
		return err
	}

	customerService.RegisterRoutes(rt.Engine)
	registerSwagger(rt.Engine)
	return nil
}

// registerSwagger 注册 Swagger UI 路由
// 中文说明：
// - 提供 /swagger/* 端点用于查看 API 文档
// - 可通过配置 swagger.enable 控制是否启用
func registerSwagger(engine *gin.Engine) {
	// 检查是否启用 swagger（生产环境应关闭）
	enableSwagger := os.Getenv("ENABLE_SWAGGER") != "false"
	if !enableSwagger {
		return
	}

	// 注册 swagger 路由
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// autoMigrate 执行数据库表结构迁移。
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Customer{},
		&models.CustomerRole{},
		&models.Address{},
		&models.CustomerPassword{},
		&models.ExternalAuthenticationRecord{},
		&models.RewardPointsHistory{},
		// GDPR 模型
		&models.GdprConsent{},
		&models.GdprLog{},
		&models.GdprRequest{},
		&models.CustomerConsent{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}

	// 初始化预定义角色
	initRoles(db)

	return nil
}

// initRoles 初始化预定义角色。
func initRoles(db *gorm.DB) {
	roles := []models.CustomerRole{
		{Name: "游客", SystemName: models.RoleGuests, IsSystem: true, IsActive: true},
		{Name: "注册用户", SystemName: models.RoleRegistered, IsSystem: true, IsActive: true},
		{Name: "供应商", SystemName: models.RoleVendors, IsSystem: true, IsActive: true},
		{Name: "管理员", SystemName: models.RoleAdmins, IsSystem: true, IsActive: true},
		{Name: "论坛管理员", SystemName: models.RoleForumMods, IsSystem: true, IsActive: true},
	}

	for _, role := range roles {
		var existing models.CustomerRole
		if err := db.Where("system_name = ?", role.SystemName).First(&existing).Error; err == gorm.ErrRecordNotFound {
			db.Create(&role)
		}
	}
}