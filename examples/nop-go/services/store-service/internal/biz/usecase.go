// Package biz 店铺服务业务逻辑层
package biz

import (
	"context"
	"errors"

	"nop-go/services/store-service/internal/data"
	"nop-go/services/store-service/internal/models"
)

// StoreUseCase 店铺用例
type StoreUseCase struct {
	storeRepo   data.StoreRepository
	vendorRepo  data.VendorRepository
	svRepo      data.StoreVendorRepository
}

// NewStoreUseCase 创建店铺用例
func NewStoreUseCase(storeRepo data.StoreRepository, vendorRepo data.VendorRepository, svRepo data.StoreVendorRepository) *StoreUseCase {
	return &StoreUseCase{storeRepo: storeRepo, vendorRepo: vendorRepo, svRepo: svRepo}
}

// CreateStore 创建店铺
func (uc *StoreUseCase) CreateStore(ctx context.Context, req *models.StoreCreateRequest) (*models.Store, error) {
	// 检查名称是否已存在
	if existing, _ := uc.storeRepo.GetByName(ctx, req.Name); existing != nil {
		return nil, data.ErrStoreNameExists
	}

	store := &models.Store{
		Name:             req.Name,
		URL:              req.URL,
		SSL:              req.SSL,
		Hosts:            req.Hosts,
		DefaultLanguageID: req.DefaultLanguageID,
		DisplayOrder:     req.DisplayOrder,
		CompanyName:      req.CompanyName,
		CompanyAddress:   req.CompanyAddress,
		CompanyPhone:     req.CompanyPhone,
		CompanyEmail:     req.CompanyEmail,
		Active:           req.Active,
	}

	if err := uc.storeRepo.Create(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// GetStore 获取店铺
func (uc *StoreUseCase) GetStore(ctx context.Context, id uint) (*models.Store, error) {
	store, err := uc.storeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrStoreNotFound
	}
	return store, nil
}

// ListStores 店铺列表
func (uc *StoreUseCase) ListStores(ctx context.Context, page, pageSize int) ([]*models.Store, int64, error) {
	return uc.storeRepo.List(ctx, page, pageSize)
}

// UpdateStore 更新店铺
func (uc *StoreUseCase) UpdateStore(ctx context.Context, id uint, req *models.StoreUpdateRequest) (*models.Store, error) {
	store, err := uc.storeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrStoreNotFound
	}

	if req.Name != "" {
		store.Name = req.Name
	}
	if req.URL != "" {
		store.URL = req.URL
	}
	if req.SSL != nil {
		store.SSL = *req.SSL
	}
	if req.Hosts != "" {
		store.Hosts = req.Hosts
	}
	if req.DefaultLanguageID != 0 {
		store.DefaultLanguageID = req.DefaultLanguageID
	}
	if req.DisplayOrder != nil {
		store.DisplayOrder = *req.DisplayOrder
	}
	if req.CompanyName != "" {
		store.CompanyName = req.CompanyName
	}
	if req.CompanyAddress != "" {
		store.CompanyAddress = req.CompanyAddress
	}
	if req.CompanyPhone != "" {
		store.CompanyPhone = req.CompanyPhone
	}
	if req.CompanyEmail != "" {
		store.CompanyEmail = req.CompanyEmail
	}
	if req.Active != nil {
		store.Active = *req.Active
	}

	if err := uc.storeRepo.Update(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// DeleteStore 删除店铺
func (uc *StoreUseCase) DeleteStore(ctx context.Context, id uint) error {
	_, err := uc.storeRepo.GetByID(ctx, id)
	if err != nil {
		return data.ErrStoreNotFound
	}
	return uc.storeRepo.Delete(ctx, id)
}

// GetAllActiveStores 获取所有活跃店铺
func (uc *StoreUseCase) GetAllActiveStores(ctx context.Context) ([]*models.Store, error) {
	return uc.storeRepo.GetAllActive(ctx)
}

// GetStoreVendors 获取店铺的供应商
func (uc *StoreUseCase) GetStoreVendors(ctx context.Context, storeID uint) ([]*models.Vendor, error) {
	return uc.vendorRepo.GetVendorsByStoreID(ctx, storeID)
}

// VendorUseCase 供应商用例
type VendorUseCase struct {
	vendorRepo data.VendorRepository
	storeRepo  data.StoreRepository
	svRepo     data.StoreVendorRepository
	noteRepo   data.VendorNoteRepository
}

// NewVendorUseCase 创建供应商用例
func NewVendorUseCase(vendorRepo data.VendorRepository, storeRepo data.StoreRepository, svRepo data.StoreVendorRepository, noteRepo data.VendorNoteRepository) *VendorUseCase {
	return &VendorUseCase{vendorRepo: vendorRepo, storeRepo: storeRepo, svRepo: svRepo, noteRepo: noteRepo}
}

// CreateVendor 创建供应商
func (uc *VendorUseCase) CreateVendor(ctx context.Context, req *models.VendorCreateRequest) (*models.Vendor, error) {
	// 检查名称是否已存在
	if existing, _ := uc.vendorRepo.GetByName(ctx, req.Name); existing != nil {
		return nil, errors.New("vendor name already exists")
	}

	vendor := &models.Vendor{
		Name:                            req.Name,
		Email:                           req.Email,
		Description:                     req.Description,
		AdminComment:                    req.AdminComment,
		AddressID:                       req.AddressID,
		Active:                          req.Active,
		DisplayOrder:                    req.DisplayOrder,
		AllowCustomersToContactVendors:  req.AllowCustomersToContactVendors,
	}

	if err := uc.vendorRepo.Create(ctx, vendor); err != nil {
		return nil, err
	}

	return vendor, nil
}

// GetVendor 获取供应商
func (uc *VendorUseCase) GetVendor(ctx context.Context, id uint) (*models.Vendor, error) {
	vendor, err := uc.vendorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrVendorNotFound
	}
	return vendor, nil
}

// ListVendors 供应商列表
func (uc *VendorUseCase) ListVendors(ctx context.Context, page, pageSize int) ([]*models.Vendor, int64, error) {
	return uc.vendorRepo.List(ctx, page, pageSize)
}

// UpdateVendor 更新供应商
func (uc *VendorUseCase) UpdateVendor(ctx context.Context, id uint, req *models.VendorUpdateRequest) (*models.Vendor, error) {
	vendor, err := uc.vendorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrVendorNotFound
	}

	if req.Name != "" {
		vendor.Name = req.Name
	}
	if req.Email != "" {
		vendor.Email = req.Email
	}
	if req.Description != "" {
		vendor.Description = req.Description
	}
	if req.AdminComment != "" {
		vendor.AdminComment = req.AdminComment
	}
	if req.AddressID != 0 {
		vendor.AddressID = req.AddressID
	}
	if req.Active != nil {
		vendor.Active = *req.Active
	}
	if req.DisplayOrder != nil {
		vendor.DisplayOrder = *req.DisplayOrder
	}
	if req.AllowCustomersToContactVendors != nil {
		vendor.AllowCustomersToContactVendors = *req.AllowCustomersToContactVendors
	}

	if err := uc.vendorRepo.Update(ctx, vendor); err != nil {
		return nil, err
	}

	return vendor, nil
}

// DeleteVendor 删除供应商
func (uc *VendorUseCase) DeleteVendor(ctx context.Context, id uint) error {
	_, err := uc.vendorRepo.GetByID(ctx, id)
	if err != nil {
		return data.ErrVendorNotFound
	}
	return uc.vendorRepo.Delete(ctx, id)
}

// AddVendorToStore 将供应商添加到店铺
func (uc *VendorUseCase) AddVendorToStore(ctx context.Context, storeID, vendorID uint, isDefault bool) error {
	// 验证店铺和供应商存在
	if _, err := uc.storeRepo.GetByID(ctx, storeID); err != nil {
		return data.ErrStoreNotFound
	}
	if _, err := uc.vendorRepo.GetByID(ctx, vendorID); err != nil {
		return data.ErrVendorNotFound
	}

	sv := &models.StoreVendor{
		StoreID:   storeID,
		VendorID:  vendorID,
		IsDefault: isDefault,
	}
	return uc.svRepo.Create(ctx, sv)
}

// RemoveVendorFromStore 从店铺移除供应商
func (uc *VendorUseCase) RemoveVendorFromStore(ctx context.Context, storeID, vendorID uint) error {
	return uc.svRepo.Delete(ctx, storeID, vendorID)
}

// AddVendorNote 添加供应商备注
func (uc *VendorUseCase) AddVendorNote(ctx context.Context, vendorID uint, note string) error {
	_, err := uc.vendorRepo.GetByID(ctx, vendorID)
	if err != nil {
		return data.ErrVendorNotFound
	}
	return uc.noteRepo.Create(ctx, &models.VendorNote{VendorID: vendorID, Note: note})
}

// GetVendorNotes 获取供应商备注
func (uc *VendorUseCase) GetVendorNotes(ctx context.Context, vendorID uint) ([]*models.VendorNote, error) {
	return uc.noteRepo.GetByVendorID(ctx, vendorID)
}