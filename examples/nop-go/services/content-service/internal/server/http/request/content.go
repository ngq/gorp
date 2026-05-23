package request

// ==================== 博客请求 ====================

// CreateBlog 创建博客请求
type CreateBlog struct {
	Title   string `json:"title" binding:"required"`   // 标题（必填）
	Content string `json:"content" binding:"required"` // 正文（必填）
	Author  string `json:"author"`                     // 作者
	Status  string `json:"status"`                      // 状态：draft / published / archived
	Tags    string `json:"tags"`                        // 标签
}

// UpdateBlog 更新博客请求
type UpdateBlog struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Author  string `json:"author"`
	Status  string `json:"status"`
	Tags    string `json:"tags"`
}

// ==================== 新闻请求 ====================

// CreateNews 创建新闻请求
type CreateNews struct {
	Title    string `json:"title" binding:"required"`   // 标题（必填）
	Content  string `json:"content" binding:"required"` // 正文（必填）
	Source   string `json:"source"`                     // 新闻来源
	Category string `json:"category"`                   // 新闻分类
	Priority int    `json:"priority"`                   // 优先级
	Status   string `json:"status"`                      // 状态
}

// UpdateNews 更新新闻请求
type UpdateNews struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Source   string `json:"source"`
	Category string `json:"category"`
	Priority int    `json:"priority"`
	Status   string `json:"status"`
}

// ==================== 话题请求 ====================

// CreateTopic 创建话题请求
type CreateTopic struct {
	Title       string `json:"title" binding:"required"` // 标题（必填）
	Description string `json:"description"`              // 描述
	CoverImage  string `json:"cover_image"`              // 封面图 URL
	SortOrder   int    `json:"sort_order"`               // 排序权重
	IsActive    bool   `json:"is_active"`                 // 是否启用
}

// UpdateTopic 更新话题请求
type UpdateTopic struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

// ==================== 投票请求 ====================

// CreatePoll 创建投票请求
type CreatePoll struct {
	Title    string `json:"title" binding:"required"`    // 标题（必填）
	Question string `json:"question" binding:"required"` // 投票问题（必填）
	Options  string `json:"options"`                      // 选项 JSON 数组
	IsActive bool   `json:"is_active"`                    // 是否启用
	EndTime  string `json:"end_time"`                     // 结束时间 RFC3339
}

// UpdatePoll 更新投票请求
type UpdatePoll struct {
	Title    string `json:"title" binding:"required"`
	Question string `json:"question" binding:"required"`
	Options  string `json:"options"`
	IsActive bool   `json:"is_active"`
	EndTime  string `json:"end_time"`
}