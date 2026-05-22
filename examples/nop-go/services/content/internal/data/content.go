// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/content/internal/biz"

	"gorm.io/gorm"
)

// ==================== 博客 ====================

// BlogPO 博客持久化对象。
type BlogPO struct {
	ID            uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Title         string    `gorm:"size:512;column:title" db:"title" json:"title"`
	Body          string    `gorm:"type:text;column:body" db:"body" json:"body"`
	Tags          string    `gorm:"size:512;column:tags" db:"tags" json:"tags"`
	AllowComments bool      `gorm:"column:allow_comments" db:"allow_comments" json:"allow_comments"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

func (BlogPO) TableName() string { return "blogs" }

func (po *BlogPO) ToEntity() *biz.Blog {
	return &biz.Blog{ID: po.ID, Title: po.Title, Body: po.Body, Tags: po.Tags, AllowComments: po.AllowComments, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt}
}

type blogRepo struct{ db *gorm.DB }

func NewBlogRepo(db *gorm.DB) biz.BlogRepository { return &blogRepo{db: db} }

func (r *blogRepo) Create(ctx context.Context, blog *biz.Blog) error {
	po := &BlogPO{Title: blog.Title, Body: blog.Body, Tags: blog.Tags, AllowComments: blog.AllowComments, CreatedAt: blog.CreatedAt, UpdatedAt: blog.UpdatedAt}
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *blogRepo) GetByID(ctx context.Context, id uint) (*biz.Blog, error) {
	var po BlogPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

func (r *blogRepo) List(ctx context.Context, page, size int) ([]*biz.Blog, int64, error) {
	var pos []BlogPO
	var total int64
	r.db.WithContext(ctx).Model(&BlogPO{}).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Blog, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

func (r *blogRepo) Update(ctx context.Context, blog *biz.Blog) error {
	return r.db.WithContext(ctx).Model(&BlogPO{}).Where("id = ?", blog.ID).Updates(map[string]interface{}{
		"title": blog.Title, "body": blog.Body, "tags": blog.Tags, "allow_comments": blog.AllowComments, "updated_at": blog.UpdatedAt,
	}).Error
}

func (r *blogRepo) Delete(ctx context.Context, id uint) error { return r.db.WithContext(ctx).Delete(&BlogPO{}, id).Error }

// ==================== 新闻 ====================

// NewsPO 新闻持久化对象。
type NewsPO struct {
	ID            uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Title         string    `gorm:"size:512;column:title" db:"title" json:"title"`
	Body          string    `gorm:"type:text;column:body" db:"body" json:"body"`
	AllowComments bool      `gorm:"column:allow_comments" db:"allow_comments" json:"allow_comments"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

func (NewsPO) TableName() string { return "news" }

func (po *NewsPO) ToEntity() *biz.News {
	return &biz.News{ID: po.ID, Title: po.Title, Body: po.Body, AllowComments: po.AllowComments, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt}
}

type newsRepo struct{ db *gorm.DB }

func NewNewsRepo(db *gorm.DB) biz.NewsRepository { return &newsRepo{db: db} }

func (r *newsRepo) Create(ctx context.Context, news *biz.News) error {
	po := &NewsPO{Title: news.Title, Body: news.Body, AllowComments: news.AllowComments, CreatedAt: news.CreatedAt, UpdatedAt: news.UpdatedAt}
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *newsRepo) GetByID(ctx context.Context, id uint) (*biz.News, error) {
	var po NewsPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *newsRepo) List(ctx context.Context, page, size int) ([]*biz.News, int64, error) {
	var pos []NewsPO
	var total int64
	r.db.WithContext(ctx).Model(&NewsPO{}).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil { return nil, 0, err }
	items := make([]*biz.News, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

func (r *newsRepo) Update(ctx context.Context, news *biz.News) error {
	return r.db.WithContext(ctx).Model(&NewsPO{}).Where("id = ?", news.ID).Updates(map[string]interface{}{
		"title": news.Title, "body": news.Body, "allow_comments": news.AllowComments, "updated_at": news.UpdatedAt,
	}).Error
}

func (r *newsRepo) Delete(ctx context.Context, id uint) error { return r.db.WithContext(ctx).Delete(&NewsPO{}, id).Error }

// ==================== 页面 ====================

// TopicPO 页面持久化对象。
type TopicPO struct {
	ID         uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Title      string    `gorm:"size:512;column:title" db:"title" json:"title"`
	Body       string    `gorm:"type:text;column:body" db:"body" json:"body"`
	IsPublished bool      `gorm:"column:is_published" db:"is_published" json:"is_published"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

func (TopicPO) TableName() string { return "topics" }

func (po *TopicPO) ToEntity() *biz.Topic {
	return &biz.Topic{ID: po.ID, Title: po.Title, Body: po.Body, IsPublished: po.IsPublished, CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt}
}

type topicRepo struct{ db *gorm.DB }

func NewTopicRepo(db *gorm.DB) biz.TopicRepository { return &topicRepo{db: db} }

func (r *topicRepo) Create(ctx context.Context, topic *biz.Topic) error {
	po := &TopicPO{Title: topic.Title, Body: topic.Body, IsPublished: topic.IsPublished, CreatedAt: topic.CreatedAt, UpdatedAt: topic.UpdatedAt}
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *topicRepo) GetByID(ctx context.Context, id uint) (*biz.Topic, error) {
	var po TopicPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *topicRepo) List(ctx context.Context, page, size int) ([]*biz.Topic, int64, error) {
	var pos []TopicPO
	var total int64
	r.db.WithContext(ctx).Model(&TopicPO{}).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil { return nil, 0, err }
	items := make([]*biz.Topic, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

func (r *topicRepo) Update(ctx context.Context, topic *biz.Topic) error {
	return r.db.WithContext(ctx).Model(&TopicPO{}).Where("id = ?", topic.ID).Updates(map[string]interface{}{
		"title": topic.Title, "body": topic.Body, "is_published": topic.IsPublished, "updated_at": topic.UpdatedAt,
	}).Error
}

func (r *topicRepo) Delete(ctx context.Context, id uint) error { return r.db.WithContext(ctx).Delete(&TopicPO{}, id).Error }

// ==================== 投票 ====================

// PollPO 投票持久化对象。
type PollPO struct {
	ID                  uint   `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name                string `gorm:"size:256;column:name" db:"name" json:"name"`
	AllowSelectMultiple bool   `gorm:"column:allow_select_multiple" db:"allow_select_multiple" json:"allow_select_multiple"`
}

func (PollPO) TableName() string { return "polls" }

func (po *PollPO) ToEntity() *biz.Poll {
	return &biz.Poll{ID: po.ID, Name: po.Name, AllowSelectMultiple: po.AllowSelectMultiple}
}

// PollAnswerPO 投票选项持久化对象。
type PollAnswerPO struct {
	ID        uint   `gorm:"primaryKey;column:id" db:"id" json:"id"`
	PollID    uint   `gorm:"index;column:poll_id" db:"poll_id" json:"poll_id"`
	Name      string `gorm:"size:256;column:name" db:"name" json:"name"`
	VoteCount int    `gorm:"column:vote_count" db:"vote_count" json:"vote_count"`
}

func (PollAnswerPO) TableName() string { return "poll_answers" }

type pollRepo struct{ db *gorm.DB }

func NewPollRepo(db *gorm.DB) biz.PollRepository { return &pollRepo{db: db} }

func (r *pollRepo) Create(ctx context.Context, poll *biz.Poll) error {
	po := &PollPO{Name: poll.Name, AllowSelectMultiple: poll.AllowSelectMultiple}
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *pollRepo) GetByID(ctx context.Context, id uint) (*biz.Poll, error) {
	var po PollPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *pollRepo) List(ctx context.Context, page, size int) ([]*biz.Poll, int64, error) {
	var pos []PollPO
	var total int64
	r.db.WithContext(ctx).Model(&PollPO{}).Count(&total)
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil { return nil, 0, err }
	items := make([]*biz.Poll, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

func (r *pollRepo) Update(ctx context.Context, poll *biz.Poll) error {
	return r.db.WithContext(ctx).Model(&PollPO{}).Where("id = ?", poll.ID).Updates(map[string]interface{}{
		"name": poll.Name, "allow_select_multiple": poll.AllowSelectMultiple,
	}).Error
}

func (r *pollRepo) Delete(ctx context.Context, id uint) error { return r.db.WithContext(ctx).Delete(&PollPO{}, id).Error }

// IncrementVote 增加投票选项的计数。
func (r *pollRepo) IncrementVote(ctx context.Context, pollID uint, answerID uint) error {
	return r.db.WithContext(ctx).Model(&PollAnswerPO{}).Where("id = ? AND poll_id = ?", answerID, pollID).
		UpdateColumn("vote_count", gorm.Expr("vote_count + 1")).Error
}