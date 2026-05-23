// Package data 门店模块数据层 —— 门店的持久化对象与仓储实现
package data

import (
	"context"

	"nop-go/services/admin-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== 持久化对象（PO） ====================

// StorePO 门店持久化对象 —— 映射数据库 stores 表
type StorePO struct {
	ID        uint    `gorm:"primaryKey" db:"id"`
	Name      string  `gorm:"column:name;type:varchar(100);not null" db:"name"`             // 门店名称
	Code      string  `gorm:"column:code;type:varchar(100);uniqueIndex;not null" db:"code"` // 门店编码
	Address   string  `gorm:"column:address;type:varchar(500)" db:"address"`                   // 地址
	Phone     string  `gorm:"column:phone;type:varchar(20)" db:"phone"`                      // 联系电话
	Manager   string  `gorm:"column:manager;type:varchar(50)" db:"manager"`                    // 店长
	Region    string  `gorm:"column:region;type:varchar(50);index" db:"region"`               // 区域
	Business  string  `gorm:"column:business;type:varchar(100)" db:"business"`                  // 营业时间
	Status    int     `gorm:"column:status;type:tinyint;default:1;index" db:"status"`         // 状态
	Lng       float64 `gorm:"column:lng;type:decimal(10,6)" db:"lng"`                      // 经度
	Lat       float64 `gorm:"column:lat;type:decimal(10,6)" db:"lat"`                      // 纬度
	CreatedAt string  `gorm:"column:created_at" db:"created_at"`
	UpdatedAt string  `gorm:"column:updated_at" db:"updated_at"`
}

// TableName 指定门店表名
func (StorePO) TableName() string { return "stores" }

// ==================== PO ↔ Entity 转换 ====================

// toEntity 将 StorePO 转换为 biz.Store 领域实体
func (s *StorePO) toEntity() *biz.Store {
	return &biz.Store{
		ID:        s.ID,
		Name:      s.Name,
		Code:      s.Code,
		Address:   s.Address,
		Phone:     s.Phone,
		Manager:   s.Manager,
		Region:    s.Region,
		Business:  s.Business,
		Status:    s.Status,
		Lng:       s.Lng,
		Lat:       s.Lat,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// storeToPO 将 biz.Store 领域实体转换为 StorePO
func storeToPO(s *biz.Store) *StorePO {
	return &StorePO{
		ID:        s.ID,
		Name:      s.Name,
		Code:      s.Code,
		Address:   s.Address,
		Phone:     s.Phone,
		Manager:   s.Manager,
		Region:    s.Region,
		Business:  s.Business,
		Status:    s.Status,
		Lng:       s.Lng,
		Lat:       s.Lat,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// storeRepo 门店仓储实现
type storeRepo struct {
	db *gorm.DB
}

// NewStoreRepo 创建门店仓储
func NewStoreRepo(db *gorm.DB) biz.StoreRepo {
	return &storeRepo{db: db}
}

func (r *storeRepo) Create(ctx context.Context, s *biz.Store) error {
	po := storeToPO(s)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *storeRepo) GetByID(ctx context.Context, id uint) (*biz.Store, error) {
	var po StorePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *storeRepo) GetByCode(ctx context.Context, code string) (*biz.Store, error) {
	var po StorePO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *storeRepo) List(ctx context.Context, status int, region string, page, pageSize int) ([]*biz.Store, int64, error) {
	var pos []*StorePO
	var total int64
	q := r.db.WithContext(ctx).Model(&StorePO{})
	// 状态过滤：status < 0 表示不过滤
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	// 区域过滤
	if region != "" {
		q = q.Where("region = ?", region)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.Store, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *storeRepo) Update(ctx context.Context, s *biz.Store) error {
	po := storeToPO(s)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *storeRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&StorePO{}, id).Error
}
