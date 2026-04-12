// Package biz 导入导出服务业务逻辑层
package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nop-go/services/import-service/internal/data"
	"nop-go/services/import-service/internal/models"
)

// ImportConfig 导入配置
type ImportConfig struct {
	TempDir                 string
	MaxFileSize             int64
	SupportedFormats        []string
	DefaultSeparator        string
	SkipAttributeValidation bool
}

// ImportUseCase 导入用例
type ImportUseCase struct {
	profileRepo data.ImportProfileRepository
	historyRepo data.ImportHistoryRepository
	errorRepo   data.ImportErrorRepository
	config      ImportConfig
}

// NewImportUseCase 创建导入用例
func NewImportUseCase(
	profileRepo data.ImportProfileRepository,
	historyRepo data.ImportHistoryRepository,
	errorRepo data.ImportErrorRepository,
	config ImportConfig,
) *ImportUseCase {
	return &ImportUseCase{
		profileRepo: profileRepo,
		historyRepo: historyRepo,
		errorRepo:   errorRepo,
		config:      config,
	}
}

// CreateProfile 创建导入配置
func (uc *ImportUseCase) CreateProfile(ctx context.Context, req *models.ImportProfileCreateRequest) (*models.ImportProfile, error) {
	// 序列化列映射
	var columnMappingJSON string
	if req.ColumnMapping != nil {
		bytes, err := json.Marshal(req.ColumnMapping)
		if err != nil {
			return nil, err
		}
		columnMappingJSON = string(bytes)
	}

	separator := req.Separator
	if separator == "" {
		separator = uc.config.DefaultSeparator
	}

	profile := &models.ImportProfile{
		Name:                   req.Name,
		EntityType:             req.EntityType,
		FilePath:               req.FilePath,
		FileType:               req.FileType,
		Separator:              separator,
		SkipAttributeValidation: req.SkipAttributeValidation,
		ColumnMapping:          columnMappingJSON,
		IsActive:               true,
	}

	if err := uc.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfile 获取导入配置
func (uc *ImportUseCase) GetProfile(ctx context.Context, id uint) (*models.ImportProfile, error) {
	return uc.profileRepo.GetByID(ctx, id)
}

// ListProfiles 导入配置列表
func (uc *ImportUseCase) ListProfiles(ctx context.Context, page, pageSize int) ([]*models.ImportProfile, int64, error) {
	return uc.profileRepo.List(ctx, page, pageSize)
}

// UpdateProfile 更新导入配置
func (uc *ImportUseCase) UpdateProfile(ctx context.Context, id uint, req *models.ImportProfileUpdateRequest) (*models.ImportProfile, error) {
	profile, err := uc.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrProfileNotFound
	}

	if req.Name != "" {
		profile.Name = req.Name
	}
	if req.FilePath != "" {
		profile.FilePath = req.FilePath
	}
	if req.FileType != "" {
		profile.FileType = req.FileType
	}
	if req.Separator != "" {
		profile.Separator = req.Separator
	}
	profile.SkipAttributeValidation = req.SkipAttributeValidation
	if req.ColumnMapping != nil {
		bytes, err := json.Marshal(req.ColumnMapping)
		if err != nil {
			return nil, err
		}
		profile.ColumnMapping = string(bytes)
	}
	profile.IsActive = req.IsActive

	if err := uc.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// DeleteProfile 删除导入配置
func (uc *ImportUseCase) DeleteProfile(ctx context.Context, id uint) error {
	return uc.profileRepo.Delete(ctx, id)
}

// ExecuteImport 执行导入
func (uc *ImportUseCase) ExecuteImport(ctx context.Context, req *models.ImportExecuteRequest) (*models.ImportResult, error) {
	// 获取配置
	profile, err := uc.profileRepo.GetByID(ctx, req.ProfileID)
	if err != nil {
		return nil, data.ErrProfileNotFound
	}

	// 创建历史记录
	history := &models.ImportHistory{
		ProfileID:    req.ProfileID,
		FileName:     req.FileName,
		Status:       "processing",
		TotalRecords: 0,
		SuccessCount: 0,
		ErrorCount:   0,
		StartedAt:    time.Now(),
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		return nil, err
	}

	// TODO: 实际导入逻辑
	// 这里需要根据 EntityType 调用相应的服务进行导入
	// 模拟导入过程
	result := &models.ImportResult{
		HistoryID:    history.ID,
		TotalRecords: 0,
		SuccessCount: 0,
		ErrorCount:   0,
		Status:       "completed",
		Message:      fmt.Sprintf("导入配置 %s 执行成功", profile.Name),
	}

	// 更新历史记录
	history.Status = result.Status
	history.TotalRecords = result.TotalRecords
	history.SuccessCount = result.SuccessCount
	history.ErrorCount = result.ErrorCount
	history.CompletedAt = time.Now()
	uc.historyRepo.Update(ctx, history)

	return result, nil
}

// GetImportHistory 获取导入历史
func (uc *ImportUseCase) GetImportHistory(ctx context.Context, id uint64) (*models.ImportHistory, error) {
	return uc.historyRepo.GetByID(ctx, id)
}

// ListImportHistory 导入历史列表
func (uc *ImportUseCase) ListImportHistory(ctx context.Context, page, pageSize int) ([]*models.ImportHistory, int64, error) {
	return uc.historyRepo.List(ctx, page, pageSize)
}

// ListImportHistoryByProfile 按配置获取导入历史
func (uc *ImportUseCase) ListImportHistoryByProfile(ctx context.Context, profileID uint, page, pageSize int) ([]*models.ImportHistory, int64, error) {
	return uc.historyRepo.ListByProfile(ctx, profileID, page, pageSize)
}

// GetImportErrors 获取导入错误
func (uc *ImportUseCase) GetImportErrors(ctx context.Context, historyID uint64) ([]*models.ImportError, error) {
	return uc.errorRepo.GetByHistoryID(ctx, historyID)
}

// ExportUseCase 导出用例
type ExportUseCase struct {
	profileRepo data.ExportProfileRepository
	historyRepo data.ExportHistoryRepository
	config      ImportConfig
}

// NewExportUseCase 创建导出用例
func NewExportUseCase(
	profileRepo data.ExportProfileRepository,
	historyRepo data.ExportHistoryRepository,
	config ImportConfig,
) *ExportUseCase {
	return &ExportUseCase{
		profileRepo: profileRepo,
		historyRepo: historyRepo,
		config:      config,
	}
}

// CreateProfile 创建导出配置
func (uc *ExportUseCase) CreateProfile(ctx context.Context, req *models.ExportProfileCreateRequest) (*models.ExportProfile, error) {
	var columnSelectionJSON, filterCriteriaJSON string
	if req.ColumnSelection != nil {
		bytes, err := json.Marshal(req.ColumnSelection)
		if err != nil {
			return nil, err
		}
		columnSelectionJSON = string(bytes)
	}
	if req.FilterCriteria != nil {
		bytes, err := json.Marshal(req.FilterCriteria)
		if err != nil {
			return nil, err
		}
		filterCriteriaJSON = string(bytes)
	}

	profile := &models.ExportProfile{
		Name:            req.Name,
		EntityType:      req.EntityType,
		FilePath:        req.FilePath,
		FileType:        req.FileType,
		ExportWithIds:   req.ExportWithIds,
		ColumnSelection: columnSelectionJSON,
		FilterCriteria:  filterCriteriaJSON,
		IsActive:        true,
	}

	if err := uc.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfile 获取导出配置
func (uc *ExportUseCase) GetProfile(ctx context.Context, id uint) (*models.ExportProfile, error) {
	return uc.profileRepo.GetByID(ctx, id)
}

// ListProfiles 导出配置列表
func (uc *ExportUseCase) ListProfiles(ctx context.Context, page, pageSize int) ([]*models.ExportProfile, int64, error) {
	return uc.profileRepo.List(ctx, page, pageSize)
}

// UpdateProfile 更新导出配置
func (uc *ExportUseCase) UpdateProfile(ctx context.Context, id uint, req *models.ExportProfileUpdateRequest) (*models.ExportProfile, error) {
	profile, err := uc.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrProfileNotFound
	}

	if req.Name != "" {
		profile.Name = req.Name
	}
	if req.FilePath != "" {
		profile.FilePath = req.FilePath
	}
	if req.FileType != "" {
		profile.FileType = req.FileType
	}
	profile.ExportWithIds = req.ExportWithIds
	if req.ColumnSelection != nil {
		bytes, err := json.Marshal(req.ColumnSelection)
		if err != nil {
			return nil, err
		}
		profile.ColumnSelection = string(bytes)
	}
	if req.FilterCriteria != nil {
		bytes, err := json.Marshal(req.FilterCriteria)
		if err != nil {
			return nil, err
		}
		profile.FilterCriteria = string(bytes)
	}
	profile.IsActive = req.IsActive

	if err := uc.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// DeleteProfile 删除导出配置
func (uc *ExportUseCase) DeleteProfile(ctx context.Context, id uint) error {
	return uc.profileRepo.Delete(ctx, id)
}

// ExecuteExport 执行导出
func (uc *ExportUseCase) ExecuteExport(ctx context.Context, req *models.ExportExecuteRequest) (*models.ExportResult, error) {
	// 获取配置
	profile, err := uc.profileRepo.GetByID(ctx, req.ProfileID)
	if err != nil {
		return nil, data.ErrProfileNotFound
	}

	// 创建历史记录
	history := &models.ExportHistory{
		ProfileID: req.ProfileID,
		FileName:  fmt.Sprintf("export_%s_%d.%s", profile.EntityType, time.Now().Unix(), profile.FileType),
		Status:    "processing",
		StartedAt: time.Now(),
	}
	if err := uc.historyRepo.Create(ctx, history); err != nil {
		return nil, err
	}

	// TODO: 实际导出逻辑
	// 这里需要根据 EntityType 调用相应的服务进行导出
	result := &models.ExportResult{
		HistoryID:   history.ID,
		RecordCount: 0,
		FileSize:    0,
		FilePath:    "",
		Status:      "completed",
	}

	// 更新历史记录
	history.Status = result.Status
	history.RecordCount = result.RecordCount
	history.FileSize = result.FileSize
	history.FilePath = result.FilePath
	history.CompletedAt = time.Now()
	uc.historyRepo.Update(ctx, history)

	return result, nil
}

// GetExportHistory 获取导出历史
func (uc *ExportUseCase) GetExportHistory(ctx context.Context, id uint64) (*models.ExportHistory, error) {
	return uc.historyRepo.GetByID(ctx, id)
}

// ListExportHistory 导出历史列表
func (uc *ExportUseCase) ListExportHistory(ctx context.Context, page, pageSize int) ([]*models.ExportHistory, int64, error) {
	return uc.historyRepo.List(ctx, page, pageSize)
}

// ListExportHistoryByProfile 按配置获取导出历史
func (uc *ExportUseCase) ListExportHistoryByProfile(ctx context.Context, profileID uint, page, pageSize int) ([]*models.ExportHistory, int64, error) {
	return uc.historyRepo.ListByProfile(ctx, profileID, page, pageSize)
}

// GetEntityTypes 获取支持的实体类型
func GetEntityTypes() []models.EntityType {
	return models.EntityTypes
}