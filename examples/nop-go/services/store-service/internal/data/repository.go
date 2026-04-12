// Package data 店铺服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/store-service/internal/models"

	"gorm.io/gorm"
)

// StoreRepository 店铺仓储接口
type StoreRepository interface {
	Create(ctx context.Context, store *models.Store) error
	GetByID(ctx context.Context, id uint) (*models.Store, error)
	GetByName(ctx context.Context, name string) (*models.Store, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Store, int64, error)
	Update(ctx context.Context, store *models.Store) error
	Delete(ctx context.Context, id uint) error
	GetAllActive(ctx context.Context) ([]*models.Store, error)
}

// storeRepository 店铺仓储实现
type storeRepository struct {
	db *gorm.DB
}

// NewStoreRepository 创建店铺仓储
func NewStoreRepository(db *gorm.DB) StoreRepository {
	return &storeRepository{db: db}
}

func (r *storeRepository) Create(ctx context.Context, store *models.Store) error {
	return r.db.WithContext(ctx).Create(store).Error
}

func (r *storeRepository) GetByID(ctx context.Context, id uint) (*models.Store, error) {
	var store models.Store
	err := r.db.WithContext(ctx).First(&store, id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) GetByName(ctx context.Context, name string) (*models.Store, error) {
	var store models.Store
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) List(ctx context.Context, page, pageSize int) ([]*models.Store, int64, error) {
	var stores []*models.Store
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Store{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("display_order, id").Offset(offset).Limit(pageSize).Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

func (r *storeRepository) Update(ctx context.Context, store *models.Store) error {
	return r.db.WithContext(ctx).Save(store).Error
}

func (r *storeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Store{}, id).Error
}

func (r *storeRepository) GetAllActive(ctx context.Context) ([]*models.Store, error) {
	var stores []*models.Store
	err := r.db.WithContext(ctx).Where("active = ?", true).Order("display_order, id").Find(&stores).Error
	return stores, err
}

// VendorRepository 供应商仓储接口
type VendorRepository interface {
	Create(ctx context.Context, vendor *models.Vendor) error
	GetByID(ctx context.Context, id uint) (*models.Vendor, error)
	GetByName(ctx context.Context, name string) (*models.Vendor, error)
	GetByEmail(ctx context.Context, email string) (*models.Vendor, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Vendor, int64, error)
	Update(ctx context.Context, vendor *models.Vendor) error
	Delete(ctx context.Context, id uint) error
	GetVendorsByStoreID(ctx context.Context, storeID uint) ([]*models.Vendor, error)
}

// vendorRepository 供应商仓储实现
type vendorRepository struct {
	db *gorm.DB
}

// NewVendorRepository 创建供应商仓储
func NewVendorRepository(db *gorm.DB) VendorRepository {
	return &vendorRepository{db: db}
}

func (r *vendorRepository) Create(ctx context.Context, vendor *models.Vendor) error {
	return r.db.WithContext(ctx).Create(vendor).Error
}

func (r *vendorRepository) GetByID(ctx context.Context, id uint) (*models.Vendor, error) {
	var vendor models.Vendor
	err := r.db.WithContext(ctx).First(&vendor, id).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (r *vendorRepository) GetByName(ctx context.Context, name string) (*models.Vendor, error) {
	var vendor models.Vendor
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (r *vendorRepository) GetByEmail(ctx context.Context, email string) (*models.Vendor, error) {
	var vendor models.Vendor
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (r *vendorRepository) List(ctx context.Context, page, pageSize int) ([]*models.Vendor, int64, error) {
	var vendors []*models.Vendor
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Vendor{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("display_order, id").Offset(offset).Limit(pageSize).Find(&vendors).Error; err != nil {
		return nil, 0, err
	}

	return vendors, total, nil
}

func (r *vendorRepository) Update(ctx context.Context, vendor *models.Vendor) error {
	return r.db.WithContext(ctx).Save(vendor).Error
}

func (r *vendorRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Vendor{}, id).Error
}

func (r *vendorRepository) GetVendorsByStoreID(ctx context.Context, storeID uint) ([]*models.Vendor, error) {
	var vendors []*models.Vendor
	err := r.db.WithContext(ctx).
		Joins("JOIN store_vendors ON store_vendors.vendor_id = vendors.id").
		Where("store_vendors.store_id = ?", storeID).
		Find(&vendors).Error
	return vendors, err
}

// StoreVendorRepository 店铺-供应商仓储接口
type StoreVendorRepository interface {
	Create(ctx context.Context, sv *models.StoreVendor) error
	GetByStoreAndVendor(ctx context.Context, storeID, vendorID uint) (*models.StoreVendor, error)
	Delete(ctx context.Context, storeID, vendorID uint) error
	GetByStoreID(ctx context.Context, storeID uint) ([]*models.StoreVendor, error)
}

// storeVendorRepository 店铺-供应商仓储实现
type storeVendorRepository struct {
	db *gorm.DB
}

// NewStoreVendorRepository 创建店铺-供应商仓储
func NewStoreVendorRepository(db *gorm.DB) StoreVendorRepository {
	return &storeVendorRepository{db: db}
}

func (r *storeVendorRepository) Create(ctx context.Context, sv *models.StoreVendor) error {
	return r.db.WithContext(ctx).Create(sv).Error
}

func (r *storeVendorRepository) GetByStoreAndVendor(ctx context.Context, storeID, vendorID uint) (*models.StoreVendor, error) {
	var sv models.StoreVendor
	err := r.db.WithContext(ctx).Where("store_id = ? AND vendor_id = ?", storeID, vendorID).First(&sv).Error
	if err != nil {
		return nil, err
	}
	return &sv, nil
}

func (r *storeVendorRepository) Delete(ctx context.Context, storeID, vendorID uint) error {
	return r.db.WithContext(ctx).Where("store_id = ? AND vendor_id = ?", storeID, vendorID).Delete(&models.StoreVendor{}).Error
}

func (r *storeVendorRepository) GetByStoreID(ctx context.Context, storeID uint) ([]*models.StoreVendor, error) {
	var svs []*models.StoreVendor
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Find(&svs).Error
	return svs, err
}

// VendorNoteRepository 供应商备注仓储
type VendorNoteRepository interface {
	Create(ctx context.Context, note *models.VendorNote) error
	GetByVendorID(ctx context.Context, vendorID uint) ([]*models.VendorNote, error)
	Delete(ctx context.Context, id uint) error
}

type vendorNoteRepository struct {
	db *gorm.DB
}

func NewVendorNoteRepository(db *gorm.DB) VendorNoteRepository {
	return &vendorNoteRepository{db: db}
}

func (r *vendorNoteRepository) Create(ctx context.Context, note *models.VendorNote) error {
	return r.db.WithContext(ctx).Create(note).Error
}

func (r *vendorNoteRepository) GetByVendorID(ctx context.Context, vendorID uint) ([]*models.VendorNote, error) {
	var notes []*models.VendorNote
	err := r.db.WithContext(ctx).Where("vendor_id = ?", vendorID).Order("created_at desc").Find(&notes).Error
	return notes, err
}

func (r *vendorNoteRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.VendorNote{}, id).Error
}

// 常见错误
var (
	ErrStoreNotFound   = errors.New("store not found")
	ErrStoreNameExists = errors.New("store name already exists")
	ErrVendorNotFound  = errors.New("vendor not found")
)