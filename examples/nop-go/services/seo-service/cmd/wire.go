//go:build wireinject

package main

import (
	"nop-go/services/seo-service/internal/biz"
	"nop-go/services/seo-service/internal/data"
	"nop-go/services/seo-service/internal/service"

	"github.com/google/wire"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"gorm.io/gorm"
)

// wireSEOService 浣跨敤 Wire 鐢熸垚 seo-service 鐨勮閰嶄唬鐮併€?
func wireSEOService(db *gorm.DB, jwtSvc securitycontract.JWTService, enabled, sitemapEnabled, canonicalUrlsEnabled, customMetaEnabled bool) (*service.SEOService, error) {
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
