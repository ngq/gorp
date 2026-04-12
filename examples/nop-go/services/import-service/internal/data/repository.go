// Package data 导入导出服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/import-service/internal/models"

	"gorm.io/gorm"
)

// ImportProfileRepository 导入配置仓储接口
type ImportProfileRepository interface {
	Create(ctx context.Context, profile *models.ImportProfile) error
	GetByID(ctx context.Context, id uint) (*models.ImportProfile, error)
	List(ctx context.Context, page, pageSize int) ([]*models.ImportProfile, int64, error)
	ListByEntityType(ctx context.Context, entityType string) ([]*models.ImportProfile, error)
	Update(ctx context.Context, profile *models.ImportProfile) error
	Delete(ctx context.Context, id uint) error
}

type importProfileRepo struct{ db *gorm.DB }

func NewImportProfileRepository(db *gorm.DB) ImportProfileRepository {
	return &importProfileRepo{db: db}
}

func (r *importProfileRepo) Create(ctx context.Context, p *models.ImportProfile) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *importProfileRepo) GetByID(ctx context.Context, id uint) (*models.ImportProfile, error) {
	var p models.ImportProfile
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *importProfileRepo) List(ctx context.Context, page, pageSize int) ([]*models.ImportProfile, int64, error) {
	var list []*models.ImportProfile
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ImportProfile{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *importProfileRepo) ListByEntityType(ctx context.Context, entityType string) ([]*models.ImportProfile, error) {
	var list []*models.ImportProfile
	err := r.db.WithContext(ctx).Where("entity_type = ? AND is_active = ?", entityType, true).Find(&list).Error
	return list, err
}

func (r *importProfileRepo) Update(ctx context.Context, p *models.ImportProfile) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *importProfileRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ImportProfile{}, id).Error
}

// ExportProfileRepository 导出配置仓储接口
type ExportProfileRepository interface {
	Create(ctx context.Context, profile *models.ExportProfile) error
	GetByID(ctx context.Context, id uint) (*models.ExportProfile, error)
	List(ctx context.Context, page, pageSize int) ([]*models.ExportProfile, int64, error)
	ListByEntityType(ctx context.Context, entityType string) ([]*models.ExportProfile, error)
	Update(ctx context.Context, profile *models.ExportProfile) error
	Delete(ctx context.Context, id uint) error
}

type exportProfileRepo struct{ db *gorm.DB }

func NewExportProfileRepository(db *gorm.DB) ExportProfileRepository {
	return &exportProfileRepo{db: db}
}

func (r *exportProfileRepo) Create(ctx context.Context, p *models.ExportProfile) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *exportProfileRepo) GetByID(ctx context.Context, id uint) (*models.ExportProfile, error) {
	var p models.ExportProfile
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *exportProfileRepo) List(ctx context.Context, page, pageSize int) ([]*models.ExportProfile, int64, error) {
	var list []*models.ExportProfile
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ExportProfile{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *exportProfileRepo) ListByEntityType(ctx context.Context, entityType string) ([]*models.ExportProfile, error) {
	var list []*models.ExportProfile
	err := r.db.WithContext(ctx).Where("entity_type = ? AND is_active = ?", entityType, true).Find(&list).Error
	return list, err
}

func (r *exportProfileRepo) Update(ctx context.Context, p *models.ExportProfile) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *exportProfileRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ExportProfile{}, id).Error
}

// ImportHistoryRepository 导入历史仓储接口
type ImportHistoryRepository interface {
	Create(ctx context.Context, history *models.ImportHistory) error
	GetByID(ctx context.Context, id uint64) (*models.ImportHistory, error)
	List(ctx context.Context, page, pageSize int) ([]*models.ImportHistory, int64, error)
	ListByProfile(ctx context.Context, profileID uint, page, pageSize int) ([]*models.ImportHistory, int64, error)
	Update(ctx context.Context, history *models.ImportHistory) error
}

type importHistoryRepo struct{ db *gorm.DB }

func NewImportHistoryRepository(db *gorm.DB) ImportHistoryRepository {
	return &importHistoryRepo{db: db}
}

func (r *importHistoryRepo) Create(ctx context.Context, h *models.ImportHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *importHistoryRepo) GetByID(ctx context.Context, id uint64) (*models.ImportHistory, error) {
	var h models.ImportHistory
	err := r.db.WithContext(ctx).First(&h, id).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *importHistoryRepo) List(ctx context.Context, page, pageSize int) ([]*models.ImportHistory, int64, error) {
	var list []*models.ImportHistory
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ImportHistory{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *importHistoryRepo) ListByProfile(ctx context.Context, profileID uint, page, pageSize int) ([]*models.ImportHistory, int64, error) {
	var list []*models.ImportHistory
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ImportHistory{}).Where("profile_id = ?", profileID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *importHistoryRepo) Update(ctx context.Context, h *models.ImportHistory) error {
	return r.db.WithContext(ctx).Save(h).Error
}

// ExportHistoryRepository 导出历史仓储接口
type ExportHistoryRepository interface {
	Create(ctx context.Context, history *models.ExportHistory) error
	GetByID(ctx context.Context, id uint64) (*models.ExportHistory, error)
	List(ctx context.Context, page, pageSize int) ([]*models.ExportHistory, int64, error)
	ListByProfile(ctx context.Context, profileID uint, page, pageSize int) ([]*models.ExportHistory, int64, error)
	Update(ctx context.Context, history *models.ExportHistory) error
}

type exportHistoryRepo struct{ db *gorm.DB }

func NewExportHistoryRepository(db *gorm.DB) ExportHistoryRepository {
	return &exportHistoryRepo{db: db}
}

func (r *exportHistoryRepo) Create(ctx context.Context, h *models.ExportHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *exportHistoryRepo) GetByID(ctx context.Context, id uint64) (*models.ExportHistory, error) {
	var h models.ExportHistory
	err := r.db.WithContext(ctx).First(&h, id).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *exportHistoryRepo) List(ctx context.Context, page, pageSize int) ([]*models.ExportHistory, int64, error) {
	var list []*models.ExportHistory
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ExportHistory{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *exportHistoryRepo) ListByProfile(ctx context.Context, profileID uint, page, pageSize int) ([]*models.ExportHistory, int64, error) {
	var list []*models.ExportHistory
	var total int64
	db := r.db.WithContext(ctx).Model(&models.ExportHistory{}).Where("profile_id = ?", profileID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *exportHistoryRepo) Update(ctx context.Context, h *models.ExportHistory) error {
	return r.db.WithContext(ctx).Save(h).Error
}

// ImportErrorRepository 导入错误仓储接口
type ImportErrorRepository interface {
	Create(ctx context.Context, err *models.ImportError) error
	CreateBatch(ctx context.Context, errors []*models.ImportError) error
	GetByHistoryID(ctx context.Context, historyID uint64) ([]*models.ImportError, error)
}

type importErrorRepo struct{ db *gorm.DB }

func NewImportErrorRepository(db *gorm.DB) ImportErrorRepository {
	return &importErrorRepo{db: db}
}

func (r *importErrorRepo) Create(ctx context.Context, e *models.ImportError) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *importErrorRepo) CreateBatch(ctx context.Context, errors []*models.ImportError) error {
	return r.db.WithContext(ctx).Create(errors).Error
}

func (r *importErrorRepo) GetByHistoryID(ctx context.Context, historyID uint64) ([]*models.ImportError, error) {
	var errors []*models.ImportError
	err := r.db.WithContext(ctx).Where("history_id = ?", historyID).Order("row_number").Find(&errors).Error
	return errors, err
}

// 常见错误
var (
	ErrProfileNotFound  = errors.New("profile not found")
	ErrHistoryNotFound  = errors.New("history not found")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrFileNotFound     = errors.New("file not found")
	ErrImportFailed     = errors.New("import failed")
	ErrExportFailed     = errors.New("export failed")
)