// Package models CMS服务数据模型
package models

import (
	"time"
)

// BlogPost 博客文章
type BlogPost struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	Title         string     `gorm:"size:256;not null" json:"title"`
	Slug          string     `gorm:"size:256;uniqueIndex" json:"slug"`
	Content       string     `gorm:"type:longtext" json:"content"`
	Summary       string     `gorm:"type:text" json:"summary"`
	AuthorID      uint64     `json:"author_id"`
	CategoryID    uint64     `gorm:"index" json:"category_id"`
	Tags          string     `gorm:"type:json" json:"tags"` // JSON array
	CoverImageURL string     `gorm:"size:512" json:"cover_image_url"`
	IsPublished   bool       `gorm:"default:false;index" json:"is_published"`
	ViewCount     int        `gorm:"default:0" json:"view_count"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (BlogPost) TableName() string {
	return "blog_posts"
}

// BlogCategory 博客分类
type BlogCategory struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Slug        string    `gorm:"size:128;uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BlogCategory) TableName() string {
	return "blog_categories"
}

// News 新闻
type News struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	Title         string     `gorm:"size:256;not null" json:"title"`
	Slug          string     `gorm:"size:256;uniqueIndex" json:"slug"`
	Content       string     `gorm:"type:longtext" json:"content"`
	Summary       string     `gorm:"type:text" json:"summary"`
	CoverImageURL string     `gorm:"size:512" json:"cover_image_url"`
	IsPublished   bool       `gorm:"default:false;index" json:"is_published"`
	ViewCount     int        `gorm:"default:0" json:"view_count"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (News) TableName() string {
	return "news"
}

// Topic 页面
type Topic struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:256;not null" json:"title"`
	Slug      string    `gorm:"size:256;uniqueIndex" json:"slug"`
	Content   string    `gorm:"type:longtext" json:"content"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Topic) TableName() string {
	return "topics"
}

// Forum 论坛
type Forum struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Slug        string    `gorm:"size:128;uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *uint64   `gorm:"index" json:"parent_id"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Forum) TableName() string {
	return "forums"
}

// ForumTopic 论坛主题
type ForumTopic struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	ForumID     uint64     `gorm:"not null;index" json:"forum_id"`
	Title       string     `gorm:"size:256;not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	AuthorID    uint64     `gorm:"not null;index" json:"author_id"`
	IsPinned    bool       `gorm:"default:false" json:"is_pinned"`
	IsLocked    bool       `gorm:"default:false" json:"is_locked"`
	ViewCount   int        `gorm:"default:0" json:"view_count"`
	ReplyCount  int        `gorm:"default:0" json:"reply_count"`
	LastReplyAt *time.Time `json:"last_reply_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ForumTopic) TableName() string {
	return "forum_topics"
}

// ForumPost 论坛回复
type ForumPost struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TopicID   uint64    `gorm:"not null;index" json:"topic_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	AuthorID  uint64    `gorm:"not null;index" json:"author_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ForumPost) TableName() string {
	return "forum_posts"
}

// DTO
type BlogPostResponse struct {
	ID            uint64 `json:"id"`
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	Summary       string `json:"summary"`
	Content       string `json:"content"`
	CoverImageURL string `json:"cover_image_url"`
	Author        string `json:"author"`
	Category      string `json:"category"`
	ViewCount     int    `json:"view_count"`
	PublishedAt   string `json:"published_at"`
}

func ToBlogPostResponse(p *BlogPost) BlogPostResponse {
	resp := BlogPostResponse{
		ID:            p.ID,
		Title:         p.Title,
		Slug:          p.Slug,
		Summary:       p.Summary,
		Content:       p.Content,
		CoverImageURL: p.CoverImageURL,
		ViewCount:     p.ViewCount,
	}
	if p.PublishedAt != nil {
		resp.PublishedAt = p.PublishedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}

// ========== 菜单管理 ==========

// Menu 导航菜单
type Menu struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	SystemName  string    `gorm:"size:128;uniqueIndex" json:"system_name"` // 系统标识
	Title       string    `gorm:"size:256" json:"title"`                   // 显示标题
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}

// MenuItem 菜单项
type MenuItem struct {
	ID           uint64  `gorm:"primaryKey" json:"id"`
	MenuID       uint64  `gorm:"not null;index" json:"menu_id"`
	ParentID     *uint64 `gorm:"index" json:"parent_id"`           // 父菜单项ID（支持多级菜单）
	Name         string  `gorm:"size:128;not null" json:"name"`    // 菜单项名称
	URL          string  `gorm:"size:512" json:"url"`              // 链接地址
	IconClass    string  `gorm:"size:64" json:"icon_class"`        // 图标CSS类
	CssClass     string  `gorm:"size:128" json:"css_class"`        // 自定义CSS类
	Target       string  `gorm:"size:16" json:"target"`            // 打开方式：_self, _blank
	DisplayOrder int     `gorm:"default:0" json:"display_order"`   // 排序
	IsActive     bool    `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (MenuItem) TableName() string {
	return "menu_items"
}

// ========== 投票管理 ==========

// Poll 投票
type Poll struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	Name             string     `gorm:"size:256;not null" json:"name"`
	SystemKeyword    string     `gorm:"size:128;uniqueIndex" json:"system_keyword"` // 系统关键字
	Question         string     `gorm:"size:512;not null" json:"question"`          // 投票问题
	ShowOnHomepage   bool       `gorm:"default:false" json:"show_on_homepage"`      // 是否显示在首页
	AllowGuestVotes  bool       `gorm:"default:true" json:"allow_guest_votes"`      // 是否允许游客投票
	DisplayOrder     int        `gorm:"default:0" json:"display_order"`
	IsActive         bool       `gorm:"default:true" json:"is_active"`
	StartDateUtc     *time.Time `json:"start_date_utc"`
	EndDateUtc       *time.Time `json:"end_date_utc"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Poll) TableName() string {
	return "polls"
}

// PollAnswer 投票选项
type PollAnswer struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	PollID        uint64    `gorm:"not null;index" json:"poll_id"`
	Name          string    `gorm:"size:256;not null" json:"name"`       // 选项名称
	NumberOfVotes int       `gorm:"default:0" json:"number_of_votes"`    // 投票数
	DisplayOrder  int       `gorm:"default:0" json:"display_order"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (PollAnswer) TableName() string {
	return "poll_answers"
}

// PollVotingRecord 投票记录
type PollVotingRecord struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	PollAnswerID uint64    `gorm:"not null;index" json:"poll_answer_id"`
	CustomerID   uint64    `gorm:"index" json:"customer_id"`      // 投票用户ID（游客为0）
	IPAddress    string    `gorm:"size:50" json:"ip_address"`     // 投票IP
	CreatedOnUtc time.Time `json:"created_on_utc"`                // 投票时间
}

// TableName 指定表名
func (PollVotingRecord) TableName() string {
	return "poll_voting_records"
}

// ========== HTML内容块 ==========

// HtmlBody HTML内容块
type HtmlBody struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:256;not null;uniqueIndex" json:"name"` // 内容块名称
	Title       string    `gorm:"size:256" json:"title"`                     // 显示标题
	Content     string    `gorm:"type:longtext" json:"content"`              // HTML内容
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (HtmlBody) TableName() string {
	return "html_bodies"
}

// ========== DTO ==========

// MenuCreateRequest 菜单创建请求
type MenuCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	SystemName string `json:"system_name" binding:"required"`
	Title      string `json:"title"`
	IsActive   bool   `json:"is_active"`
}

// MenuUpdateRequest 菜单更新请求
type MenuUpdateRequest struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"`
}

// MenuItemCreateRequest 菜单项创建请求
type MenuItemCreateRequest struct {
	MenuID       uint64  `json:"menu_id" binding:"required"`
	ParentID     *uint64 `json:"parent_id"`
	Name         string  `json:"name" binding:"required"`
	URL          string  `json:"url"`
	IconClass    string  `json:"icon_class"`
	CssClass     string  `json:"css_class"`
	Target       string  `json:"target"`
	DisplayOrder int     `json:"display_order"`
	IsActive     bool    `json:"is_active"`
}

// MenuItemUpdateRequest 菜单项更新请求
type MenuItemUpdateRequest struct {
	ParentID     *uint64 `json:"parent_id"`
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	IconClass    string  `json:"icon_class"`
	CssClass     string  `json:"css_class"`
	Target       string  `json:"target"`
	DisplayOrder int     `json:"display_order"`
	IsActive     bool    `json:"is_active"`
}

// PollCreateRequest 投票创建请求
type PollCreateRequest struct {
	Name            string     `json:"name" binding:"required"`
	SystemKeyword   string     `json:"system_keyword" binding:"required"`
	Question        string     `json:"question" binding:"required"`
	ShowOnHomepage  bool       `json:"show_on_homepage"`
	AllowGuestVotes bool       `json:"allow_guest_votes"`
	DisplayOrder    int        `json:"display_order"`
	StartDateUtc    *time.Time `json:"start_date_utc"`
	EndDateUtc      *time.Time `json:"end_date_utc"`
	Answers         []string   `json:"answers"` // 选项列表
}

// PollVoteRequest 投票请求
type PollVoteRequest struct {
	PollID       uint64 `json:"poll_id" binding:"required"`
	AnswerID     uint64 `json:"answer_id" binding:"required"`
	CustomerID   uint64 `json:"customer_id"`
	IPAddress    string `json:"ip_address"`
}

// PollResult 投票结果
type PollResult struct {
	PollID      uint64         `json:"poll_id"`
	Question    string         `json:"question"`
	TotalVotes  int            `json:"total_votes"`
	Answers     []PollAnswerResult `json:"answers"`
}

// PollAnswerResult 投票选项结果
type PollAnswerResult struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Votes     int     `json:"votes"`
	Percent   float64 `json:"percent"`
}

// HtmlBodyCreateRequest HTML内容块创建请求
type HtmlBodyCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	IsActive bool   `json:"is_active"`
}

// HtmlBodyUpdateRequest HTML内容块更新请求
type HtmlBodyUpdateRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	IsActive bool   `json:"is_active"`
}