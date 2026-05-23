package service

import (
	"context"
	"time"

	"nop-go/services/content-service/internal/biz"
	"nop-go/services/content-service/internal/data"

	"gorm.io/gorm"
)

// Services 内容服务集合，聚合内容/本地化/推广三大子服务
type Services struct {
	Content      *ContentService
	Localization *LocalizationService
	Affiliate    *AffiliateService
}

// NewServices 创建内容服务集合。
// 通过 db 初始化所有仓储 -> 用例 -> 服务，完成完整依赖注入链。
func NewServices(db *gorm.DB) *Services {
	// ---- 内容子服务 ----
	blogRepo := data.NewBlogRepo(db)
	newsRepo := data.NewNewsRepo(db)
	topicRepo := data.NewTopicRepo(db)
	pollRepo := data.NewPollRepo(db)

	blogUC := biz.NewBlogUseCase(blogRepo)
	newsUC := biz.NewNewsUseCase(newsRepo)
	topicUC := biz.NewTopicUseCase(topicRepo)
	pollUC := biz.NewPollUseCase(pollRepo)

	contentUC := biz.NewContentUseCase(blogUC, newsUC, topicUC, pollUC)

	// ---- 本地化子服务 ----
	langRepo := data.NewLanguageRepo(db)
	resRepo := data.NewLocaleResourceRepo(db)
	locUC := biz.NewLocalizationUseCase(langRepo, resRepo)

	// ---- 推广子服务 ----
	affRepo := data.NewAffiliateRepo(db)
	affOrderRepo := data.NewAffiliateOrderRepo(db)
	affCustRepo := data.NewAffiliateCustomerRepo(db)
	affUC := biz.NewAffiliateUseCase(affRepo, affOrderRepo, affCustRepo)

	return &Services{
		Content:      &ContentService{uc: contentUC},
		Localization: &LocalizationService{uc: locUC},
		Affiliate:    &AffiliateService{uc: affUC},
	}
}

// ========================================================================
// 内容服务
// ========================================================================

// ContentService 内容服务，对外暴露博客/新闻/话题/投票的 CRUD
type ContentService struct {
	uc *biz.ContentUseCase
}

// ---- 博客 ----

// BlogRequest 博客请求（创建和更新共用字段）
type BlogRequest struct {
	Title   string `json:"title" binding:"required"`   // 标题
	Content string `json:"content" binding:"required"` // 正文
	Author  string `json:"author"`                     // 作者
	Status  string `json:"status"`                      // 状态：draft / published / archived
	Tags    string `json:"tags"`                        // 标签
}

// BlogResponse 博客响应
type BlogResponse struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	Status    string `json:"status"`
	Tags      string `json:"tags"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toBlogResponse(b *biz.Blog) *BlogResponse {
	return &BlogResponse{
		ID:        b.ID,
		Title:     b.Title,
		Content:   b.Content,
		Author:    b.Author,
		Status:    b.Status,
		Tags:      b.Tags,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

// ListBlogs 获取博客列表
func (s *ContentService) ListBlogs(ctx context.Context, page, size int) ([]*BlogResponse, int64, error) {
	offset := (page - 1) * size
	blogs, err := s.uc.BlogUC.ListBlogs(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*BlogResponse, 0, len(blogs))
	for _, b := range blogs {
		items = append(items, toBlogResponse(b))
	}
	return items, int64(len(items)), nil
}

// CreateBlog 创建博客
func (s *ContentService) CreateBlog(ctx context.Context, req BlogRequest) (*BlogResponse, error) {
	now := time.Now()
	blog := &biz.Blog{
		Title:     req.Title,
		Content:   req.Content,
		Author:    req.Author,
		Status:    req.Status,
		Tags:      req.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.uc.BlogUC.CreateBlog(ctx, blog); err != nil {
		return nil, err
	}
	return toBlogResponse(blog), nil
}

// GetBlog 获取博客详情
func (s *ContentService) GetBlog(ctx context.Context, id uint64) (*BlogResponse, error) {
	blog, err := s.uc.BlogUC.GetBlog(ctx, id)
	if err != nil {
		return nil, err
	}
	return toBlogResponse(blog), nil
}

// UpdateBlog 更新博客
func (s *ContentService) UpdateBlog(ctx context.Context, id uint64, req BlogRequest) (*BlogResponse, error) {
	blog, err := s.uc.BlogUC.GetBlog(ctx, id)
	if err != nil {
		return nil, err
	}
	blog.Title = req.Title
	blog.Content = req.Content
	blog.Author = req.Author
	blog.Status = req.Status
	blog.Tags = req.Tags
	blog.UpdatedAt = time.Now()
	if err := s.uc.BlogUC.UpdateBlog(ctx, blog); err != nil {
		return nil, err
	}
	return toBlogResponse(blog), nil
}

// DeleteBlog 删除博客
func (s *ContentService) DeleteBlog(ctx context.Context, id uint64) error {
	return s.uc.BlogUC.DeleteBlog(ctx, id)
}

// ---- 新闻 ----

// NewsRequest 新闻请求
type NewsRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Source   string `json:"source"`
	Category string `json:"category"`
	Priority int    `json:"priority"`
	Status   string `json:"status"`
}

// NewsResponse 新闻响应
type NewsResponse struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Source    string `json:"source"`
	Category  string `json:"category"`
	Priority  int    `json:"priority"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toNewsResponse(n *biz.News) *NewsResponse {
	return &NewsResponse{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Source:    n.Source,
		Category:  n.Category,
		Priority:  n.Priority,
		Status:    n.Status,
		CreatedAt: n.CreatedAt.Format(time.DateTime),
		UpdatedAt: n.UpdatedAt.Format(time.DateTime),
	}
}

// ListNews 获取新闻列表
func (s *ContentService) ListNews(ctx context.Context, page, size int) ([]*NewsResponse, int64, error) {
	offset := (page - 1) * size
	newsList, err := s.uc.NewsUC.ListNews(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*NewsResponse, 0, len(newsList))
	for _, n := range newsList {
		items = append(items, toNewsResponse(n))
	}
	return items, int64(len(items)), nil
}

// CreateNews 创建新闻
func (s *ContentService) CreateNews(ctx context.Context, req NewsRequest) (*NewsResponse, error) {
	now := time.Now()
	news := &biz.News{
		Title:     req.Title,
		Content:   req.Content,
		Source:    req.Source,
		Category:  req.Category,
		Priority:  req.Priority,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.uc.NewsUC.CreateNews(ctx, news); err != nil {
		return nil, err
	}
	return toNewsResponse(news), nil
}

// GetNews 获取新闻详情
func (s *ContentService) GetNews(ctx context.Context, id uint64) (*NewsResponse, error) {
	news, err := s.uc.NewsUC.GetNews(ctx, id)
	if err != nil {
		return nil, err
	}
	return toNewsResponse(news), nil
}

// UpdateNews 更新新闻
func (s *ContentService) UpdateNews(ctx context.Context, id uint64, req NewsRequest) (*NewsResponse, error) {
	news, err := s.uc.NewsUC.GetNews(ctx, id)
	if err != nil {
		return nil, err
	}
	news.Title = req.Title
	news.Content = req.Content
	news.Source = req.Source
	news.Category = req.Category
	news.Priority = req.Priority
	news.Status = req.Status
	news.UpdatedAt = time.Now()
	if err := s.uc.NewsUC.UpdateNews(ctx, news); err != nil {
		return nil, err
	}
	return toNewsResponse(news), nil
}

// DeleteNews 删除新闻
func (s *ContentService) DeleteNews(ctx context.Context, id uint64) error {
	return s.uc.NewsUC.DeleteNews(ctx, id)
}

// ---- 话题 ----

// TopicRequest 话题请求
type TopicRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

// TopicResponse 话题响应
type TopicResponse struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toTopicResponse(t *biz.Topic) *TopicResponse {
	return &TopicResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		CoverImage:  t.CoverImage,
		SortOrder:   t.SortOrder,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Format(time.DateTime),
		UpdatedAt:   t.UpdatedAt.Format(time.DateTime),
	}
}

// ListTopics 获取话题列表
func (s *ContentService) ListTopics(ctx context.Context, page, size int) ([]*TopicResponse, int64, error) {
	offset := (page - 1) * size
	topics, err := s.uc.TopicUC.ListTopics(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*TopicResponse, 0, len(topics))
	for _, t := range topics {
		items = append(items, toTopicResponse(t))
	}
	return items, int64(len(items)), nil
}

// CreateTopic 创建话题
func (s *ContentService) CreateTopic(ctx context.Context, req TopicRequest) (*TopicResponse, error) {
	now := time.Now()
	topic := &biz.Topic{
		Title:       req.Title,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.uc.TopicUC.CreateTopic(ctx, topic); err != nil {
		return nil, err
	}
	return toTopicResponse(topic), nil
}

// GetTopic 获取话题详情
func (s *ContentService) GetTopic(ctx context.Context, id uint64) (*TopicResponse, error) {
	topic, err := s.uc.TopicUC.GetTopic(ctx, id)
	if err != nil {
		return nil, err
	}
	return toTopicResponse(topic), nil
}

// UpdateTopic 更新话题
func (s *ContentService) UpdateTopic(ctx context.Context, id uint64, req TopicRequest) (*TopicResponse, error) {
	topic, err := s.uc.TopicUC.GetTopic(ctx, id)
	if err != nil {
		return nil, err
	}
	topic.Title = req.Title
	topic.Description = req.Description
	topic.CoverImage = req.CoverImage
	topic.SortOrder = req.SortOrder
	topic.IsActive = req.IsActive
	topic.UpdatedAt = time.Now()
	if err := s.uc.TopicUC.UpdateTopic(ctx, topic); err != nil {
		return nil, err
	}
	return toTopicResponse(topic), nil
}

// DeleteTopic 删除话题
func (s *ContentService) DeleteTopic(ctx context.Context, id uint64) error {
	return s.uc.TopicUC.DeleteTopic(ctx, id)
}

// ---- 投票 ----

// PollRequest 投票请求
type PollRequest struct {
	Title    string `json:"title" binding:"required"`
	Question string `json:"question" binding:"required"`
	Options  string `json:"options"`          // JSON 数组
	IsActive bool   `json:"is_active"`
	EndTime  string `json:"end_time"`         // RFC3339 格式
}

// PollResponse 投票响应
type PollResponse struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Question  string `json:"question"`
	Options   string `json:"options"`
	IsActive  bool   `json:"is_active"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toPollResponse(p *biz.Poll) *PollResponse {
	return &PollResponse{
		ID:        p.ID,
		Title:     p.Title,
		Question:  p.Question,
		Options:   p.Options,
		IsActive:  p.IsActive,
		StartTime: p.StartTime.Format(time.DateTime),
		EndTime:   p.EndTime.Format(time.DateTime),
		CreatedAt: p.CreatedAt.Format(time.DateTime),
		UpdatedAt: p.UpdatedAt.Format(time.DateTime),
	}
}

// ListPolls 获取投票列表
func (s *ContentService) ListPolls(ctx context.Context, page, size int) ([]*PollResponse, int64, error) {
	offset := (page - 1) * size
	polls, err := s.uc.PollUC.ListPolls(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*PollResponse, 0, len(polls))
	for _, p := range polls {
		items = append(items, toPollResponse(p))
	}
	return items, int64(len(items)), nil
}

// CreatePoll 创建投票
func (s *ContentService) CreatePoll(ctx context.Context, req PollRequest) (*PollResponse, error) {
	now := time.Now()
	endTime := now.Add(30 * 24 * time.Hour) // 默认 30 天
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endTime = t
		}
	}
	poll := &biz.Poll{
		Title:     req.Title,
		Question:  req.Question,
		Options:   req.Options,
		IsActive:  req.IsActive,
		StartTime: now,
		EndTime:   endTime,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.uc.PollUC.CreatePoll(ctx, poll); err != nil {
		return nil, err
	}
	return toPollResponse(poll), nil
}

// GetPoll 获取投票详情
func (s *ContentService) GetPoll(ctx context.Context, id uint64) (*PollResponse, error) {
	poll, err := s.uc.PollUC.GetPoll(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPollResponse(poll), nil
}

// UpdatePoll 更新投票
func (s *ContentService) UpdatePoll(ctx context.Context, id uint64, req PollRequest) (*PollResponse, error) {
	poll, err := s.uc.PollUC.GetPoll(ctx, id)
	if err != nil {
		return nil, err
	}
	poll.Title = req.Title
	poll.Question = req.Question
	poll.Options = req.Options
	poll.IsActive = req.IsActive
	poll.UpdatedAt = time.Now()
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			poll.EndTime = t
		}
	}
	if err := s.uc.PollUC.UpdatePoll(ctx, poll); err != nil {
		return nil, err
	}
	return toPollResponse(poll), nil
}

// DeletePoll 删除投票
func (s *ContentService) DeletePoll(ctx context.Context, id uint64) error {
	return s.uc.PollUC.DeletePoll(ctx, id)
}

// ========================================================================
// 本地化服务
// ========================================================================

// LocalizationService 本地化服务，对外暴露语言/资源 CRUD
type LocalizationService struct {
	uc *biz.LocalizationUseCase
}

// LanguageRequest 语言请求
type LanguageRequest struct {
	Code      string `json:"code" binding:"required"`  // 语言代码
	Name      string `json:"name" binding:"required"`  // 语言名称
	IsDefault bool   `json:"is_default"`               // 是否默认语言
	SortOrder int    `json:"sort_order"`                // 排序权重
	IsActive  bool   `json:"is_active"`                  // 是否启用
}

// LanguageResponse 语言响应
type LanguageResponse struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toLanguageResponse(l *biz.Language) *LanguageResponse {
	return &LanguageResponse{
		ID:        l.ID,
		Code:      l.Code,
		Name:      l.Name,
		IsDefault: l.IsDefault,
		SortOrder: l.SortOrder,
		IsActive:  l.IsActive,
		CreatedAt: l.CreatedAt.Format(time.DateTime),
		UpdatedAt: l.UpdatedAt.Format(time.DateTime),
	}
}

// ListLanguages 获取语言列表
func (s *LocalizationService) ListLanguages(ctx context.Context, page, size int) ([]*LanguageResponse, int64, error) {
	offset := (page - 1) * size
	langs, err := s.uc.ListLanguages(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*LanguageResponse, 0, len(langs))
	for _, l := range langs {
		items = append(items, toLanguageResponse(l))
	}
	return items, int64(len(items)), nil
}

// CreateLanguage 创建语言
func (s *LocalizationService) CreateLanguage(ctx context.Context, req LanguageRequest) (*LanguageResponse, error) {
	now := time.Now()
	lang := &biz.Language{
		Code:      req.Code,
		Name:      req.Name,
		IsDefault: req.IsDefault,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.uc.CreateLanguage(ctx, lang); err != nil {
		return nil, err
	}
	return toLanguageResponse(lang), nil
}

// GetLanguage 获取语言详情
func (s *LocalizationService) GetLanguage(ctx context.Context, id uint64) (*LanguageResponse, error) {
	lang, err := s.uc.GetLanguage(ctx, id)
	if err != nil {
		return nil, err
	}
	return toLanguageResponse(lang), nil
}

// UpdateLanguage 更新语言
func (s *LocalizationService) UpdateLanguage(ctx context.Context, id uint64, req LanguageRequest) (*LanguageResponse, error) {
	lang, err := s.uc.GetLanguage(ctx, id)
	if err != nil {
		return nil, err
	}
	lang.Code = req.Code
	lang.Name = req.Name
	lang.IsDefault = req.IsDefault
	lang.SortOrder = req.SortOrder
	lang.IsActive = req.IsActive
	lang.UpdatedAt = time.Now()
	if err := s.uc.UpdateLanguage(ctx, lang); err != nil {
		return nil, err
	}
	return toLanguageResponse(lang), nil
}

// DeleteLanguage 删除语言
func (s *LocalizationService) DeleteLanguage(ctx context.Context, id uint64) error {
	return s.uc.DeleteLanguage(ctx, id)
}

// LocaleResourceRequest 本地化资源请求
type LocaleResourceRequest struct {
	LanguageID uint64 `json:"language_id" binding:"required"` // 关联语言 ID
	Key        string `json:"key" binding:"required"`         // 翻译键
	Value      string `json:"value" binding:"required"`        // 翻译值
	Module     string `json:"module"`                          // 所属模块
}

// LocaleResourceResponse 本地化资源响应
type LocaleResourceResponse struct {
	ID         uint64 `json:"id"`
	LanguageID uint64 `json:"language_id"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	Module     string `json:"module"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func toLocaleResourceResponse(r *biz.LocaleResource) *LocaleResourceResponse {
	return &LocaleResourceResponse{
		ID:         r.ID,
		LanguageID: r.LanguageID,
		Key:        r.Key,
		Value:      r.Value,
		Module:     r.Module,
		CreatedAt:  r.CreatedAt.Format(time.DateTime),
		UpdatedAt:  r.UpdatedAt.Format(time.DateTime),
	}
}

// ListLocaleResources 获取本地化资源列表
func (s *LocalizationService) ListLocaleResources(ctx context.Context, languageID uint64, page, size int) ([]*LocaleResourceResponse, int64, error) {
	offset := (page - 1) * size
	resources, err := s.uc.ListLocaleResources(ctx, languageID, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*LocaleResourceResponse, 0, len(resources))
	for _, r := range resources {
		items = append(items, toLocaleResourceResponse(r))
	}
	return items, int64(len(items)), nil
}

// CreateLocaleResource 创建本地化资源
func (s *LocalizationService) CreateLocaleResource(ctx context.Context, req LocaleResourceRequest) (*LocaleResourceResponse, error) {
	now := time.Now()
	resource := &biz.LocaleResource{
		LanguageID: req.LanguageID,
		Key:        req.Key,
		Value:      req.Value,
		Module:     req.Module,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.uc.CreateLocaleResource(ctx, resource); err != nil {
		return nil, err
	}
	return toLocaleResourceResponse(resource), nil
}

// GetLocaleResource 获取本地化资源详情
func (s *LocalizationService) GetLocaleResource(ctx context.Context, id uint64) (*LocaleResourceResponse, error) {
	resource, err := s.uc.GetLocaleResource(ctx, id)
	if err != nil {
		return nil, err
	}
	return toLocaleResourceResponse(resource), nil
}

// UpdateLocaleResource 更新本地化资源
func (s *LocalizationService) UpdateLocaleResource(ctx context.Context, id uint64, req LocaleResourceRequest) (*LocaleResourceResponse, error) {
	resource, err := s.uc.GetLocaleResource(ctx, id)
	if err != nil {
		return nil, err
	}
	resource.LanguageID = req.LanguageID
	resource.Key = req.Key
	resource.Value = req.Value
	resource.Module = req.Module
	resource.UpdatedAt = time.Now()
	if err := s.uc.UpdateLocaleResource(ctx, resource); err != nil {
		return nil, err
	}
	return toLocaleResourceResponse(resource), nil
}

// DeleteLocaleResource 删除本地化资源
func (s *LocalizationService) DeleteLocaleResource(ctx context.Context, id uint64) error {
	return s.uc.DeleteLocaleResource(ctx, id)
}

// ========================================================================
// 推广服务
// ========================================================================

// AffiliateService 推广合作服务
type AffiliateService struct {
	uc *biz.AffiliateUseCase
}

// AffiliateRequest 推广合作方请求
type AffiliateRequest struct {
	Name       string  `json:"name" binding:"required"` // 合作方名称
	Code       string  `json:"code" binding:"required"` // 合作方编码
	Contact    string  `json:"contact"`                  // 联系方式
	Website    string  `json:"website"`                  // 网站
	Commission float64 `json:"commission"`               // 佣金比例
	Status     string  `json:"status"`                   // 状态
}

// AffiliateResponse 推广合作方响应
type AffiliateResponse struct {
	ID         uint64  `json:"id"`
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Contact    string  `json:"contact"`
	Website    string  `json:"website"`
	Commission float64 `json:"commission"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func toAffiliateResponse(a *biz.Affiliate) *AffiliateResponse {
	return &AffiliateResponse{
		ID:         a.ID,
		Name:       a.Name,
		Code:       a.Code,
		Contact:    a.Contact,
		Website:    a.Website,
		Commission: a.Commission,
		Status:     a.Status,
		CreatedAt:  a.CreatedAt.Format(time.DateTime),
		UpdatedAt:  a.UpdatedAt.Format(time.DateTime),
	}
}

// AffiliateOrderResponse 推广订单响应
type AffiliateOrderResponse struct {
	ID          uint64  `json:"id"`
	AffiliateID uint64  `json:"affiliate_id"`
	OrderNo     string  `json:"order_no"`
	Amount      float64 `json:"amount"`
	Commission  float64 `json:"commission"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toAffiliateOrderResponse(o *biz.AffiliateOrder) *AffiliateOrderResponse {
	return &AffiliateOrderResponse{
		ID:          o.ID,
		AffiliateID: o.AffiliateID,
		OrderNo:     o.OrderNo,
		Amount:      o.Amount,
		Commission:  o.Commission,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt.Format(time.DateTime),
		UpdatedAt:   o.UpdatedAt.Format(time.DateTime),
	}
}

// AffiliateCustomerResponse 推广客户响应
type AffiliateCustomerResponse struct {
	ID          uint64 `json:"id"`
	AffiliateID uint64 `json:"affiliate_id"`
	CustomerID  uint64 `json:"customer_id"`
	Source       string `json:"source"`
	FirstVisit   string `json:"first_visit"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toAffiliateCustomerResponse(c *biz.AffiliateCustomer) *AffiliateCustomerResponse {
	return &AffiliateCustomerResponse{
		ID:          c.ID,
		AffiliateID: c.AffiliateID,
		CustomerID:  c.CustomerID,
		Source:      c.Source,
		FirstVisit:  c.FirstVisit.Format(time.DateTime),
		CreatedAt:   c.CreatedAt.Format(time.DateTime),
		UpdatedAt:   c.UpdatedAt.Format(time.DateTime),
	}
}

// ListAffiliates 获取推广合作方列表
func (s *AffiliateService) ListAffiliates(ctx context.Context, page, size int) ([]*AffiliateResponse, int64, error) {
	offset := (page - 1) * size
	affiliates, err := s.uc.ListAffiliates(ctx, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*AffiliateResponse, 0, len(affiliates))
	for _, a := range affiliates {
		items = append(items, toAffiliateResponse(a))
	}
	return items, int64(len(items)), nil
}

// CreateAffiliate 创建推广合作方
func (s *AffiliateService) CreateAffiliate(ctx context.Context, req AffiliateRequest) (*AffiliateResponse, error) {
	now := time.Now()
	affiliate := &biz.Affiliate{
		Name:       req.Name,
		Code:       req.Code,
		Contact:    req.Contact,
		Website:    req.Website,
		Commission: req.Commission,
		Status:     req.Status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.uc.CreateAffiliate(ctx, affiliate); err != nil {
		return nil, err
	}
	return toAffiliateResponse(affiliate), nil
}

// GetAffiliate 获取推广合作方详情
func (s *AffiliateService) GetAffiliate(ctx context.Context, id uint64) (*AffiliateResponse, error) {
	affiliate, err := s.uc.GetAffiliate(ctx, id)
	if err != nil {
		return nil, err
	}
	return toAffiliateResponse(affiliate), nil
}

// UpdateAffiliate 更新推广合作方
func (s *AffiliateService) UpdateAffiliate(ctx context.Context, id uint64, req AffiliateRequest) (*AffiliateResponse, error) {
	affiliate, err := s.uc.GetAffiliate(ctx, id)
	if err != nil {
		return nil, err
	}
	affiliate.Name = req.Name
	affiliate.Code = req.Code
	affiliate.Contact = req.Contact
	affiliate.Website = req.Website
	affiliate.Commission = req.Commission
	affiliate.Status = req.Status
	affiliate.UpdatedAt = time.Now()
	if err := s.uc.UpdateAffiliate(ctx, affiliate); err != nil {
		return nil, err
	}
	return toAffiliateResponse(affiliate), nil
}

// DeleteAffiliate 删除推广合作方
func (s *AffiliateService) DeleteAffiliate(ctx context.Context, id uint64) error {
	return s.uc.DeleteAffiliate(ctx, id)
}

// ListOrders 获取推广合作方的订单列表
func (s *AffiliateService) ListOrders(ctx context.Context, affiliateID uint64, page, size int) ([]*AffiliateOrderResponse, int64, error) {
	offset := (page - 1) * size
	orders, err := s.uc.ListOrders(ctx, affiliateID, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*AffiliateOrderResponse, 0, len(orders))
	for _, o := range orders {
		items = append(items, toAffiliateOrderResponse(o))
	}
	return items, int64(len(items)), nil
}

// ListCustomers 获取推广合作方的客户列表
func (s *AffiliateService) ListCustomers(ctx context.Context, affiliateID uint64, page, size int) ([]*AffiliateCustomerResponse, int64, error) {
	offset := (page - 1) * size
	customers, err := s.uc.ListCustomers(ctx, affiliateID, offset, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*AffiliateCustomerResponse, 0, len(customers))
	for _, c := range customers {
		items = append(items, toAffiliateCustomerResponse(c))
	}
	return items, int64(len(items)), nil
}