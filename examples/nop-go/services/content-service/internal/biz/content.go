package biz

import (
	"context"
	"time"
)

// ==================== 实体定义 ====================

// Blog 博客实体
type Blog struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Status    string    `json:"status"`     // draft / published / archived
	Tags      string    `json:"tags"`       // 逗号分隔的标签
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// News 新闻实体
type News struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`     // 新闻来源
	Category  string    `json:"category"`   // 新闻分类
	Priority  int       `json:"priority"`   // 优先级，数字越大越靠前
	Status    string    `json:"status"`     // draft / published / archived
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Topic 话题/专题实体
type Topic struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CoverImage  string    `json:"cover_image"` // 封面图 URL
	SortOrder   int       `json:"sort_order"`   // 排序权重
	IsActive    bool      `json:"is_active"`    // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Poll 投票实体
type Poll struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Question  string    `json:"question"`   // 投票问题
	Options   string    `json:"options"`    // JSON 数组，存储选项列表
	IsActive  bool      `json:"is_active"`  // 投票是否进行中
	StartTime time.Time `json:"start_time"` // 投票开始时间
	EndTime   time.Time `json:"end_time"`   // 投票结束时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// BlogRepo 博客仓储接口
type BlogRepo interface {
	// Create 创建博客
	Create(ctx context.Context, blog *Blog) error
	// GetByID 根据 ID 获取博客
	GetByID(ctx context.Context, id uint64) (*Blog, error)
	// List 获取博客列表，支持分页
	List(ctx context.Context, offset, limit int) ([]*Blog, error)
	// Update 更新博客
	Update(ctx context.Context, blog *Blog) error
	// Delete 删除博客
	Delete(ctx context.Context, id uint64) error
}

// NewsRepo 新闻仓储接口
type NewsRepo interface {
	// Create 创建新闻
	Create(ctx context.Context, news *News) error
	// GetByID 根据 ID 获取新闻
	GetByID(ctx context.Context, id uint64) (*News, error)
	// List 获取新闻列表，支持分页
	List(ctx context.Context, offset, limit int) ([]*News, error)
	// Update 更新新闻
	Update(ctx context.Context, news *News) error
	// Delete 删除新闻
	Delete(ctx context.Context, id uint64) error
}

// TopicRepo 话题仓储接口
type TopicRepo interface {
	// Create 创建话题
	Create(ctx context.Context, topic *Topic) error
	// GetByID 根据 ID 获取话题
	GetByID(ctx context.Context, id uint64) (*Topic, error)
	// List 获取话题列表，支持分页
	List(ctx context.Context, offset, limit int) ([]*Topic, error)
	// Update 更新话题
	Update(ctx context.Context, topic *Topic) error
	// Delete 删除话题
	Delete(ctx context.Context, id uint64) error
}

// PollRepo 投票仓储接口
type PollRepo interface {
	// Create 创建投票
	Create(ctx context.Context, poll *Poll) error
	// GetByID 根据 ID 获取投票
	GetByID(ctx context.Context, id uint64) (*Poll, error)
	// List 获取投票列表，支持分页
	List(ctx context.Context, offset, limit int) ([]*Poll, error)
	// Update 更新投票
	Update(ctx context.Context, poll *Poll) error
	// Delete 删除投票
	Delete(ctx context.Context, id uint64) error
}

// ==================== 用例 ====================

// BlogUseCase 博客业务用例
type BlogUseCase struct {
	repo BlogRepo
}

// NewBlogUseCase 创建博客用例
func NewBlogUseCase(repo BlogRepo) *BlogUseCase {
	return &BlogUseCase{repo: repo}
}

// CreateBlog 创建博客
func (uc *BlogUseCase) CreateBlog(ctx context.Context, blog *Blog) error {
	return uc.repo.Create(ctx, blog)
}

// GetBlog 获取博客详情
func (uc *BlogUseCase) GetBlog(ctx context.Context, id uint64) (*Blog, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListBlogs 获取博客列表
func (uc *BlogUseCase) ListBlogs(ctx context.Context, offset, limit int) ([]*Blog, error) {
	return uc.repo.List(ctx, offset, limit)
}

// UpdateBlog 更新博客
func (uc *BlogUseCase) UpdateBlog(ctx context.Context, blog *Blog) error {
	return uc.repo.Update(ctx, blog)
}

// DeleteBlog 删除博客
func (uc *BlogUseCase) DeleteBlog(ctx context.Context, id uint64) error {
	return uc.repo.Delete(ctx, id)
}

// NewsUseCase 新闻业务用例
type NewsUseCase struct {
	repo NewsRepo
}

// NewNewsUseCase 创建新闻用例
func NewNewsUseCase(repo NewsRepo) *NewsUseCase {
	return &NewsUseCase{repo: repo}
}

// CreateNews 创建新闻
func (uc *NewsUseCase) CreateNews(ctx context.Context, news *News) error {
	return uc.repo.Create(ctx, news)
}

// GetNews 获取新闻详情
func (uc *NewsUseCase) GetNews(ctx context.Context, id uint64) (*News, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListNews 获取新闻列表
func (uc *NewsUseCase) ListNews(ctx context.Context, offset, limit int) ([]*News, error) {
	return uc.repo.List(ctx, offset, limit)
}

// UpdateNews 更新新闻
func (uc *NewsUseCase) UpdateNews(ctx context.Context, news *News) error {
	return uc.repo.Update(ctx, news)
}

// DeleteNews 删除新闻
func (uc *NewsUseCase) DeleteNews(ctx context.Context, id uint64) error {
	return uc.repo.Delete(ctx, id)
}

// TopicUseCase 话题业务用例
type TopicUseCase struct {
	repo TopicRepo
}

// NewTopicUseCase 创建话题用例
func NewTopicUseCase(repo TopicRepo) *TopicUseCase {
	return &TopicUseCase{repo: repo}
}

// CreateTopic 创建话题
func (uc *TopicUseCase) CreateTopic(ctx context.Context, topic *Topic) error {
	return uc.repo.Create(ctx, topic)
}

// GetTopic 获取话题详情
func (uc *TopicUseCase) GetTopic(ctx context.Context, id uint64) (*Topic, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListTopics 获取话题列表
func (uc *TopicUseCase) ListTopics(ctx context.Context, offset, limit int) ([]*Topic, error) {
	return uc.repo.List(ctx, offset, limit)
}

// UpdateTopic 更新话题
func (uc *TopicUseCase) UpdateTopic(ctx context.Context, topic *Topic) error {
	return uc.repo.Update(ctx, topic)
}

// DeleteTopic 删除话题
func (uc *TopicUseCase) DeleteTopic(ctx context.Context, id uint64) error {
	return uc.repo.Delete(ctx, id)
}

// PollUseCase 投票业务用例
type PollUseCase struct {
	repo PollRepo
}

// NewPollUseCase 创建投票用例
func NewPollUseCase(repo PollRepo) *PollUseCase {
	return &PollUseCase{repo: repo}
}

// CreatePoll 创建投票
func (uc *PollUseCase) CreatePoll(ctx context.Context, poll *Poll) error {
	return uc.repo.Create(ctx, poll)
}

// GetPoll 获取投票详情
func (uc *PollUseCase) GetPoll(ctx context.Context, id uint64) (*Poll, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListPolls 获取投票列表
func (uc *PollUseCase) ListPolls(ctx context.Context, offset, limit int) ([]*Poll, error) {
	return uc.repo.List(ctx, offset, limit)
}

// UpdatePoll 更新投票
func (uc *PollUseCase) UpdatePoll(ctx context.Context, poll *Poll) error {
	return uc.repo.Update(ctx, poll)
}

// DeletePoll 删除投票
func (uc *PollUseCase) DeletePoll(ctx context.Context, id uint64) error {
	return uc.repo.Delete(ctx, id)
}

// ==================== 聚合用例 ====================

// ContentUseCase 内容聚合用例，组合博客/新闻/话题/投票四个子用例
type ContentUseCase struct {
	BlogUC  *BlogUseCase
	NewsUC  *NewsUseCase
	TopicUC *TopicUseCase
	PollUC  *PollUseCase
}

// NewContentUseCase 创建内容聚合用例
func NewContentUseCase(
	blogUC *BlogUseCase,
	newsUC *NewsUseCase,
	topicUC *TopicUseCase,
	pollUC *PollUseCase,
) *ContentUseCase {
	return &ContentUseCase{
		BlogUC:  blogUC,
		NewsUC:  newsUC,
		TopicUC: topicUC,
		PollUC:  pollUC,
	}
}
