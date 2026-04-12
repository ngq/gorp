// Package biz SEO服务业务逻辑层
package biz

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nop-go/services/seo-service/internal/data"
	"nop-go/services/seo-service/internal/models"
)

// SEOConfig SEO配置
type SEOConfig struct {
	Enabled             bool
	SitemapEnabled      bool
	CanonicalUrlsEnabled bool
	CustomMetaEnabled   bool
}

// SEOUseCase SEO用例
type SEOUseCase struct {
	urlRecordRepo   data.UrlRecordRepository
	urlRedirectRepo data.UrlRedirectRepository
	metaInfoRepo    data.MetaInfoRepository
	sitemapNodeRepo data.SitemapNodeRepository
	config          SEOConfig
}

// NewSEOUseCase 创建SEO用例
func NewSEOUseCase(
	urlRecordRepo data.UrlRecordRepository,
	urlRedirectRepo data.UrlRedirectRepository,
	metaInfoRepo data.MetaInfoRepository,
	sitemapNodeRepo data.SitemapNodeRepository,
	config SEOConfig,
) *SEOUseCase {
	return &SEOUseCase{
		urlRecordRepo:   urlRecordRepo,
		urlRedirectRepo: urlRedirectRepo,
		metaInfoRepo:    metaInfoRepo,
		sitemapNodeRepo: sitemapNodeRepo,
		config:          config,
	}
}

// ========== URL记录管理 ==========

// CreateUrlRecord 创建URL记录
func (uc *SEOUseCase) CreateUrlRecord(ctx context.Context, req *models.UrlRecordCreateRequest) (*models.UrlRecord, error) {
	// 检查slug是否已存在
	existing, err := uc.urlRecordRepo.GetBySlug(ctx, req.Slug)
	if err == nil && existing != nil {
		return nil, data.ErrSlugExists
	}

	now := time.Now()
	record := &models.UrlRecord{
		Slug:       req.Slug,
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		LanguageID: req.LanguageID,
		IsActive:   req.IsActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.urlRecordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// GetUrlRecord 获取URL记录
func (uc *SEOUseCase) GetUrlRecord(ctx context.Context, id uint) (*models.UrlRecord, error) {
	return uc.urlRecordRepo.GetByID(ctx, id)
}

// GetUrlBySlug 通过slug获取URL记录
func (uc *SEOUseCase) GetUrlBySlug(ctx context.Context, slug string) (*models.UrlRecord, error) {
	return uc.urlRecordRepo.GetBySlug(ctx, slug)
}

// GetUrlByEntity 通过实体获取URL记录
func (uc *SEOUseCase) GetUrlByEntity(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.UrlRecord, error) {
	return uc.urlRecordRepo.GetByEntity(ctx, entityID, entityType, languageID)
}

// ListUrlRecords URL记录列表
func (uc *SEOUseCase) ListUrlRecords(ctx context.Context, page, pageSize int) ([]*models.UrlRecord, int64, error) {
	return uc.urlRecordRepo.List(ctx, page, pageSize)
}

// SearchUrlRecords 搜索URL记录
func (uc *SEOUseCase) SearchUrlRecords(ctx context.Context, keyword string, page, pageSize int) ([]*models.UrlRecord, int64, error) {
	return uc.urlRecordRepo.Search(ctx, keyword, page, pageSize)
}

// UpdateUrlRecord 更新URL记录
func (uc *SEOUseCase) UpdateUrlRecord(ctx context.Context, id uint, req *models.UrlRecordUpdateRequest) (*models.UrlRecord, error) {
	record, err := uc.urlRecordRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrUrlRecordNotFound
	}

	// 如果更新slug，检查是否已存在
	if req.Slug != "" && req.Slug != record.Slug {
		existing, err := uc.urlRecordRepo.GetBySlug(ctx, req.Slug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, data.ErrSlugExists
		}
		record.Slug = req.Slug
	}

	record.IsActive = req.IsActive
	record.UpdatedAt = time.Now()

	if err := uc.urlRecordRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteUrlRecord 删除URL记录
func (uc *SEOUseCase) DeleteUrlRecord(ctx context.Context, id uint) error {
	return uc.urlRecordRepo.Delete(ctx, id)
}

// GenerateSlug 生成SEO友好的slug
func (uc *SEOUseCase) GenerateSlug(ctx context.Context, name string) string {
	// 转换为小写
	slug := strings.ToLower(name)

	// 移除特殊字符，只保留字母、数字和空格
	reg := regexp.MustCompile("[^a-z0-9\\s-]")
	slug = reg.ReplaceAllString(slug, "")

	// 将空格和连续的横线替换为单个横线
	slug = strings.ReplaceAll(slug, " ", "-")
	reg2 := regexp.MustCompile("-+")
	slug = reg2.ReplaceAllString(slug, "-")

	// 移除开头和结尾的横线
	slug = strings.Trim(slug, "-")

	// 如果slug为空，使用时间戳
	if slug == "" {
		slug = fmt.Sprintf("item-%d", time.Now().Unix())
	}

	return slug
}

// EnsureUniqueSlug 确保slug唯一
func (uc *SEOUseCase) EnsureUniqueSlug(ctx context.Context, baseSlug string) string {
	slug := baseSlug
	counter := 1

	for {
		existing, err := uc.urlRecordRepo.GetBySlug(ctx, slug)
		if err != nil || existing == nil {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
}

// ========== URL重定向管理 ==========

// CreateUrlRedirect 创建URL重定向
func (uc *SEOUseCase) CreateUrlRedirect(ctx context.Context, req *models.UrlRedirectCreateRequest) (*models.UrlRedirect, error) {
	// 检查oldSlug是否已有重定向
	existing, err := uc.urlRedirectRepo.GetByOldSlug(ctx, req.OldSlug)
	if err == nil && existing != nil {
		return nil, data.ErrOldSlugExists
	}

	// 验证重定向类型
	redirectType := req.RedirectType
	if redirectType != 301 && redirectType != 302 {
		redirectType = 301 // 默认使用301永久重定向
	}

	now := time.Now()
	redirect := &models.UrlRedirect{
		OldSlug:      req.OldSlug,
		NewSlug:      req.NewSlug,
		RedirectType: redirectType,
		IsActive:     req.IsActive,
		HitCount:     0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.urlRedirectRepo.Create(ctx, redirect); err != nil {
		return nil, err
	}

	return redirect, nil
}

// GetUrlRedirect 获取URL重定向
func (uc *SEOUseCase) GetUrlRedirect(ctx context.Context, id uint) (*models.UrlRedirect, error) {
	return uc.urlRedirectRepo.GetByID(ctx, id)
}

// GetRedirectByOldSlug 通过旧URL获取重定向
func (uc *SEOUseCase) GetRedirectByOldSlug(ctx context.Context, oldSlug string) (*models.UrlRedirect, error) {
	redirect, err := uc.urlRedirectRepo.GetByOldSlug(ctx, oldSlug)
	if err != nil {
		return nil, nil // 未找到重定向
	}

	// 增加访问计数
	uc.urlRedirectRepo.IncrementHitCount(ctx, redirect.ID)

	return redirect, nil
}

// ListUrlRedirects URL重定向列表
func (uc *SEOUseCase) ListUrlRedirects(ctx context.Context, page, pageSize int) ([]*models.UrlRedirect, int64, error) {
	return uc.urlRedirectRepo.List(ctx, page, pageSize)
}

// UpdateUrlRedirect 更新URL重定向
func (uc *SEOUseCase) UpdateUrlRedirect(ctx context.Context, id uint, req *models.UrlRedirectUpdateRequest) (*models.UrlRedirect, error) {
	redirect, err := uc.urlRedirectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrUrlRedirectNotFound
	}

	if req.NewSlug != "" {
		redirect.NewSlug = req.NewSlug
	}
	if req.RedirectType == 301 || req.RedirectType == 302 {
		redirect.RedirectType = req.RedirectType
	}
	redirect.IsActive = req.IsActive
	redirect.UpdatedAt = time.Now()

	if err := uc.urlRedirectRepo.Update(ctx, redirect); err != nil {
		return nil, err
	}

	return redirect, nil
}

// DeleteUrlRedirect 删除URL重定向
func (uc *SEOUseCase) DeleteUrlRedirect(ctx context.Context, id uint) error {
	return uc.urlRedirectRepo.Delete(ctx, id)
}

// ========== 元信息管理 ==========

// CreateMetaInfo 创建元信息
func (uc *SEOUseCase) CreateMetaInfo(ctx context.Context, req *models.MetaInfoCreateRequest) (*models.MetaInfo, error) {
	// 检查是否已存在
	existing, err := uc.metaInfoRepo.GetByEntity(ctx, req.EntityID, req.EntityType, req.LanguageID)
	if err == nil && existing != nil {
		// 更新现有记录
		return uc.UpdateMetaInfo(ctx, existing.ID, &models.MetaInfoUpdateRequest{
			MetaTitle:       req.MetaTitle,
			MetaKeywords:    req.MetaKeywords,
			MetaDescription: req.MetaDescription,
			PageTitle:       req.PageTitle,
		})
	}

	now := time.Now()
	meta := &models.MetaInfo{
		EntityID:        req.EntityID,
		EntityType:      req.EntityType,
		LanguageID:      req.LanguageID,
		MetaTitle:       req.MetaTitle,
		MetaKeywords:    req.MetaKeywords,
		MetaDescription: req.MetaDescription,
		PageTitle:       req.PageTitle,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := uc.metaInfoRepo.Create(ctx, meta); err != nil {
		return nil, err
	}

	return meta, nil
}

// GetMetaInfo 获取元信息
func (uc *SEOUseCase) GetMetaInfo(ctx context.Context, id uint) (*models.MetaInfo, error) {
	return uc.metaInfoRepo.GetByID(ctx, id)
}

// GetMetaByEntity 通过实体获取元信息
func (uc *SEOUseCase) GetMetaByEntity(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.MetaInfo, error) {
	return uc.metaInfoRepo.GetByEntity(ctx, entityID, entityType, languageID)
}

// ListMetaInfo 元信息列表
func (uc *SEOUseCase) ListMetaInfo(ctx context.Context, page, pageSize int) ([]*models.MetaInfo, int64, error) {
	return uc.metaInfoRepo.List(ctx, page, pageSize)
}

// UpdateMetaInfo 更新元信息
func (uc *SEOUseCase) UpdateMetaInfo(ctx context.Context, id uint, req *models.MetaInfoUpdateRequest) (*models.MetaInfo, error) {
	meta, err := uc.metaInfoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrMetaInfoNotFound
	}

	meta.MetaTitle = req.MetaTitle
	meta.MetaKeywords = req.MetaKeywords
	meta.MetaDescription = req.MetaDescription
	meta.PageTitle = req.PageTitle
	meta.UpdatedAt = time.Now()

	if err := uc.metaInfoRepo.Update(ctx, meta); err != nil {
		return nil, err
	}

	return meta, nil
}

// DeleteMetaInfo 删除元信息
func (uc *SEOUseCase) DeleteMetaInfo(ctx context.Context, id uint) error {
	return uc.metaInfoRepo.Delete(ctx, id)
}

// ========== Sitemap管理 ==========

// AddSitemapNode 添加Sitemap节点
func (uc *SEOUseCase) AddSitemapNode(ctx context.Context, node *models.SitemapNode) error {
	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now
	return uc.sitemapNodeRepo.Create(ctx, node)
}

// UpdateSitemapNode 更新Sitemap节点
func (uc *SEOUseCase) UpdateSitemapNode(ctx context.Context, node *models.SitemapNode) error {
	node.UpdatedAt = time.Now()
	return uc.sitemapNodeRepo.Update(ctx, node)
}

// DeleteSitemapNode 删除Sitemap节点
func (uc *SEOUseCase) DeleteSitemapNode(ctx context.Context, id uint) error {
	return uc.sitemapNodeRepo.Delete(ctx, id)
}

// ClearSitemapByEntityType 清除指定实体类型的Sitemap节点
func (uc *SEOUseCase) ClearSitemapByEntityType(ctx context.Context, entityType string) error {
	return uc.sitemapNodeRepo.DeleteByEntityType(ctx, entityType)
}

// GetSitemapNodes 获取所有Sitemap节点
func (uc *SEOUseCase) GetSitemapNodes(ctx context.Context) ([]*models.SitemapNode, error) {
	return uc.sitemapNodeRepo.GetAll(ctx)
}

// GenerateSitemap 生成Sitemap XML
func (uc *SEOUseCase) GenerateSitemap(ctx context.Context, baseURL string) (string, *models.SitemapGenerationResult, error) {
	nodes, err := uc.sitemapNodeRepo.GetAll(ctx)
	if err != nil {
		return "", nil, err
	}

	// 生成XML
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	xmlBuilder.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")

	for _, node := range nodes {
		xmlBuilder.WriteString("  <url>\n")
		xmlBuilder.WriteString(fmt.Sprintf("    <loc>%s%s</loc>\n", baseURL, node.URL))
		xmlBuilder.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", node.LastModified.Format("2006-01-02")))
		if node.ChangeFrequency != "" {
			xmlBuilder.WriteString(fmt.Sprintf("    <changefreq>%s</changefreq>\n", node.ChangeFrequency))
		}
		xmlBuilder.WriteString(fmt.Sprintf("    <priority>%0.1f</priority>\n", node.Priority))
		xmlBuilder.WriteString("  </url>\n")
	}

	xmlBuilder.WriteString("</urlset>")

	result := &models.SitemapGenerationResult{
		TotalNodes:  int64(len(nodes)),
		GeneratedAt: time.Now().Format("2006-01-02T15:04:05Z"),
		SitemapURL:  baseURL + "/sitemap.xml",
		FileSize:    int64(len(xmlBuilder.String())),
	}

	return xmlBuilder.String(), result, nil
}

// ========== SEO分析 ==========

// AnalyzeSEO SEO分析
func (uc *SEOUseCase) AnalyzeSEO(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.SEOAnalysisResult, error) {
	// 获取URL记录
	urlRecord, err := uc.urlRecordRepo.GetByEntity(ctx, entityID, entityType, languageID)
	slug := ""
	if err == nil && urlRecord != nil {
		slug = urlRecord.Slug
	}

	// 获取元信息
	meta, err := uc.metaInfoRepo.GetByEntity(ctx, entityID, entityType, languageID)
	result := &models.SEOAnalysisResult{
		EntityID:    entityID,
		EntityType:  entityType,
		Slug:        slug,
		Issues:      []string{},
		Score:       100, // 初始满分
	}

	if meta != nil {
		result.MetaTitle = meta.MetaTitle
		result.MetaKeywords = meta.MetaKeywords
		result.MetaDescription = meta.MetaDescription
		result.TitleLength = len(meta.MetaTitle)
		result.DescriptionLength = len(meta.MetaDescription)
	}

	// 检查SEO问题并扣分
	// 1. 检查slug
	if slug == "" {
		result.Issues = append(result.Issues, "缺少SEO友好的URL")
		result.Score -= 20
	}

	// 2. 检查meta title
	if result.MetaTitle == "" {
		result.Issues = append(result.Issues, "缺少Meta标题")
		result.Score -= 15
	} else if len(result.MetaTitle) < 30 {
		result.Issues = append(result.Issues, "Meta标题太短（建议30-60字符）")
		result.Score -= 5
	} else if len(result.MetaTitle) > 60 {
		result.Issues = append(result.Issues, "Meta标题太长（建议30-60字符）")
		result.Score -= 5
	}

	// 3. 检查meta description
	if result.MetaDescription == "" {
		result.Issues = append(result.Issues, "缺少Meta描述")
		result.Score -= 15
	} else if len(result.MetaDescription) < 100 {
		result.Issues = append(result.Issues, "Meta描述太短（建议100-160字符）")
		result.Score -= 5
	} else if len(result.MetaDescription) > 160 {
		result.Issues = append(result.Issues, "Meta描述太长（建议100-160字符）")
		result.Score -= 5
	}

	// 4. 检查keywords
	if result.MetaKeywords == "" {
		result.Issues = append(result.Issues, "缺少Meta关键词")
		result.Score -= 10
	}

	// 确保分数不低于0
	if result.Score < 0 {
		result.Score = 0
	}

	return result, nil
}