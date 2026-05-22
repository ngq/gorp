// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// ==================== 博客 ====================

// Blog 博客领域实体。
type Blog struct {
	ID            uint      // 博客ID
	Title         string    // 博客标题
	Body          string    // 博客正文
	Tags          string    // 标签（逗号分隔）
	AllowComments bool      // 是否允许评论
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// BlogRepository 博客仓储接口。
type BlogRepository interface {
	Create(ctx context.Context, blog *Blog) error
	GetByID(ctx context.Context, id uint) (*Blog, error)
	List(ctx context.Context, page, size int) ([]*Blog, int64, error)
	Update(ctx context.Context, blog *Blog) error
	Delete(ctx context.Context, id uint) error
}

// BlogUseCase 博客用例。
type BlogUseCase struct {
	repo BlogRepository
}

// NewBlogUseCase 创建博客用例。
func NewBlogUseCase(repo BlogRepository) *BlogUseCase {
	return &BlogUseCase{repo: repo}
}

// Create 创建博客。
func (uc *BlogUseCase) Create(ctx context.Context, title, body, tags string, allowComments bool) (*Blog, error) {
	blog := &Blog{Title: title, Body: body, Tags: tags, AllowComments: allowComments, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := uc.repo.Create(ctx, blog); err != nil {
		return nil, err
	}
	return blog, nil
}

// GetByID 根据ID获取博客。
func (uc *BlogUseCase) GetByID(ctx context.Context, id uint) (*Blog, error) { return uc.repo.GetByID(ctx, id) }

// List 获取博客列表。
func (uc *BlogUseCase) List(ctx context.Context, page, size int) ([]*Blog, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新博客。
func (uc *BlogUseCase) Update(ctx context.Context, id uint, title, body, tags string, allowComments bool) (*Blog, error) {
	blog, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	blog.Title = title
	blog.Body = body
	blog.Tags = tags
	blog.AllowComments = allowComments
	blog.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, blog); err != nil {
		return nil, err
	}
	return blog, nil
}

// Delete 删除博客。
func (uc *BlogUseCase) Delete(ctx context.Context, id uint) error { return uc.repo.Delete(ctx, id) }

// ==================== 新闻 ====================

// News 新闻领域实体。
type News struct {
	ID            uint      // 新闻ID
	Title         string    // 新闻标题
	Body          string    // 新闻正文
	AllowComments bool      // 是否允许评论
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// NewsRepository 新闻仓储接口。
type NewsRepository interface {
	Create(ctx context.Context, news *News) error
	GetByID(ctx context.Context, id uint) (*News, error)
	List(ctx context.Context, page, size int) ([]*News, int64, error)
	Update(ctx context.Context, news *News) error
	Delete(ctx context.Context, id uint) error
}

// NewsUseCase 新闻用例。
type NewsUseCase struct {
	repo NewsRepository
}

// NewNewsUseCase 创建新闻用例。
func NewNewsUseCase(repo NewsRepository) *NewsUseCase {
	return &NewsUseCase{repo: repo}
}

// Create 创建新闻。
func (uc *NewsUseCase) Create(ctx context.Context, title, body string, allowComments bool) (*News, error) {
	news := &News{Title: title, Body: body, AllowComments: allowComments, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := uc.repo.Create(ctx, news); err != nil {
		return nil, err
	}
	return news, nil
}

// GetByID 根据ID获取新闻。
func (uc *NewsUseCase) GetByID(ctx context.Context, id uint) (*News, error) { return uc.repo.GetByID(ctx, id) }

// List 获取新闻列表。
func (uc *NewsUseCase) List(ctx context.Context, page, size int) ([]*News, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新新闻。
func (uc *NewsUseCase) Update(ctx context.Context, id uint, title, body string, allowComments bool) (*News, error) {
	news, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	news.Title = title
	news.Body = body
	news.AllowComments = allowComments
	news.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, news); err != nil {
		return nil, err
	}
	return news, nil
}

// Delete 删除新闻。
func (uc *NewsUseCase) Delete(ctx context.Context, id uint) error { return uc.repo.Delete(ctx, id) }

// ==================== 页面 ====================

// Topic 页面领域实体。
type Topic struct {
	ID         uint      // 页面ID
	Title      string    // 页面标题
	Body       string    // 页面正文
	IsPublished bool      // 是否发布
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
}

// TopicRepository 页面仓储接口。
type TopicRepository interface {
	Create(ctx context.Context, topic *Topic) error
	GetByID(ctx context.Context, id uint) (*Topic, error)
	List(ctx context.Context, page, size int) ([]*Topic, int64, error)
	Update(ctx context.Context, topic *Topic) error
	Delete(ctx context.Context, id uint) error
}

// TopicUseCase 页面用例。
type TopicUseCase struct {
	repo TopicRepository
}

// NewTopicUseCase 创建页面用例。
func NewTopicUseCase(repo TopicRepository) *TopicUseCase {
	return &TopicUseCase{repo: repo}
}

// Create 创建页面。
func (uc *TopicUseCase) Create(ctx context.Context, title, body string, isPublished bool) (*Topic, error) {
	topic := &Topic{Title: title, Body: body, IsPublished: isPublished, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := uc.repo.Create(ctx, topic); err != nil {
		return nil, err
	}
	return topic, nil
}

// GetByID 根据ID获取页面。
func (uc *TopicUseCase) GetByID(ctx context.Context, id uint) (*Topic, error) { return uc.repo.GetByID(ctx, id) }

// List 获取页面列表。
func (uc *TopicUseCase) List(ctx context.Context, page, size int) ([]*Topic, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新页面。
func (uc *TopicUseCase) Update(ctx context.Context, id uint, title, body string, isPublished bool) (*Topic, error) {
	topic, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	topic.Title = title
	topic.Body = body
	topic.IsPublished = isPublished
	topic.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, topic); err != nil {
		return nil, err
	}
	return topic, nil
}

// Delete 删除页面。
func (uc *TopicUseCase) Delete(ctx context.Context, id uint) error { return uc.repo.Delete(ctx, id) }

// ==================== 投票 ====================

// Poll 投票领域实体。
type Poll struct {
	ID                  uint   // 投票ID
	Name                string // 投票名称
	AllowSelectMultiple bool   // 是否允许多选
}

// PollAnswer 投票选项领域实体。
type PollAnswer struct {
	ID       uint   // 选项ID
	PollID   uint   // 所属投票ID
	Name     string // 选项名称
	VoteCount int   // 投票数
}

// PollRepository 投票仓储接口。
type PollRepository interface {
	Create(ctx context.Context, poll *Poll) error
	GetByID(ctx context.Context, id uint) (*Poll, error)
	List(ctx context.Context, page, size int) ([]*Poll, int64, error)
	Update(ctx context.Context, poll *Poll) error
	Delete(ctx context.Context, id uint) error
	IncrementVote(ctx context.Context, pollID uint, answerID uint) error
}

// PollUseCase 投票用例。
type PollUseCase struct {
	repo PollRepository
}

// NewPollUseCase 创建投票用例。
func NewPollUseCase(repo PollRepository) *PollUseCase {
	return &PollUseCase{repo: repo}
}

// Create 创建投票。
func (uc *PollUseCase) Create(ctx context.Context, name string, allowSelectMultiple bool) (*Poll, error) {
	poll := &Poll{Name: name, AllowSelectMultiple: allowSelectMultiple}
	if err := uc.repo.Create(ctx, poll); err != nil {
		return nil, err
	}
	return poll, nil
}

// GetByID 根据ID获取投票。
func (uc *PollUseCase) GetByID(ctx context.Context, id uint) (*Poll, error) { return uc.repo.GetByID(ctx, id) }

// List 获取投票列表。
func (uc *PollUseCase) List(ctx context.Context, page, size int) ([]*Poll, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新投票。
func (uc *PollUseCase) Update(ctx context.Context, id uint, name string, allowSelectMultiple bool) (*Poll, error) {
	poll, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	poll.Name = name
	poll.AllowSelectMultiple = allowSelectMultiple
	if err := uc.repo.Update(ctx, poll); err != nil {
		return nil, err
	}
	return poll, nil
}

// Delete 删除投票。
func (uc *PollUseCase) Delete(ctx context.Context, id uint) error { return uc.repo.Delete(ctx, id) }

// VotePoll 投票，增加选项的投票计数。
func (uc *PollUseCase) VotePoll(ctx context.Context, pollID uint, answerID uint) error {
	return uc.repo.IncrementVote(ctx, pollID, answerID)
}