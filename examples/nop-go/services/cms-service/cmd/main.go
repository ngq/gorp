// Package main 内容管理服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/cms-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("cms-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	cmsService, err := wireCMSService(rt.DB)
	if err != nil {
		return err
	}

	cmsService.RegisterRoutes(rt.Engine)
	return nil
}

func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.BlogPost{},
		&models.BlogCategory{},
		&models.News{},
		&models.Topic{},
		&models.Forum{},
		&models.ForumTopic{},
		&models.ForumPost{},
		// 新增模型
		&models.Menu{},
		&models.MenuItem{},
		&models.Poll{},
		&models.PollAnswer{},
		&models.PollVotingRecord{},
		&models.HtmlBody{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}
