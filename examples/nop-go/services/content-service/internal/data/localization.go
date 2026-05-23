package data

import (
	"context"

	"nop-go/services/content-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================

// LanguagePO 语言持久化对象
type LanguagePO struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	Code      string `gorm:"size:10;uniqueIndex;not null" db:"code"` // 语言代码唯一
	Name      string `gorm:"size:100;not null" db:"name"`
	IsDefault bool   `gorm:"default:false" db:"is_default"`
	SortOrder int    `gorm:"default:0" db:"sort_order"`
	IsActive  bool   `gorm:"default:true" db:"is_active"`
	CreatedAt int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定语言表名
func (LanguagePO) TableName() string { return "languages" }

// LocaleResourcePO 本地化资源持久化对象
type LocaleResourcePO struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	LanguageID uint64 `gorm:"index;not null" db:"language_id"`           // 关联语言 ID
	Key        string `gorm:"size:200;not null" db:"key"`        // 翻译键
	Value      string `gorm:"type:text;not null" db:"value"`       // 翻译值
	Module     string `gorm:"size:50;index" db:"module"`            // 所属模块
	CreatedAt  int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定本地化资源表名
func (LocaleResourcePO) TableName() string { return "locale_resources" }

// ==================== 仓储实现 ====================

// languageRepo 语言仓储实现
type languageRepo struct {
	db *gorm.DB
}

// NewLanguageRepo 创建语言仓储
func NewLanguageRepo(db *gorm.DB) biz.LanguageRepo {
	return &languageRepo{db: db}
}

func (r *languageRepo) Create(ctx context.Context, lang *biz.Language) error {
	po := r.toPO(lang)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *languageRepo) GetByID(ctx context.Context, id uint64) (*biz.Language, error) {
	var po LanguagePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *languageRepo) GetByCode(ctx context.Context, code string) (*biz.Language, error) {
	var po LanguagePO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *languageRepo) List(ctx context.Context, offset, limit int) ([]*biz.Language, error) {
	var pos []*LanguagePO
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("sort_order ASC, id ASC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.Language, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *languageRepo) Update(ctx context.Context, lang *biz.Language) error {
	po := r.toPO(lang)
	return r.db.WithContext(ctx).Model(&LanguagePO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *languageRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&LanguagePO{}, id).Error
}

func (r *languageRepo) toPO(lang *biz.Language) *LanguagePO {
	return &LanguagePO{
		ID:        lang.ID,
		Code:      lang.Code,
		Name:      lang.Name,
		IsDefault: lang.IsDefault,
		SortOrder: lang.SortOrder,
		IsActive:  lang.IsActive,
		CreatedAt: lang.CreatedAt.Unix(),
		UpdatedAt: lang.UpdatedAt.Unix(),
	}
}

func (r *languageRepo) toEntity(po *LanguagePO) *biz.Language {
	return &biz.Language{
		ID:        po.ID,
		Code:      po.Code,
		Name:      po.Name,
		IsDefault: po.IsDefault,
		SortOrder: po.SortOrder,
		IsActive:  po.IsActive,
		CreatedAt: unixToTime(po.CreatedAt),
		UpdatedAt: unixToTime(po.UpdatedAt),
	}
}

// localeResourceRepo 本地化资源仓储实现
type localeResourceRepo struct {
	db *gorm.DB
}

// NewLocaleResourceRepo 创建本地化资源仓储
func NewLocaleResourceRepo(db *gorm.DB) biz.LocaleResourceRepo {
	return &localeResourceRepo{db: db}
}

func (r *localeResourceRepo) Create(ctx context.Context, resource *biz.LocaleResource) error {
	po := r.toPO(resource)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *localeResourceRepo) GetByID(ctx context.Context, id uint64) (*biz.LocaleResource, error) {
	var po LocaleResourcePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *localeResourceRepo) ListByLanguageID(ctx context.Context, languageID uint64, offset, limit int) ([]*biz.LocaleResource, error) {
	var pos []*LocaleResourcePO
	if err := r.db.WithContext(ctx).Where("language_id = ?", languageID).Offset(offset).Limit(limit).Order("module ASC, key ASC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.LocaleResource, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *localeResourceRepo) GetByKey(ctx context.Context, languageID uint64, key string) (*biz.LocaleResource, error) {
	var po LocaleResourcePO
	if err := r.db.WithContext(ctx).Where("language_id = ? AND key = ?", languageID, key).First(&po).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *localeResourceRepo) Update(ctx context.Context, resource *biz.LocaleResource) error {
	po := r.toPO(resource)
	return r.db.WithContext(ctx).Model(&LocaleResourcePO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *localeResourceRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&LocaleResourcePO{}, id).Error
}

func (r *localeResourceRepo) toPO(resource *biz.LocaleResource) *LocaleResourcePO {
	return &LocaleResourcePO{
		ID:         resource.ID,
		LanguageID: resource.LanguageID,
		Key:        resource.Key,
		Value:      resource.Value,
		Module:     resource.Module,
		CreatedAt:  resource.CreatedAt.Unix(),
		UpdatedAt:  resource.UpdatedAt.Unix(),
	}
}

func (r *localeResourceRepo) toEntity(po *LocaleResourcePO) *biz.LocaleResource {
	return &biz.LocaleResource{
		ID:         po.ID,
		LanguageID: po.LanguageID,
		Key:        po.Key,
		Value:      po.Value,
		Module:     po.Module,
		CreatedAt:  unixToTime(po.CreatedAt),
		UpdatedAt:  unixToTime(po.UpdatedAt),
	}
}
