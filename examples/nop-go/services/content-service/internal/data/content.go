package data

import (
	"context"

	"nop-go/services/content-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================
// PO 结构体映射数据库表，与 biz 层实体解耦

// BlogPO 博客持久化对象
type BlogPO struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	Title     string `gorm:"size:200;not null" db:"title"`
	Content   string `gorm:"type:text;not null" db:"content"`
	Author    string `gorm:"size:100" db:"author"`
	Status    string `gorm:"size:20;default:draft" db:"status"`
	Tags      string `gorm:"size:500" db:"tags"`
	CreatedAt int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定博客表名
func (BlogPO) TableName() string { return "blogs" }

// NewsPO 新闻持久化对象
type NewsPO struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	Title     string `gorm:"size:200;not null" db:"title"`
	Content   string `gorm:"type:text;not null" db:"content"`
	Source    string `gorm:"size:100" db:"source"`
	Category  string `gorm:"size:50" db:"category"`
	Priority  int    `gorm:"default:0" db:"priority"`
	Status    string `gorm:"size:20;default:draft" db:"status"`
	CreatedAt int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定新闻表名
func (NewsPO) TableName() string { return "news" }

// TopicPO 话题持久化对象
type TopicPO struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	Title       string `gorm:"size:200;not null" db:"title"`
	Description string `gorm:"type:text" db:"description"`
	CoverImage  string `gorm:"size:500" db:"cover_image"`
	SortOrder   int    `gorm:"default:0" db:"sort_order"`
	IsActive    bool   `gorm:"default:true" db:"is_active"`
	CreatedAt   int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定话题表名
func (TopicPO) TableName() string { return "topics" }

// PollPO 投票持久化对象
type PollPO struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	Title     string `gorm:"size:200;not null" db:"title"`
	Question  string `gorm:"size:500;not null" db:"question"`
	Options   string `gorm:"type:text" db:"options"`        // JSON 数组
	IsActive  bool   `gorm:"default:true" db:"is_active"`
	StartTime int64  `gorm:"column:start_time" db:"start_time"` // 时间戳
	EndTime   int64  `gorm:"column:end_time" db:"end_time"`
	CreatedAt int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定投票表名
func (PollPO) TableName() string { return "polls" }

// ==================== 仓储实现 ====================

// blogRepo 博客仓储实现
type blogRepo struct {
	db *gorm.DB
}

// NewBlogRepo 创建博客仓储
func NewBlogRepo(db *gorm.DB) biz.BlogRepo {
	return &blogRepo{db: db}
}

func (r *blogRepo) Create(ctx context.Context, blog *biz.Blog) error {
	po := r.toPO(blog)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *blogRepo) GetByID(ctx context.Context, id uint64) (*biz.Blog, error) {
	var po BlogPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *blogRepo) List(ctx context.Context, offset, limit int) ([]*biz.Blog, error) {
	var pos []*BlogPO
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.Blog, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *blogRepo) Update(ctx context.Context, blog *biz.Blog) error {
	po := r.toPO(blog)
	return r.db.WithContext(ctx).Model(&BlogPO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *blogRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&BlogPO{}, id).Error
}

// toPO 将业务实体转换为持久化对象
func (r *blogRepo) toPO(blog *biz.Blog) *BlogPO {
	return &BlogPO{
		ID:        blog.ID,
		Title:     blog.Title,
		Content:   blog.Content,
		Author:    blog.Author,
		Status:    blog.Status,
		Tags:      blog.Tags,
		CreatedAt: blog.CreatedAt.Unix(),
		UpdatedAt: blog.UpdatedAt.Unix(),
	}
}

// toEntity 将持久化对象转换为业务实体
func (r *blogRepo) toEntity(po *BlogPO) *biz.Blog {
	return &biz.Blog{
		ID:        po.ID,
		Title:     po.Title,
		Content:   po.Content,
		Author:    po.Author,
		Status:    po.Status,
		Tags:      po.Tags,
		CreatedAt: unixToTime(po.CreatedAt),
		UpdatedAt: unixToTime(po.UpdatedAt),
	}
}

// newsRepo 新闻仓储实现
type newsRepo struct {
	db *gorm.DB
}

// NewNewsRepo 创建新闻仓储
func NewNewsRepo(db *gorm.DB) biz.NewsRepo {
	return &newsRepo{db: db}
}

func (r *newsRepo) Create(ctx context.Context, news *biz.News) error {
	po := r.toPO(news)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *newsRepo) GetByID(ctx context.Context, id uint64) (*biz.News, error) {
	var po NewsPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *newsRepo) List(ctx context.Context, offset, limit int) ([]*biz.News, error) {
	var pos []*NewsPO
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("priority DESC, id DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.News, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *newsRepo) Update(ctx context.Context, news *biz.News) error {
	po := r.toPO(news)
	return r.db.WithContext(ctx).Model(&NewsPO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *newsRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&NewsPO{}, id).Error
}

func (r *newsRepo) toPO(news *biz.News) *NewsPO {
	return &NewsPO{
		ID:        news.ID,
		Title:     news.Title,
		Content:   news.Content,
		Source:    news.Source,
		Category:  news.Category,
		Priority:  news.Priority,
		Status:    news.Status,
		CreatedAt: news.CreatedAt.Unix(),
		UpdatedAt: news.UpdatedAt.Unix(),
	}
}

func (r *newsRepo) toEntity(po *NewsPO) *biz.News {
	return &biz.News{
		ID:        po.ID,
		Title:     po.Title,
		Content:   po.Content,
		Source:    po.Source,
		Category:  po.Category,
		Priority:  po.Priority,
		Status:    po.Status,
		CreatedAt: unixToTime(po.CreatedAt),
		UpdatedAt: unixToTime(po.UpdatedAt),
	}
}

// topicRepo 话题仓储实现
type topicRepo struct {
	db *gorm.DB
}

// NewTopicRepo 创建话题仓储
func NewTopicRepo(db *gorm.DB) biz.TopicRepo {
	return &topicRepo{db: db}
}

func (r *topicRepo) Create(ctx context.Context, topic *biz.Topic) error {
	po := r.toPO(topic)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *topicRepo) GetByID(ctx context.Context, id uint64) (*biz.Topic, error) {
	var po TopicPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *topicRepo) List(ctx context.Context, offset, limit int) ([]*biz.Topic, error) {
	var pos []*TopicPO
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("sort_order ASC, id DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.Topic, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *topicRepo) Update(ctx context.Context, topic *biz.Topic) error {
	po := r.toPO(topic)
	return r.db.WithContext(ctx).Model(&TopicPO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *topicRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&TopicPO{}, id).Error
}

func (r *topicRepo) toPO(topic *biz.Topic) *TopicPO {
	return &TopicPO{
		ID:          topic.ID,
		Title:       topic.Title,
		Description: topic.Description,
		CoverImage:  topic.CoverImage,
		SortOrder:   topic.SortOrder,
		IsActive:    topic.IsActive,
		CreatedAt:   topic.CreatedAt.Unix(),
		UpdatedAt:   topic.UpdatedAt.Unix(),
	}
}

func (r *topicRepo) toEntity(po *TopicPO) *biz.Topic {
	return &biz.Topic{
		ID:          po.ID,
		Title:       po.Title,
		Description: po.Description,
		CoverImage:  po.CoverImage,
		SortOrder:   po.SortOrder,
		IsActive:    po.IsActive,
		CreatedAt:   unixToTime(po.CreatedAt),
		UpdatedAt:   unixToTime(po.UpdatedAt),
	}
}

// pollRepo 投票仓储实现
type pollRepo struct {
	db *gorm.DB
}

// NewPollRepo 创建投票仓储
func NewPollRepo(db *gorm.DB) biz.PollRepo {
	return &pollRepo{db: db}
}

func (r *pollRepo) Create(ctx context.Context, poll *biz.Poll) error {
	po := r.toPO(poll)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *pollRepo) GetByID(ctx context.Context, id uint64) (*biz.Poll, error) {
	var po PollPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *pollRepo) List(ctx context.Context, offset, limit int) ([]*biz.Poll, error) {
	var pos []*PollPO
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.Poll, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *pollRepo) Update(ctx context.Context, poll *biz.Poll) error {
	po := r.toPO(poll)
	return r.db.WithContext(ctx).Model(&PollPO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *pollRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&PollPO{}, id).Error
}

func (r *pollRepo) toPO(poll *biz.Poll) *PollPO {
	return &PollPO{
		ID:        poll.ID,
		Title:     poll.Title,
		Question:  poll.Question,
		Options:   poll.Options,
		IsActive:  poll.IsActive,
		StartTime: poll.StartTime.Unix(),
		EndTime:   poll.EndTime.Unix(),
		CreatedAt: poll.CreatedAt.Unix(),
		UpdatedAt: poll.UpdatedAt.Unix(),
	}
}

func (r *pollRepo) toEntity(po *PollPO) *biz.Poll {
	return &biz.Poll{
		ID:        po.ID,
		Title:     po.Title,
		Question:  po.Question,
		Options:   po.Options,
		IsActive:  po.IsActive,
		StartTime: unixToTime(po.StartTime),
		EndTime:   unixToTime(po.EndTime),
		CreatedAt: unixToTime(po.CreatedAt),
		UpdatedAt: unixToTime(po.UpdatedAt),
	}
}
