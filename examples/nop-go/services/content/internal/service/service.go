package service

import (
	"context"

	"nop-go/services/content/internal/biz"
	"nop-go/services/content/internal/data"

	"gorm.io/gorm"
)

// Services 内容服务集合。
type Services struct {
	Content *ContentService
}

// NewServices 创建内容服务集合。
func NewServices(db *gorm.DB) *Services {
	blogRepo := data.NewBlogRepo(db)
	newsRepo := data.NewNewsRepo(db)
	topicRepo := data.NewTopicRepo(db)
	pollRepo := data.NewPollRepo(db)

	blogUC := biz.NewBlogUseCase(blogRepo)
	newsUC := biz.NewNewsUseCase(newsRepo)
	topicUC := biz.NewTopicUseCase(topicRepo)
	pollUC := biz.NewPollUseCase(pollRepo)

	return &Services{
		Content: &ContentService{
			blogUC:  blogUC,
			newsUC:  newsUC,
			topicUC: topicUC,
			pollUC:  pollUC,
		},
	}
}

// ContentService 内容服务，聚合博客/新闻/页面/投票。
type ContentService struct {
	blogUC  *biz.BlogUseCase
	newsUC  *biz.NewsUseCase
	topicUC *biz.TopicUseCase
	pollUC  *biz.PollUseCase
}

// ==================== 博客请求/响应 ====================

// CreateBlogRequest 创建博客请求。
type CreateBlogRequest struct {
	Title         string `json:"title" binding:"required"`  // 博客标题
	Body          string `json:"body" binding:"required"`   // 博客正文
	Tags          string `json:"tags"`                      // 标签
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// UpdateBlogRequest 更新博客请求。
type UpdateBlogRequest struct {
	Title         string `json:"title" binding:"required"`  // 博客标题
	Body          string `json:"body" binding:"required"`   // 博客正文
	Tags          string `json:"tags"`                      // 标签
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// BlogResponse 博客响应。
type BlogResponse struct {
	ID            uint   `json:"id"`             // 博客ID
	Title         string `json:"title"`          // 博客标题
	Body          string `json:"body"`           // 博客正文
	Tags          string `json:"tags"`           // 标签
	AllowComments bool   `json:"allow_comments"` // 是否允许评论
	CreatedAt     string `json:"created_at"`     // 创建时间
	UpdatedAt     string `json:"updated_at"`     // 更新时间
}

// ==================== 新闻请求/响应 ====================

// CreateNewsRequest 创建新闻请求。
type CreateNewsRequest struct {
	Title         string `json:"title" binding:"required"`  // 新闻标题
	Body          string `json:"body" binding:"required"`   // 新闻正文
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// UpdateNewsRequest 更新新闻请求。
type UpdateNewsRequest struct {
	Title         string `json:"title" binding:"required"`  // 新闻标题
	Body          string `json:"body" binding:"required"`   // 新闻正文
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// NewsResponse 新闻响应。
type NewsResponse struct {
	ID            uint   `json:"id"`             // 新闻ID
	Title         string `json:"title"`          // 新闻标题
	Body          string `json:"body"`           // 新闻正文
	AllowComments bool   `json:"allow_comments"` // 是否允许评论
	CreatedAt     string `json:"created_at"`     // 创建时间
	UpdatedAt     string `json:"updated_at"`     // 更新时间
}

// ==================== 页面请求/响应 ====================

// CreateTopicRequest 创建页面请求。
type CreateTopicRequest struct {
	Title       string `json:"title" binding:"required"`  // 页面标题
	Body        string `json:"body" binding:"required"`   // 页面正文
	IsPublished bool   `json:"is_published"`              // 是否发布
}

// UpdateTopicRequest 更新页面请求。
type UpdateTopicRequest struct {
	Title       string `json:"title" binding:"required"`  // 页面标题
	Body        string `json:"body" binding:"required"`   // 页面正文
	IsPublished bool   `json:"is_published"`              // 是否发布
}

// TopicResponse 页面响应。
type TopicResponse struct {
	ID         uint   `json:"id"`          // 页面ID
	Title      string `json:"title"`       // 页面标题
	Body       string `json:"body"`        // 页面正文
	IsPublished bool   `json:"is_published"` // 是否发布
	CreatedAt  string `json:"created_at"`  // 创建时间
	UpdatedAt  string `json:"updated_at"`  // 更新时间
}

// ==================== 投票请求/响应 ====================

// CreatePollRequest 创建投票请求。
type CreatePollRequest struct {
	Name               string `json:"name" binding:"required"` // 投票名称
	AllowSelectMultiple bool   `json:"allow_select_multiple"`   // 是否允许多选
}

// UpdatePollRequest 更新投票请求。
type UpdatePollRequest struct {
	Name               string `json:"name" binding:"required"` // 投票名称
	AllowSelectMultiple bool   `json:"allow_select_multiple"`   // 是否允许多选
}

// PollResponse 投票响应。
type PollResponse struct {
	ID                  uint   `json:"id"`                    // 投票ID
	Name                string `json:"name"`                  // 投票名称
	AllowSelectMultiple bool   `json:"allow_select_multiple"` // 是否允许多选
}

// ==================== 博客方法 ====================

func (s *ContentService) ListBlog(ctx context.Context, page, size int) ([]BlogResponse, int64, error) {
	blogs, total, err := s.blogUC.List(ctx, page, size)
	if err != nil { return nil, 0, err }
	items := make([]BlogResponse, len(blogs))
	for i, b := range blogs {
		items[i] = BlogResponse{ID: b.ID, Title: b.Title, Body: b.Body, Tags: b.Tags, AllowComments: b.AllowComments, CreatedAt: b.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: b.UpdatedAt.Format("2006-01-02 15:04:05")}
	}
	return items, total, nil
}

func (s *ContentService) CreateBlog(ctx context.Context, req CreateBlogRequest) (*BlogResponse, error) {
	b, err := s.blogUC.Create(ctx, req.Title, req.Body, req.Tags, req.AllowComments)
	if err != nil { return nil, err }
	return &BlogResponse{ID: b.ID, Title: b.Title, Body: b.Body, Tags: b.Tags, AllowComments: b.AllowComments, CreatedAt: b.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: b.UpdatedAt.Format("2006-01-02 15:04:05")}, nil
}

func (s *ContentService) UpdateBlog(ctx context.Context, id uint, req UpdateBlogRequest) (*BlogResponse, error) {
	b, err := s.blogUC.Update(ctx, id, req.Title, req.Body, req.Tags, req.AllowComments)
	if err != nil { return nil, err }
	return &BlogResponse{ID: b.ID, Title: b.Title, Body: b.Body, Tags: b.Tags, AllowComments: b.AllowComments, CreatedAt: b.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: b.UpdatedAt.Format("2006-01-02 15:04:05")}, nil
}

func (s *ContentService) DeleteBlog(ctx context.Context, id uint) error { return s.blogUC.Delete(ctx, id) }

// ==================== 新闻方法 ====================

func (s *ContentService) ListNews(ctx context.Context, page, size int) ([]NewsResponse, int64, error) {
	newsList, total, err := s.newsUC.List(ctx, page, size)
	if err != nil { return nil, 0, err }
	items := make([]NewsResponse, len(newsList))
	for i, n := range newsList {
		items[i] = NewsResponse{ID: n.ID, Title: n.Title, Body: n.Body, AllowComments: n.AllowComments, CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: n.UpdatedAt.Format("2006-01-02 15:04:05")}
	}
	return items, total, nil
}

func (s *ContentService) CreateNews(ctx context.Context, req CreateNewsRequest) (*NewsResponse, error) {
	n, err := s.newsUC.Create(ctx, req.Title, req.Body, req.AllowComments)
	if err != nil { return nil, err }
	return &NewsResponse{ID: n.ID, Title: n.Title, Body: n.Body, AllowComments: n.AllowComments, CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: n.UpdatedAt.Format("2006-01-02 15:04:05")}, nil
}

func (s *ContentService) UpdateNews(ctx context.Context, id uint, req UpdateNewsRequest) (*NewsResponse, error) {
	n, err := s.newsUC.Update(ctx, id, req.Title, req.Body, req.AllowComments)
	if err != nil { return nil, err }
	return &NewsResponse{ID: n.ID, Title: n.Title, Body: n.Body, AllowComments: n.AllowComments, CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: n.UpdatedAt.Format("2006-01-02 15:04:05")}, nil
}

func (s *ContentService) DeleteNews(ctx context.Context, id uint) error { return s.newsUC.Delete(ctx, id) }

// ==================== 页面方法 ====================

func (s *ContentService) ListTopic(ctx context.Context, page, size int) ([]TopicResponse, int64, error) {
	topics, total, err := s.topicUC.List(ctx, page, size)
	if err != nil { return nil, 0, err }
	items := make([]TopicResponse, len(topics))
	for i, t := range topics {
		items[i] = TopicResponse{ID: t.ID, Title: t.Title, Body: t.Body, IsPublished: t.IsPublished, CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05")}
	}
	return items, total, nil
}

func (s *ContentService) CreateTopic(ctx context.Context, req CreateTopicRequest) (*TopicResponse, error) {
	t, err := s.topicUC.Create(ctx, req.Title, req.Body, req.IsPublished)
	if err != nil { return nil, err }
	return &TopicResponse{ID: t.ID, Title: t.Title, Body: t.Body, IsPublished: t.IsPublished, CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05")}, nil
}

func (s *ContentService) UpdateTopic(ctx context.Context, id uint, req UpdateTopicRequest) (*TopicResponse, error) {
	t, err := s.topicUC.Update(ctx, id, req.Title, req.Body, req.IsPublished)
	if err != nil { return nil, err }
	return &TopicResponse{ID: t.ID, Title: t.Title, Body: t.Body, IsPublished: t.IsPublished, CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: t.UpdatedAt.Format("2006-01-02 15:04:05")}, nil
}

func (s *ContentService) DeleteTopic(ctx context.Context, id uint) error { return s.topicUC.Delete(ctx, id) }

// ==================== 投票方法 ====================

func (s *ContentService) ListPoll(ctx context.Context, page, size int) ([]PollResponse, int64, error) {
	polls, total, err := s.pollUC.List(ctx, page, size)
	if err != nil { return nil, 0, err }
	items := make([]PollResponse, len(polls))
	for i, p := range polls {
		items[i] = PollResponse{ID: p.ID, Name: p.Name, AllowSelectMultiple: p.AllowSelectMultiple}
	}
	return items, total, nil
}

func (s *ContentService) CreatePoll(ctx context.Context, req CreatePollRequest) (*PollResponse, error) {
	p, err := s.pollUC.Create(ctx, req.Name, req.AllowSelectMultiple)
	if err != nil { return nil, err }
	return &PollResponse{ID: p.ID, Name: p.Name, AllowSelectMultiple: p.AllowSelectMultiple}, nil
}

func (s *ContentService) UpdatePoll(ctx context.Context, id uint, req UpdatePollRequest) (*PollResponse, error) {
	p, err := s.pollUC.Update(ctx, id, req.Name, req.AllowSelectMultiple)
	if err != nil { return nil, err }
	return &PollResponse{ID: p.ID, Name: p.Name, AllowSelectMultiple: p.AllowSelectMultiple}, nil
}

func (s *ContentService) DeletePoll(ctx context.Context, id uint) error { return s.pollUC.Delete(ctx, id) }

// VotePoll 投票。
func (s *ContentService) VotePoll(ctx context.Context, pollID uint, answerID uint) error {
	return s.pollUC.VotePoll(ctx, pollID, answerID)
}