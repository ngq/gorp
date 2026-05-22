package main

import (
	"fmt"
	"os"

	contentdata "nop-go/services/content/internal/data"
	contenthttp "nop-go/services/content/internal/server/http"
	"github.com/ngq/gorp"
	_ "nop-go/shared" // 微服务治理组件统一导入
)

func main() {
	if err := gorp.Run(
		gorp.GRPC(),
		gorp.WithMicroGovernance(),
		gorp.WithMigrate(migrate),
		gorp.WithSetup(setup),
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	// 自动迁移所有内容持久化对象
	return rt.DB.AutoMigrate(
		&contentdata.BlogPO{},
		&contentdata.NewsPO{},
		&contentdata.TopicPO{},
		&contentdata.PollPO{},
		&contentdata.PollAnswerPO{},
	)
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("content-service requires gorm database")
	}
	services, err := wireContentServices(rt.DB)
	if err != nil {
		return err
	}
	contenthttp.RegisterRoutes(rt.Router, services)
	return nil
}
