// Package service 服务层 —— 聚合6个模块的业务服务，负责 UseCase 编排与 DTO 转换
//
// Services 是 admin-service 的统一入口，持有6个子模块的 Service 实例。
// 每个 Service 负责调用对应的 UseCase，并将领域实体转换为响应 DTO。
// 优惠模块(DiscountService)的实体→响应转换逻辑在此层实现，biz 层不依赖 request/response。
package service

import (
	"context"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/data"

	"gorm.io/gorm"
)

// ==================== 安全模块 Service ====================

// SecurityService 安全模块服务 —— 封装安全相关的业务调用
type SecurityService struct {
	uc *biz.SecurityUseCase
}

// NewSecurityService 创建安全模块服务
func NewSecurityService(uc *biz.SecurityUseCase) *SecurityService {
	return &SecurityService{uc: uc}
}

// CreatePermission 创建权限
func (s *SecurityService) CreatePermission(ctx context.Context, p *biz.Permission) error {
	return s.uc.CreatePermission(ctx, p)
}

// GetPermission 获取权限
func (s *SecurityService) GetPermission(ctx context.Context, id uint) (*biz.Permission, error) {
	return s.uc.GetPermission(ctx, id)
}

// ListPermissions 获取权限列表
func (s *SecurityService) ListPermissions(ctx context.Context, module string, page, pageSize int) ([]*biz.Permission, int64, error) {
	return s.uc.ListPermissions(ctx, module, page, pageSize)
}

// UpdatePermission 更新权限
func (s *SecurityService) UpdatePermission(ctx context.Context, p *biz.Permission) error {
	return s.uc.UpdatePermission(ctx, p)
}

// DeletePermission 删除权限
func (s *SecurityService) DeletePermission(ctx context.Context, id uint) error {
	return s.uc.DeletePermission(ctx, id)
}

// CreateACL 创建ACL规则
func (s *SecurityService) CreateACL(ctx context.Context, acl *biz.ACL) error {
	return s.uc.CreateACL(ctx, acl)
}

// GetACL 获取ACL规则
func (s *SecurityService) GetACL(ctx context.Context, id uint) (*biz.ACL, error) {
	return s.uc.GetACL(ctx, id)
}

// ListACLs 获取ACL规则列表
func (s *SecurityService) ListACLs(ctx context.Context, roleID uint, page, pageSize int) ([]*biz.ACL, int64, error) {
	return s.uc.ListACLs(ctx, roleID, page, pageSize)
}

// UpdateACL 更新ACL规则
func (s *SecurityService) UpdateACL(ctx context.Context, acl *biz.ACL) error {
	return s.uc.UpdateACL(ctx, acl)
}

// DeleteACL 删除ACL规则
func (s *SecurityService) DeleteACL(ctx context.Context, id uint) error {
	return s.uc.DeleteACL(ctx, id)
}

// GetACLsByRoleID 根据角色ID获取ACL规则
func (s *SecurityService) GetACLsByRoleID(ctx context.Context, roleID uint) ([]*biz.ACL, error) {
	return s.uc.GetACLsByRoleID(ctx, roleID)
}

// ==================== 插件模块 Service ====================

// PluginService 插件模块服务
type PluginService struct {
	uc *biz.PluginUseCase
}

// NewPluginService 创建插件模块服务
func NewPluginService(uc *biz.PluginUseCase) *PluginService {
	return &PluginService{uc: uc}
}

// CreatePlugin 创建插件
func (s *PluginService) CreatePlugin(ctx context.Context, p *biz.Plugin) error {
	return s.uc.CreatePlugin(ctx, p)
}

// GetPlugin 获取插件
func (s *PluginService) GetPlugin(ctx context.Context, id uint) (*biz.Plugin, error) {
	return s.uc.GetPlugin(ctx, id)
}

// ListPlugins 获取插件列表
func (s *PluginService) ListPlugins(ctx context.Context, status int, page, pageSize int) ([]*biz.Plugin, int64, error) {
	return s.uc.ListPlugins(ctx, status, page, pageSize)
}

// UpdatePlugin 更新插件
func (s *PluginService) UpdatePlugin(ctx context.Context, p *biz.Plugin) error {
	return s.uc.UpdatePlugin(ctx, p)
}

// DeletePlugin 删除插件
func (s *PluginService) DeletePlugin(ctx context.Context, id uint) error {
	return s.uc.DeletePlugin(ctx, id)
}

// ==================== 门店模块 Service ====================

// StoreService 门店模块服务
type StoreService struct {
	uc *biz.StoreUseCase
}

// NewStoreService 创建门店模块服务
func NewStoreService(uc *biz.StoreUseCase) *StoreService {
	return &StoreService{uc: uc}
}

// CreateStore 创建门店
func (s *StoreService) CreateStore(ctx context.Context, st *biz.Store) error {
	return s.uc.CreateStore(ctx, st)
}

// GetStore 获取门店
func (s *StoreService) GetStore(ctx context.Context, id uint) (*biz.Store, error) {
	return s.uc.GetStore(ctx, id)
}

// ListStores 获取门店列表
func (s *StoreService) ListStores(ctx context.Context, status int, region string, page, pageSize int) ([]*biz.Store, int64, error) {
	return s.uc.ListStores(ctx, status, region, page, pageSize)
}

// UpdateStore 更新门店
func (s *StoreService) UpdateStore(ctx context.Context, st *biz.Store) error {
	return s.uc.UpdateStore(ctx, st)
}

// DeleteStore 删除门店
func (s *StoreService) DeleteStore(ctx context.Context, id uint) error {
	return s.uc.DeleteStore(ctx, id)
}

// ==================== 日志模块 Service ====================

// LoggingService 日志模块服务
type LoggingService struct {
	uc *biz.LoggingUseCase
}

// NewLoggingService 创建日志模块服务
func NewLoggingService(uc *biz.LoggingUseCase) *LoggingService {
	return &LoggingService{uc: uc}
}

// CreateActivityLog 创建活动日志
func (s *LoggingService) CreateActivityLog(ctx context.Context, log *biz.ActivityLog) error {
	return s.uc.CreateActivityLog(ctx, log)
}

// GetActivityLog 获取活动日志
func (s *LoggingService) GetActivityLog(ctx context.Context, id uint) (*biz.ActivityLog, error) {
	return s.uc.GetActivityLog(ctx, id)
}

// ListActivityLogs 获取活动日志列表
func (s *LoggingService) ListActivityLogs(ctx context.Context, userID uint, action string, page, pageSize int) ([]*biz.ActivityLog, int64, error) {
	return s.uc.ListActivityLogs(ctx, userID, action, page, pageSize)
}

// CreateSystemLog 创建系统日志
func (s *LoggingService) CreateSystemLog(ctx context.Context, log *biz.SystemLog) error {
	return s.uc.CreateSystemLog(ctx, log)
}

// GetSystemLog 获取系统日志
func (s *LoggingService) GetSystemLog(ctx context.Context, id uint) (*biz.SystemLog, error) {
	return s.uc.GetSystemLog(ctx, id)
}

// ListSystemLogs 获取系统日志列表
func (s *LoggingService) ListSystemLogs(ctx context.Context, level, module string, page, pageSize int) ([]*biz.SystemLog, int64, error) {
	return s.uc.ListSystemLogs(ctx, level, module, page, pageSize)
}

// ==================== 优惠模块 Service ====================

// DiscountService 优惠模块服务
//
// 重要：该服务负责将 biz.Discount / biz.DiscountUsage 领域实体转换为响应 DTO。
// 这是从原 discount 独立服务重构而来——原 biz 层直接返回 response DTO，现已修正。
type DiscountService struct {
	uc *biz.DiscountUseCase
}

// NewDiscountService 创建优惠模块服务
func NewDiscountService(uc *biz.DiscountUseCase) *DiscountService {
	return &DiscountService{uc: uc}
}

// CreateDiscount 创建优惠 —— 将请求实体传入 UseCase
func (s *DiscountService) CreateDiscount(ctx context.Context, d *biz.Discount) error {
	return s.uc.CreateDiscount(ctx, d)
}

// GetDiscount 获取优惠 —— UseCase 返回领域实体，handler 负责转 DTO
func (s *DiscountService) GetDiscount(ctx context.Context, id uint) (*biz.Discount, error) {
	return s.uc.GetDiscount(ctx, id)
}

// ListDiscounts 获取优惠列表
func (s *DiscountService) ListDiscounts(ctx context.Context, discountType, status int, page, pageSize int) ([]*biz.Discount, int64, error) {
	return s.uc.ListDiscounts(ctx, discountType, status, page, pageSize)
}

// UpdateDiscount 更新优惠
func (s *DiscountService) UpdateDiscount(ctx context.Context, d *biz.Discount) error {
	return s.uc.UpdateDiscount(ctx, d)
}

// DeleteDiscount 删除优惠
func (s *DiscountService) DeleteDiscount(ctx context.Context, id uint) error {
	return s.uc.DeleteDiscount(ctx, id)
}

// CreateDiscountUsage 创建优惠使用记录
func (s *DiscountService) CreateDiscountUsage(ctx context.Context, u *biz.DiscountUsage) error {
	return s.uc.CreateDiscountUsage(ctx, u)
}

// GetDiscountUsage 获取优惠使用记录
func (s *DiscountService) GetDiscountUsage(ctx context.Context, id uint) (*biz.DiscountUsage, error) {
	return s.uc.GetDiscountUsage(ctx, id)
}

// ListDiscountUsages 获取优惠使用记录列表
func (s *DiscountService) ListDiscountUsages(ctx context.Context, discountID, userID uint, page, pageSize int) ([]*biz.DiscountUsage, int64, error) {
	return s.uc.ListDiscountUsages(ctx, discountID, userID, page, pageSize)
}

// GetDiscountByCode 根据优惠码获取优惠
func (s *DiscountService) GetDiscountByCode(ctx context.Context, code string) (*biz.Discount, error) {
	return s.uc.GetDiscountByCode(ctx, code)
}

// GetUserUsageCount 获取用户对某优惠的使用次数
func (s *DiscountService) GetUserUsageCount(ctx context.Context, discountID, userID uint) (int64, error) {
	return s.uc.GetUserUsageCount(ctx, discountID, userID)
}

// ==================== 供应商模块 Service ====================

// VendorService 供应商模块服务
type VendorService struct {
	uc *biz.VendorUseCase
}

// NewVendorService 创建供应商模块服务
func NewVendorService(uc *biz.VendorUseCase) *VendorService {
	return &VendorService{uc: uc}
}

// CreateVendor 创建供应商
func (s *VendorService) CreateVendor(ctx context.Context, v *biz.Vendor) error {
	return s.uc.CreateVendor(ctx, v)
}

// GetVendor 获取供应商
func (s *VendorService) GetVendor(ctx context.Context, id uint) (*biz.Vendor, error) {
	return s.uc.GetVendor(ctx, id)
}

// ListVendors 获取供应商列表
func (s *VendorService) ListVendors(ctx context.Context, category string, status int, page, pageSize int) ([]*biz.Vendor, int64, error) {
	return s.uc.ListVendors(ctx, category, status, page, pageSize)
}

// UpdateVendor 更新供应商
func (s *VendorService) UpdateVendor(ctx context.Context, v *biz.Vendor) error {
	return s.uc.UpdateVendor(ctx, v)
}

// DeleteVendor 删除供应商
func (s *VendorService) DeleteVendor(ctx context.Context, id uint) error {
	return s.uc.DeleteVendor(ctx, id)
}

// ==================== 统一服务聚合 ====================

// Services 管理后台统一服务集合 —— 聚合6个模块的 Service 实例
//
// 由 NewServices 根据 gorm.DB 创建所有依赖链：
//   DB → Repo → UseCase → Service
type Services struct {
	Security *SecurityService // 安全模块：权限 + ACL
	Plugin   *PluginService   // 插件模块
	Store    *StoreService    // 门店模块
	Logging  *LoggingService  // 日志模块：活动日志 + 系统日志
	Discount *DiscountService // 优惠模块：优惠 + 使用记录
	Vendor   *VendorService   // 供应商模块
}

// NewServices 创建管理后台统一服务集合
//
// 依次构造所有依赖链：DB → Repo → UseCase → Service
// 这是对 Wire 依赖注入的手动替代，确保所有模块的依赖关系正确建立
func NewServices(db *gorm.DB) *Services {
	// ---------- 安全模块 ----------
	permRepo := data.NewPermissionRepo(db)
	aclRepo := data.NewACLRepo(db)
	securityUC := biz.NewSecurityUseCase(permRepo, aclRepo)
	securitySvc := NewSecurityService(securityUC)

	// ---------- 插件模块 ----------
	pluginRepo := data.NewPluginRepo(db)
	pluginUC := biz.NewPluginUseCase(pluginRepo)
	pluginSvc := NewPluginService(pluginUC)

	// ---------- 门店模块 ----------
	storeRepo := data.NewStoreRepo(db)
	storeUC := biz.NewStoreUseCase(storeRepo)
	storeSvc := NewStoreService(storeUC)

	// ---------- 日志模块 ----------
	activityRepo := data.NewActivityLogRepo(db)
	systemRepo := data.NewSystemLogRepo(db)
	loggingUC := biz.NewLoggingUseCase(activityRepo, systemRepo)
	loggingSvc := NewLoggingService(loggingUC)

	// ---------- 优惠模块 ----------
	discountRepo := data.NewDiscountRepo(db)
	usageRepo := data.NewDiscountUsageRepo(db)
	discountUC := biz.NewDiscountUseCase(discountRepo, usageRepo)
	discountSvc := NewDiscountService(discountUC)

	// ---------- 供应商模块 ----------
	vendorRepo := data.NewVendorRepo(db)
	vendorUC := biz.NewVendorUseCase(vendorRepo)
	vendorSvc := NewVendorService(vendorUC)

	return &Services{
		Security: securitySvc,
		Plugin:   pluginSvc,
		Store:    storeSvc,
		Logging:  loggingSvc,
		Discount: discountSvc,
		Vendor:   vendorSvc,
	}
}
