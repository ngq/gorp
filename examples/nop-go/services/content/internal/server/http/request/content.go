package request

// CreateBlog 创建博客请求。
type CreateBlog struct {
	Title         string `json:"title" binding:"required"`  // 博客标题
	Body          string `json:"body" binding:"required"`   // 博客正文
	Tags          string `json:"tags"`                      // 标签（逗号分隔）
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// UpdateBlog 更新博客请求。
type UpdateBlog struct {
	Title         string `json:"title" binding:"required"`  // 博客标题
	Body          string `json:"body" binding:"required"`   // 博客正文
	Tags          string `json:"tags"`                      // 标签（逗号分隔）
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// CreateNews 创建新闻请求。
type CreateNews struct {
	Title         string `json:"title" binding:"required"`  // 新闻标题
	Body          string `json:"body" binding:"required"`   // 新闻正文
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// UpdateNews 更新新闻请求。
type UpdateNews struct {
	Title         string `json:"title" binding:"required"`  // 新闻标题
	Body          string `json:"body" binding:"required"`   // 新闻正文
	AllowComments bool   `json:"allow_comments"`            // 是否允许评论
}

// CreateTopic 创建页面请求。
type CreateTopic struct {
	Title       string `json:"title" binding:"required"`  // 页面标题
	Body        string `json:"body" binding:"required"`   // 页面正文
	IsPublished bool   `json:"is_published"`              // 是否发布
}

// UpdateTopic 更新页面请求。
type UpdateTopic struct {
	Title       string `json:"title" binding:"required"`  // 页面标题
	Body        string `json:"body" binding:"required"`   // 页面正文
	IsPublished bool   `json:"is_published"`              // 是否发布
}

// CreatePoll 创建投票请求。
type CreatePoll struct {
	Name               string `json:"name" binding:"required"` // 投票名称
	AllowSelectMultiple bool   `json:"allow_select_multiple"`   // 是否允许多选
}

// UpdatePoll 更新投票请求。
type UpdatePoll struct {
	Name               string `json:"name" binding:"required"` // 投票名称
	AllowSelectMultiple bool   `json:"allow_select_multiple"`   // 是否允许多选
}

// VotePoll 投票请求。
type VotePoll struct {
	AnswerID uint `json:"answer_id" binding:"required"` // 投票选项ID
}