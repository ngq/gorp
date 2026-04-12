// Package main SEO服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/seo-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("seo-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	seoService, err := wireSEOService(
		rt.DB,
		bootstrap.MustMakeJWTService(rt.Container),
		rt.Config.GetBool("seo.enabled"),
		rt.Config.GetBool("seo.sitemap_enabled"),
		rt.Config.GetBool("seo.canonical_urls_enabled"),
		rt.Config.GetBool("seo.custom_meta_enabled"),
	)
	if err != nil {
		return err
	}

	seoService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.UrlRecord{},
		&models.UrlRedirect{},
		&models.MetaInfo{},
		&models.SitemapNode{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}