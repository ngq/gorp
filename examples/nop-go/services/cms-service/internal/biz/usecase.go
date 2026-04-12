// Package biz CMS服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"time"

	"nop-go/services/cms-service/internal/data"
	"nop-go/services/cms-service/internal/models"
)

type BlogUseCase struct {
	postRepo     data.BlogPostRepository
	categoryRepo data.BlogCategoryRepository
}

func NewBlogUseCase(postRepo data.BlogPostRepository, categoryRepo data.BlogCategoryRepository) *BlogUseCase {
	return &BlogUseCase{postRepo: postRepo, categoryRepo: categoryRepo}
}

func (uc *BlogUseCase) CreatePost(ctx context.Context, post *models.BlogPost) error {
	return uc.postRepo.Create(ctx, post)
}

func (uc *BlogUseCase) GetPost(ctx context.Context, id uint64) (*models.BlogPost, error) {
	return uc.postRepo.GetByID(ctx, id)
}

func (uc *BlogUseCase) GetPostBySlug(ctx context.Context, slug string) (*models.BlogPost, error) {
	return uc.postRepo.GetBySlug(ctx, slug)
}

func (uc *BlogUseCase) UpdatePost(ctx context.Context, post *models.BlogPost) error {
	return uc.postRepo.Update(ctx, post)
}

func (uc *BlogUseCase) DeletePost(ctx context.Context, id uint64) error {
	return uc.postRepo.Delete(ctx, id)
}

func (uc *BlogUseCase) ListPosts(ctx context.Context, page, pageSize int) ([]*models.BlogPost, int64, error) {
	return uc.postRepo.List(ctx, page, pageSize)
}

func (uc *BlogUseCase) ListPublishedPosts(ctx context.Context, page, pageSize int) ([]*models.BlogPost, int64, error) {
	return uc.postRepo.ListPublished(ctx, page, pageSize)
}

func (uc *BlogUseCase) PublishPost(ctx context.Context, id uint64) error {
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	post.IsPublished = true
	post.PublishedAt = &now

	return uc.postRepo.Update(ctx, post)
}

func (uc *BlogUseCase) CreateCategory(ctx context.Context, category *models.BlogCategory) error {
	return uc.categoryRepo.Create(ctx, category)
}

func (uc *BlogUseCase) ListCategories(ctx context.Context) ([]*models.BlogCategory, error) {
	return uc.categoryRepo.List(ctx)
}

type NewsUseCase struct {
	newsRepo data.NewsRepository
}

func NewNewsUseCase(newsRepo data.NewsRepository) *NewsUseCase {
	return &NewsUseCase{newsRepo: newsRepo}
}

func (uc *NewsUseCase) CreateNews(ctx context.Context, news *models.News) error {
	return uc.newsRepo.Create(ctx, news)
}

func (uc *NewsUseCase) GetNews(ctx context.Context, id uint64) (*models.News, error) {
	return uc.newsRepo.GetByID(ctx, id)
}

func (uc *NewsUseCase) GetNewsBySlug(ctx context.Context, slug string) (*models.News, error) {
	return uc.newsRepo.GetBySlug(ctx, slug)
}

func (uc *NewsUseCase) UpdateNews(ctx context.Context, news *models.News) error {
	return uc.newsRepo.Update(ctx, news)
}

func (uc *NewsUseCase) DeleteNews(ctx context.Context, id uint64) error {
	return uc.newsRepo.Delete(ctx, id)
}

func (uc *NewsUseCase) ListNews(ctx context.Context, page, pageSize int) ([]*models.News, int64, error) {
	return uc.newsRepo.List(ctx, page, pageSize)
}

type TopicUseCase struct {
	topicRepo data.TopicRepository
}

func NewTopicUseCase(topicRepo data.TopicRepository) *TopicUseCase {
	return &TopicUseCase{topicRepo: topicRepo}
}

func (uc *TopicUseCase) CreateTopic(ctx context.Context, topic *models.Topic) error {
	return uc.topicRepo.Create(ctx, topic)
}

func (uc *TopicUseCase) GetTopic(ctx context.Context, id uint64) (*models.Topic, error) {
	return uc.topicRepo.GetByID(ctx, id)
}

func (uc *TopicUseCase) GetTopicBySlug(ctx context.Context, slug string) (*models.Topic, error) {
	return uc.topicRepo.GetBySlug(ctx, slug)
}

func (uc *TopicUseCase) ListTopics(ctx context.Context) ([]*models.Topic, error) {
	return uc.topicRepo.List(ctx)
}

func (uc *TopicUseCase) UpdateTopic(ctx context.Context, topic *models.Topic) error {
	return uc.topicRepo.Update(ctx, topic)
}

func (uc *TopicUseCase) DeleteTopic(ctx context.Context, id uint64) error {
	return uc.topicRepo.Delete(ctx, id)
}

type ForumUseCase struct {
	forumRepo data.ForumRepository
}

func NewForumUseCase(forumRepo data.ForumRepository) *ForumUseCase {
	return &ForumUseCase{forumRepo: forumRepo}
}

func (uc *ForumUseCase) CreateForum(ctx context.Context, forum *models.Forum) error {
	return uc.forumRepo.Create(ctx, forum)
}

func (uc *ForumUseCase) ListForums(ctx context.Context) ([]*models.Forum, error) {
	return uc.forumRepo.List(ctx)
}

// ========== 菜单用例 ==========

// MenuUseCase 菜单用例
type MenuUseCase struct {
	menuRepo     data.MenuRepository
	menuItemRepo data.MenuItemRepository
}

// NewMenuUseCase 创建菜单用例
func NewMenuUseCase(menuRepo data.MenuRepository, menuItemRepo data.MenuItemRepository) *MenuUseCase {
	return &MenuUseCase{menuRepo: menuRepo, menuItemRepo: menuItemRepo}
}

// CreateMenu 创建菜单
func (uc *MenuUseCase) CreateMenu(ctx context.Context, req *models.MenuCreateRequest) (*models.Menu, error) {
	menu := &models.Menu{
		Name:       req.Name,
		SystemName: req.SystemName,
		Title:      req.Title,
		IsActive:   req.IsActive,
	}
	if err := uc.menuRepo.Create(ctx, menu); err != nil {
		return nil, err
	}
	return menu, nil
}

// GetMenu 获取菜单
func (uc *MenuUseCase) GetMenu(ctx context.Context, id uint64) (*models.Menu, error) {
	return uc.menuRepo.GetByID(ctx, id)
}

// GetMenuBySystemName 通过系统名称获取菜单
func (uc *MenuUseCase) GetMenuBySystemName(ctx context.Context, systemName string) (*models.Menu, error) {
	return uc.menuRepo.GetBySystemName(ctx, systemName)
}

// ListMenus 菜单列表
func (uc *MenuUseCase) ListMenus(ctx context.Context) ([]*models.Menu, error) {
	return uc.menuRepo.List(ctx)
}

// UpdateMenu 更新菜单
func (uc *MenuUseCase) UpdateMenu(ctx context.Context, id uint64, req *models.MenuUpdateRequest) (*models.Menu, error) {
	menu, err := uc.menuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		menu.Name = req.Name
	}
	menu.Title = req.Title
	menu.IsActive = req.IsActive
	if err := uc.menuRepo.Update(ctx, menu); err != nil {
		return nil, err
	}
	return menu, nil
}

// DeleteMenu 删除菜单
func (uc *MenuUseCase) DeleteMenu(ctx context.Context, id uint64) error {
	// 删除菜单项
	if err := uc.menuItemRepo.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	return uc.menuRepo.Delete(ctx, id)
}

// CreateMenuItem 创建菜单项
func (uc *MenuUseCase) CreateMenuItem(ctx context.Context, req *models.MenuItemCreateRequest) (*models.MenuItem, error) {
	item := &models.MenuItem{
		MenuID:       req.MenuID,
		ParentID:     req.ParentID,
		Name:         req.Name,
		URL:          req.URL,
		IconClass:    req.IconClass,
		CssClass:     req.CssClass,
		Target:       req.Target,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	}
	if err := uc.menuItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// GetMenuItems 获取菜单的所有菜单项
func (uc *MenuUseCase) GetMenuItems(ctx context.Context, menuID uint64) ([]*models.MenuItem, error) {
	return uc.menuItemRepo.GetByMenuID(ctx, menuID)
}

// UpdateMenuItem 更新菜单项
func (uc *MenuUseCase) UpdateMenuItem(ctx context.Context, id uint64, req *models.MenuItemUpdateRequest) (*models.MenuItem, error) {
	item, err := uc.menuItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	item.ParentID = req.ParentID
	if req.Name != "" {
		item.Name = req.Name
	}
	item.URL = req.URL
	item.IconClass = req.IconClass
	item.CssClass = req.CssClass
	item.Target = req.Target
	item.DisplayOrder = req.DisplayOrder
	item.IsActive = req.IsActive
	if err := uc.menuItemRepo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteMenuItem 删除菜单项
func (uc *MenuUseCase) DeleteMenuItem(ctx context.Context, id uint64) error {
	return uc.menuItemRepo.Delete(ctx, id)
}

// ========== 投票用例 ==========

// PollUseCase 投票用例
type PollUseCase struct {
	pollRepo    data.PollRepository
	answerRepo  data.PollAnswerRepository
	recordRepo  data.PollVotingRecordRepository
}

// NewPollUseCase 创建投票用例
func NewPollUseCase(pollRepo data.PollRepository, answerRepo data.PollAnswerRepository, recordRepo data.PollVotingRecordRepository) *PollUseCase {
	return &PollUseCase{pollRepo: pollRepo, answerRepo: answerRepo, recordRepo: recordRepo}
}

// CreatePoll 创建投票
func (uc *PollUseCase) CreatePoll(ctx context.Context, req *models.PollCreateRequest) (*models.Poll, error) {
	poll := &models.Poll{
		Name:            req.Name,
		SystemKeyword:   req.SystemKeyword,
		Question:        req.Question,
		ShowOnHomepage:  req.ShowOnHomepage,
		AllowGuestVotes: req.AllowGuestVotes,
		DisplayOrder:    req.DisplayOrder,
		StartDateUtc:    req.StartDateUtc,
		EndDateUtc:      req.EndDateUtc,
		IsActive:        true,
	}
	if err := uc.pollRepo.Create(ctx, poll); err != nil {
		return nil, err
	}

	// 创建选项
	if len(req.Answers) > 0 {
		answers := make([]*models.PollAnswer, len(req.Answers))
		for i, name := range req.Answers {
			answers[i] = &models.PollAnswer{
				PollID:        poll.ID,
				Name:          name,
				DisplayOrder:  i,
			}
		}
		if err := uc.answerRepo.CreateBatch(ctx, answers); err != nil {
			return nil, err
		}
	}

	return poll, nil
}

// GetPoll 获取投票
func (uc *PollUseCase) GetPoll(ctx context.Context, id uint64) (*models.Poll, error) {
	return uc.pollRepo.GetByID(ctx, id)
}

// ListPolls 投票列表
func (uc *PollUseCase) ListPolls(ctx context.Context) ([]*models.Poll, error) {
	return uc.pollRepo.List(ctx)
}

// ListHomepagePolls 获取首页投票
func (uc *PollUseCase) ListHomepagePolls(ctx context.Context) ([]*models.Poll, error) {
	return uc.pollRepo.ListForHomepage(ctx)
}

// DeletePoll 删除投票
func (uc *PollUseCase) DeletePoll(ctx context.Context, id uint64) error {
	// 删除选项
	if err := uc.answerRepo.DeleteByPollID(ctx, id); err != nil {
		return err
	}
	return uc.pollRepo.Delete(ctx, id)
}

// Vote 投票
func (uc *PollUseCase) Vote(ctx context.Context, req *models.PollVoteRequest) error {
	// 获取投票
	poll, err := uc.pollRepo.GetByID(ctx, req.PollID)
	if err != nil {
		return err
	}

	// 检查是否允许游客投票
	if !poll.AllowGuestVotes && req.CustomerID == 0 {
		return ErrGuestVoteNotAllowed
	}

	// 检查是否已投票
	hasVoted, err := uc.recordRepo.HasVoted(ctx, req.CustomerID, req.IPAddress, req.PollID)
	if err != nil {
		return err
	}
	if hasVoted {
		return ErrAlreadyVoted
	}

	// 增加投票计数
	if err := uc.answerRepo.IncrementVoteCount(ctx, req.AnswerID); err != nil {
		return err
	}

	// 记录投票
	record := &models.PollVotingRecord{
		PollAnswerID: req.AnswerID,
		CustomerID:   req.CustomerID,
		IPAddress:    req.IPAddress,
		CreatedOnUtc: time.Now().UTC(),
	}
	return uc.recordRepo.Create(ctx, record)
}

// GetPollResult 获取投票结果
func (uc *PollUseCase) GetPollResult(ctx context.Context, pollID uint64) (*models.PollResult, error) {
	poll, err := uc.pollRepo.GetByID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	answers, err := uc.answerRepo.GetByPollID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	totalVotes := 0
	for _, a := range answers {
		totalVotes += a.NumberOfVotes
	}

	result := &models.PollResult{
		PollID:     poll.ID,
		Question:   poll.Question,
		TotalVotes: totalVotes,
		Answers:    make([]models.PollAnswerResult, len(answers)),
	}

	for i, a := range answers {
		percent := 0.0
		if totalVotes > 0 {
			percent = float64(a.NumberOfVotes) / float64(totalVotes) * 100
		}
		result.Answers[i] = models.PollAnswerResult{
			ID:      a.ID,
			Name:    a.Name,
			Votes:   a.NumberOfVotes,
			Percent: percent,
		}
	}

	return result, nil
}

// ========== HTML内容块用例 ==========

// HtmlBodyUseCase HTML内容块用例
type HtmlBodyUseCase struct {
	htmlRepo data.HtmlBodyRepository
}

// NewHtmlBodyUseCase 创建HTML内容块用例
func NewHtmlBodyUseCase(htmlRepo data.HtmlBodyRepository) *HtmlBodyUseCase {
	return &HtmlBodyUseCase{htmlRepo: htmlRepo}
}

// CreateHtmlBody 创建HTML内容块
func (uc *HtmlBodyUseCase) CreateHtmlBody(ctx context.Context, req *models.HtmlBodyCreateRequest) (*models.HtmlBody, error) {
	html := &models.HtmlBody{
		Name:     req.Name,
		Title:    req.Title,
		Content:  req.Content,
		IsActive: req.IsActive,
	}
	if err := uc.htmlRepo.Create(ctx, html); err != nil {
		return nil, err
	}
	return html, nil
}

// GetHtmlBody 获取HTML内容块
func (uc *HtmlBodyUseCase) GetHtmlBody(ctx context.Context, id uint64) (*models.HtmlBody, error) {
	return uc.htmlRepo.GetByID(ctx, id)
}

// GetHtmlBodyByName 通过名称获取HTML内容块
func (uc *HtmlBodyUseCase) GetHtmlBodyByName(ctx context.Context, name string) (*models.HtmlBody, error) {
	return uc.htmlRepo.GetByName(ctx, name)
}

// ListHtmlBodies HTML内容块列表
func (uc *HtmlBodyUseCase) ListHtmlBodies(ctx context.Context) ([]*models.HtmlBody, error) {
	return uc.htmlRepo.List(ctx)
}

// UpdateHtmlBody 更新HTML内容块
func (uc *HtmlBodyUseCase) UpdateHtmlBody(ctx context.Context, id uint64, req *models.HtmlBodyUpdateRequest) (*models.HtmlBody, error) {
	html, err := uc.htmlRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	html.Title = req.Title
	html.Content = req.Content
	html.IsActive = req.IsActive
	if err := uc.htmlRepo.Update(ctx, html); err != nil {
		return nil, err
	}
	return html, nil
}

// DeleteHtmlBody 删除HTML内容块
func (uc *HtmlBodyUseCase) DeleteHtmlBody(ctx context.Context, id uint64) error {
	return uc.htmlRepo.Delete(ctx, id)
}

// 错误定义
var (
	ErrGuestVoteNotAllowed = errors.New("guest voting not allowed")
	ErrAlreadyVoted        = errors.New("already voted")
)