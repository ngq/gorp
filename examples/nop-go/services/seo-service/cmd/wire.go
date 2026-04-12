//go:build wireinject

package main

import (
	"nop-go/services/seo-service/internal/biz"
	"nop-go/services/seo-service/internal/data"
	"nop-go/services/seo-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

// wireSEOService 使用 Wire 生成 seo-service 的装配代码。
func wireSEOService(db *gorm.DB, jwtSvc contract.JWTService, enabled, sitemapEnabled, canonicalUrlsEnabled, customMetaEnabled bool) (*service.SEOService, error) {
	panic(wire.Build(
		data.NewUrlRecordRepository,
		data.NewUrlRedirectRepository,
		data.NewMetaInfoRepository,
		data.NewSitemapNodeRepository,
		provideSEOConfig,
		biz.NewSEOUseCase,
		service.NewSEOService,
	))
}

func provideSEOConfig(enabled, sitemapEnabled, canonicalUrlsEnabled, customMetaEnabled bool) biz.SEOConfig {
	return biz.SEOConfig{
		Enabled:              enabled,
		SitemapEnabled:       sitemapEnabled,
		CanonicalUrlsEnabled: canonicalUrlsEnabled,
		CustomMetaEnabled:    customMetaEnabled,
	}
}
