// Package data CMS服务数据访问层
package data

import (
	"context"

	"nop-go/services/cms-service/internal/models"

	"gorm.io/gorm"
)

type BlogPostRepository interface {
	Create(ctx context.Context, post *models.BlogPost) error
	GetByID(ctx context.Context, id uint64) (*models.BlogPost, error)
	GetBySlug(ctx context.Context, slug string) (*models.BlogPost, error)
	Update(ctx context.Context, post *models.BlogPost) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, page, pageSize int) ([]*models.BlogPost, int64, error)
	ListPublished(ctx context.Context, page, pageSize int) ([]*models.BlogPost, int64, error)
}

type BlogCategoryRepository interface {
	Create(ctx context.Context, category *models.BlogCategory) error
	GetByID(ctx context.Context, id uint64) (*models.BlogCategory, error)
	Update(ctx context.Context, category *models.BlogCategory) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.BlogCategory, error)
}

type NewsRepository interface {
	Create(ctx context.Context, news *models.News) error
	GetByID(ctx context.Context, id uint64) (*models.News, error)
	GetBySlug(ctx context.Context, slug string) (*models.News, error)
	Update(ctx context.Context, news *models.News) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, page, pageSize int) ([]*models.News, int64, error)
}

type TopicRepository interface {
	Create(ctx context.Context, topic *models.Topic) error
	GetByID(ctx context.Context, id uint64) (*models.Topic, error)
	GetBySlug(ctx context.Context, slug string) (*models.Topic, error)
	Update(ctx context.Context, topic *models.Topic) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.Topic, error)
}

type ForumRepository interface {
	Create(ctx context.Context, forum *models.Forum) error
	GetByID(ctx context.Context, id uint64) (*models.Forum, error)
	Update(ctx context.Context, forum *models.Forum) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.Forum, error)
}

type blogPostRepo struct{ db *gorm.DB }

func NewBlogPostRepository(db *gorm.DB) BlogPostRepository {
	return &blogPostRepo{db: db}
}

func (r *blogPostRepo) Create(ctx context.Context, p *models.BlogPost) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *blogPostRepo) GetByID(ctx context.Context, id uint64) (*models.BlogPost, error) {
	var p models.BlogPost
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *blogPostRepo) GetBySlug(ctx context.Context, slug string) (*models.BlogPost, error) {
	var p models.BlogPost
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *blogPostRepo) Update(ctx context.Context, p *models.BlogPost) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *blogPostRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.BlogPost{}, id).Error
}

func (r *blogPostRepo) List(ctx context.Context, page, pageSize int) ([]*models.BlogPost, int64, error) {
	var list []*models.BlogPost
	var total int64
	db := r.db.WithContext(ctx).Model(&models.BlogPost{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *blogPostRepo) ListPublished(ctx context.Context, page, pageSize int) ([]*models.BlogPost, int64, error) {
	var list []*models.BlogPost
	var total int64
	db := r.db.WithContext(ctx).Model(&models.BlogPost{}).Where("is_published = ?", true)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("published_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type blogCategoryRepo struct{ db *gorm.DB }

func NewBlogCategoryRepository(db *gorm.DB) BlogCategoryRepository {
	return &blogCategoryRepo{db: db}
}

func (r *blogCategoryRepo) Create(ctx context.Context, c *models.BlogCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *blogCategoryRepo) GetByID(ctx context.Context, id uint64) (*models.BlogCategory, error) {
	var c models.BlogCategory
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *blogCategoryRepo) Update(ctx context.Context, c *models.BlogCategory) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *blogCategoryRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.BlogCategory{}, id).Error
}

func (r *blogCategoryRepo) List(ctx context.Context) ([]*models.BlogCategory, error) {
	var list []*models.BlogCategory
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

type newsRepo struct{ db *gorm.DB }

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepo{db: db}
}

func (r *newsRepo) Create(ctx context.Context, n *models.News) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *newsRepo) GetByID(ctx context.Context, id uint64) (*models.News, error) {
	var n models.News
	err := r.db.WithContext(ctx).First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepo) GetBySlug(ctx context.Context, slug string) (*models.News, error) {
	var n models.News
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepo) Update(ctx context.Context, n *models.News) error {
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *newsRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.News{}, id).Error
}

func (r *newsRepo) List(ctx context.Context, page, pageSize int) ([]*models.News, int64, error) {
	var list []*models.News
	var total int64
	db := r.db.WithContext(ctx).Model(&models.News{}).Where("is_published = ?", true)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("published_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type topicRepo struct{ db *gorm.DB }

func NewTopicRepository(db *gorm.DB) TopicRepository {
	return &topicRepo{db: db}
}

func (r *topicRepo) Create(ctx context.Context, t *models.Topic) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *topicRepo) GetByID(ctx context.Context, id uint64) (*models.Topic, error) {
	var t models.Topic
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *topicRepo) GetBySlug(ctx context.Context, slug string) (*models.Topic, error) {
	var t models.Topic
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *topicRepo) Update(ctx context.Context, t *models.Topic) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *topicRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Topic{}, id).Error
}

func (r *topicRepo) List(ctx context.Context) ([]*models.Topic, error) {
	var list []*models.Topic
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&list).Error
	return list, err
}

type forumRepo struct{ db *gorm.DB }

func NewForumRepository(db *gorm.DB) ForumRepository {
	return &forumRepo{db: db}
}

func (r *forumRepo) Create(ctx context.Context, f *models.Forum) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *forumRepo) GetByID(ctx context.Context, id uint64) (*models.Forum, error) {
	var f models.Forum
	err := r.db.WithContext(ctx).First(&f, id).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *forumRepo) Update(ctx context.Context, f *models.Forum) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *forumRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Forum{}, id).Error
}

func (r *forumRepo) List(ctx context.Context) ([]*models.Forum, error) {
	var list []*models.Forum
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("sort_order").Find(&list).Error
	return list, err
}

// ========== 菜单仓储 ==========

// MenuRepository 菜单仓储接口
type MenuRepository interface {
	Create(ctx context.Context, menu *models.Menu) error
	GetByID(ctx context.Context, id uint64) (*models.Menu, error)
	GetBySystemName(ctx context.Context, systemName string) (*models.Menu, error)
	List(ctx context.Context) ([]*models.Menu, error)
	Update(ctx context.Context, menu *models.Menu) error
	Delete(ctx context.Context, id uint64) error
}

type menuRepo struct{ db *gorm.DB }

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepo{db: db}
}

func (r *menuRepo) Create(ctx context.Context, m *models.Menu) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *menuRepo) GetByID(ctx context.Context, id uint64) (*models.Menu, error) {
	var m models.Menu
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *menuRepo) GetBySystemName(ctx context.Context, systemName string) (*models.Menu, error) {
	var m models.Menu
	err := r.db.WithContext(ctx).Where("system_name = ?", systemName).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *menuRepo) List(ctx context.Context) ([]*models.Menu, error) {
	var list []*models.Menu
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&list).Error
	return list, err
}

func (r *menuRepo) Update(ctx context.Context, m *models.Menu) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *menuRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Menu{}, id).Error
}

// MenuItemRepository 菜单项仓储接口
type MenuItemRepository interface {
	Create(ctx context.Context, item *models.MenuItem) error
	GetByID(ctx context.Context, id uint64) (*models.MenuItem, error)
	GetByMenuID(ctx context.Context, menuID uint64) ([]*models.MenuItem, error)
	GetByParentID(ctx context.Context, parentID uint64) ([]*models.MenuItem, error)
	Update(ctx context.Context, item *models.MenuItem) error
	Delete(ctx context.Context, id uint64) error
	DeleteByMenuID(ctx context.Context, menuID uint64) error
}

type menuItemRepo struct{ db *gorm.DB }

func NewMenuItemRepository(db *gorm.DB) MenuItemRepository {
	return &menuItemRepo{db: db}
}

func (r *menuItemRepo) Create(ctx context.Context, item *models.MenuItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *menuItemRepo) GetByID(ctx context.Context, id uint64) (*models.MenuItem, error) {
	var item models.MenuItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *menuItemRepo) GetByMenuID(ctx context.Context, menuID uint64) ([]*models.MenuItem, error) {
	var items []*models.MenuItem
	err := r.db.WithContext(ctx).Where("menu_id = ? AND is_active = ?", menuID, true).
		Order("display_order").Find(&items).Error
	return items, err
}

func (r *menuItemRepo) GetByParentID(ctx context.Context, parentID uint64) ([]*models.MenuItem, error) {
	var items []*models.MenuItem
	err := r.db.WithContext(ctx).Where("parent_id = ? AND is_active = ?", parentID, true).
		Order("display_order").Find(&items).Error
	return items, err
}

func (r *menuItemRepo) Update(ctx context.Context, item *models.MenuItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *menuItemRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.MenuItem{}, id).Error
}

func (r *menuItemRepo) DeleteByMenuID(ctx context.Context, menuID uint64) error {
	return r.db.WithContext(ctx).Where("menu_id = ?", menuID).Delete(&models.MenuItem{}).Error
}

// ========== 投票仓储 ==========

// PollRepository 投票仓储接口
type PollRepository interface {
	Create(ctx context.Context, poll *models.Poll) error
	GetByID(ctx context.Context, id uint64) (*models.Poll, error)
	GetBySystemKeyword(ctx context.Context, keyword string) (*models.Poll, error)
	List(ctx context.Context) ([]*models.Poll, error)
	ListForHomepage(ctx context.Context) ([]*models.Poll, error)
	Update(ctx context.Context, poll *models.Poll) error
	Delete(ctx context.Context, id uint64) error
}

type pollRepo struct{ db *gorm.DB }

func NewPollRepository(db *gorm.DB) PollRepository {
	return &pollRepo{db: db}
}

func (r *pollRepo) Create(ctx context.Context, p *models.Poll) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *pollRepo) GetByID(ctx context.Context, id uint64) (*models.Poll, error) {
	var p models.Poll
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pollRepo) GetBySystemKeyword(ctx context.Context, keyword string) (*models.Poll, error) {
	var p models.Poll
	err := r.db.WithContext(ctx).Where("system_keyword = ?", keyword).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pollRepo) List(ctx context.Context) ([]*models.Poll, error) {
	var list []*models.Poll
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("display_order").Find(&list).Error
	return list, err
}

func (r *pollRepo) ListForHomepage(ctx context.Context) ([]*models.Poll, error) {
	var list []*models.Poll
	err := r.db.WithContext(ctx).Where("is_active = ? AND show_on_homepage = ?", true, true).
		Order("display_order").Find(&list).Error
	return list, err
}

func (r *pollRepo) Update(ctx context.Context, p *models.Poll) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *pollRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Poll{}, id).Error
}

// PollAnswerRepository 投票选项仓储接口
type PollAnswerRepository interface {
	Create(ctx context.Context, answer *models.PollAnswer) error
	CreateBatch(ctx context.Context, answers []*models.PollAnswer) error
	GetByID(ctx context.Context, id uint64) (*models.PollAnswer, error)
	GetByPollID(ctx context.Context, pollID uint64) ([]*models.PollAnswer, error)
	Update(ctx context.Context, answer *models.PollAnswer) error
	Delete(ctx context.Context, id uint64) error
	DeleteByPollID(ctx context.Context, pollID uint64) error
	IncrementVoteCount(ctx context.Context, id uint64) error
}

type pollAnswerRepo struct{ db *gorm.DB }

func NewPollAnswerRepository(db *gorm.DB) PollAnswerRepository {
	return &pollAnswerRepo{db: db}
}

func (r *pollAnswerRepo) Create(ctx context.Context, a *models.PollAnswer) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *pollAnswerRepo) CreateBatch(ctx context.Context, answers []*models.PollAnswer) error {
	return r.db.WithContext(ctx).Create(answers).Error
}

func (r *pollAnswerRepo) GetByID(ctx context.Context, id uint64) (*models.PollAnswer, error) {
	var a models.PollAnswer
	err := r.db.WithContext(ctx).First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *pollAnswerRepo) GetByPollID(ctx context.Context, pollID uint64) ([]*models.PollAnswer, error) {
	var answers []*models.PollAnswer
	err := r.db.WithContext(ctx).Where("poll_id = ?", pollID).Order("display_order").Find(&answers).Error
	return answers, err
}

func (r *pollAnswerRepo) Update(ctx context.Context, a *models.PollAnswer) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *pollAnswerRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.PollAnswer{}, id).Error
}

func (r *pollAnswerRepo) DeleteByPollID(ctx context.Context, pollID uint64) error {
	return r.db.WithContext(ctx).Where("poll_id = ?", pollID).Delete(&models.PollAnswer{}).Error
}

func (r *pollAnswerRepo) IncrementVoteCount(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&models.PollAnswer{}).Where("id = ?", id).
		UpdateColumn("number_of_votes", gorm.Expr("number_of_votes + 1")).Error
}

// PollVotingRecordRepository 投票记录仓储接口
type PollVotingRecordRepository interface {
	Create(ctx context.Context, record *models.PollVotingRecord) error
	GetByCustomerAndPoll(ctx context.Context, customerID, pollID uint64) (*models.PollVotingRecord, error)
	GetByIPAndPoll(ctx context.Context, ipAddress string, pollID uint64) (*models.PollVotingRecord, error)
	HasVoted(ctx context.Context, customerID uint64, ipAddress string, pollID uint64) (bool, error)
}

type pollVotingRecordRepo struct{ db *gorm.DB }

func NewPollVotingRecordRepository(db *gorm.DB) PollVotingRecordRepository {
	return &pollVotingRecordRepo{db: db}
}

func (r *pollVotingRecordRepo) Create(ctx context.Context, rec *models.PollVotingRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *pollVotingRecordRepo) GetByCustomerAndPoll(ctx context.Context, customerID, pollID uint64) (*models.PollVotingRecord, error) {
	var rec models.PollVotingRecord
	err := r.db.WithContext(ctx).
		Joins("JOIN poll_answers pa ON pa.id = poll_voting_records.poll_answer_id").
		Where("poll_voting_records.customer_id = ? AND pa.poll_id = ?", customerID, pollID).
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *pollVotingRecordRepo) GetByIPAndPoll(ctx context.Context, ipAddress string, pollID uint64) (*models.PollVotingRecord, error) {
	var rec models.PollVotingRecord
	err := r.db.WithContext(ctx).
		Joins("JOIN poll_answers pa ON pa.id = poll_voting_records.poll_answer_id").
		Where("poll_voting_records.ip_address = ? AND pa.poll_id = ?", ipAddress, pollID).
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *pollVotingRecordRepo) HasVoted(ctx context.Context, customerID uint64, ipAddress string, pollID uint64) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Joins("JOIN poll_answers pa ON pa.id = poll_voting_records.poll_answer_id").
		Where("pa.poll_id = ?", pollID)

	if customerID > 0 {
		query = query.Where("poll_voting_records.customer_id = ?", customerID)
	} else {
		query = query.Where("poll_voting_records.ip_address = ?", ipAddress)
	}

	err := query.Model(&models.PollVotingRecord{}).Count(&count).Error
	return count > 0, err
}

// ========== HTML内容块仓储 ==========

// HtmlBodyRepository HTML内容块仓储接口
type HtmlBodyRepository interface {
	Create(ctx context.Context, html *models.HtmlBody) error
	GetByID(ctx context.Context, id uint64) (*models.HtmlBody, error)
	GetByName(ctx context.Context, name string) (*models.HtmlBody, error)
	List(ctx context.Context) ([]*models.HtmlBody, error)
	Update(ctx context.Context, html *models.HtmlBody) error
	Delete(ctx context.Context, id uint64) error
}

type htmlBodyRepo struct{ db *gorm.DB }

func NewHtmlBodyRepository(db *gorm.DB) HtmlBodyRepository {
	return &htmlBodyRepo{db: db}
}

func (r *htmlBodyRepo) Create(ctx context.Context, h *models.HtmlBody) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *htmlBodyRepo) GetByID(ctx context.Context, id uint64) (*models.HtmlBody, error) {
	var h models.HtmlBody
	err := r.db.WithContext(ctx).First(&h, id).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *htmlBodyRepo) GetByName(ctx context.Context, name string) (*models.HtmlBody, error) {
	var h models.HtmlBody
	err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&h).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *htmlBodyRepo) List(ctx context.Context) ([]*models.HtmlBody, error) {
	var list []*models.HtmlBody
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&list).Error
	return list, err
}

func (r *htmlBodyRepo) Update(ctx context.Context, h *models.HtmlBody) error {
	return r.db.WithContext(ctx).Save(h).Error
}

func (r *htmlBodyRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.HtmlBody{}, id).Error
}