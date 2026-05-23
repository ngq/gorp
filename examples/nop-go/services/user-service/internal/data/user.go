// Package data 定义用户服务的数据访问层，包含持久化对象(PO)和仓储实现。
// 本文件合并了原 user 服务的数据层：用户、地址、外部关联、可下载产品。
package data

import (
	"context"
	"time"

	"nop-go/services/user-service/internal/biz"

	"gorm.io/gorm"
)

// ======================== 持久化对象(PO) ========================

// UserPO 用户持久化对象，映射数据库 users 表
type UserPO struct {
	ID        uint           `gorm:"primaryKey" db:"id" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64;not null" db:"username" json:"username"`
	Email     string         `gorm:"size:128" db:"email" json:"email"`
	Phone     string         `gorm:"size:20" db:"phone" json:"phone"`
	Password  string         `gorm:"size:256;not null" db:"password" json:"-"`
	Nickname  string         `gorm:"size:64" db:"nickname" json:"nickname"`
	Avatar    string         `gorm:"size:512" db:"avatar" json:"avatar"`
	Status    int            `gorm:"default:1" db:"status" json:"status"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" db:"deleted_at" json:"-"`
}

// TableName 指定用户表名
func (UserPO) TableName() string { return "users" }

// AddressPO 用户地址持久化对象，映射数据库 addresses 表
type AddressPO struct {
	ID            uint           `gorm:"primaryKey" db:"id" json:"id"`
	UserID        uint           `gorm:"index;not null" db:"user_id" json:"user_id"`
	RecipientName string         `gorm:"size:64;not null" db:"recipient_name" json:"recipient_name"`
	Phone         string         `gorm:"size:20;not null" db:"phone" json:"phone"`
	Province      string         `gorm:"size:32;not null" db:"province" json:"province"`
	City          string         `gorm:"size:32;not null" db:"city" json:"city"`
	District      string         `gorm:"size:32;not null" db:"district" json:"district"`
	Detail        string         `gorm:"size:256;not null" db:"detail" json:"detail"`
	IsDefault     bool           `gorm:"default:false" db:"is_default" json:"is_default"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" db:"deleted_at" json:"-"`
}

// TableName 指定地址表名
func (AddressPO) TableName() string { return "addresses" }

// ExternalAssociationPO 外部关联持久化对象，映射数据库 external_associations 表
type ExternalAssociationPO struct {
	ID           uint           `gorm:"primaryKey" db:"id" json:"id"`
	UserID       uint           `gorm:"index;not null" db:"user_id" json:"user_id"`
	Platform     string         `gorm:"size:32;not null" db:"platform" json:"platform"`
	ExternalID   string         `gorm:"size:128;not null" db:"external_id" json:"external_id"`
	ExternalData string         `gorm:"size:1024" db:"external_data" json:"external_data"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" db:"deleted_at" json:"-"`
}

// TableName 指定外部关联表名
func (ExternalAssociationPO) TableName() string { return "external_associations" }

// DownloadableProductPO 可下载产品持久化对象，映射数据库 downloadable_products 表
type DownloadableProductPO struct {
	ID          uint           `gorm:"primaryKey" db:"id" json:"id"`
	UserID      uint           `gorm:"index;not null" db:"user_id" json:"user_id"`
	ProductID   string         `gorm:"size:64;not null" db:"product_id" json:"product_id"`
	ProductName string         `gorm:"size:128;not null" db:"product_name" json:"product_name"`
	DownloadURL string         `gorm:"size:512" db:"download_url" json:"download_url"`
	ExpireAt    *time.Time     `db:"expire_at" json:"expire_at"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" db:"deleted_at" json:"-"`
}

// TableName 指定可下载产品表名
func (DownloadableProductPO) TableName() string { return "downloadable_products" }

// ======================== PO <-> 领域模型转换 ========================

// toUser 将 UserPO 转换为用户领域模型
func (p *UserPO) toUser() *biz.User {
	return &biz.User{
		ID:        p.ID,
		Username:  p.Username,
		Email:     p.Email,
		Phone:     p.Phone,
		Password:  p.Password,
		Nickname:  p.Nickname,
		Avatar:    p.Avatar,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// toUserPO 将用户领域模型转换为 UserPO
func toUserPO(user *biz.User) *UserPO {
	return &UserPO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Password: user.Password,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   user.Status,
	}
}

// toAddress 将 AddressPO 转换为地址领域模型
func (p *AddressPO) toAddress() *biz.Address {
	return &biz.Address{
		ID:            p.ID,
		UserID:        p.UserID,
		RecipientName: p.RecipientName,
		Phone:         p.Phone,
		Province:      p.Province,
		City:          p.City,
		District:      p.District,
		Detail:        p.Detail,
		IsDefault:     p.IsDefault,
	}
}

// toAddressPO 将地址领域模型转换为 AddressPO
func toAddressPO(addr *biz.Address) *AddressPO {
	return &AddressPO{
		ID:            addr.ID,
		UserID:        addr.UserID,
		RecipientName: addr.RecipientName,
		Phone:         addr.Phone,
		Province:      addr.Province,
		City:          addr.City,
		District:      addr.District,
		Detail:        addr.Detail,
		IsDefault:     addr.IsDefault,
	}
}

// toExternalAssociation 将 ExternalAssociationPO 转换为外部关联领域模型
func (p *ExternalAssociationPO) toExternalAssociation() *biz.ExternalAssociation {
	return &biz.ExternalAssociation{
		ID:           p.ID,
		UserID:       p.UserID,
		Platform:     p.Platform,
		ExternalID:   p.ExternalID,
		ExternalData: p.ExternalData,
	}
}

// toExternalAssociationPO 将外部关联领域模型转换为 ExternalAssociationPO
func toExternalAssociationPO(ea *biz.ExternalAssociation) *ExternalAssociationPO {
	return &ExternalAssociationPO{
		ID:           ea.ID,
		UserID:       ea.UserID,
		Platform:     ea.Platform,
		ExternalID:   ea.ExternalID,
		ExternalData: ea.ExternalData,
	}
}

// toDownloadableProduct 将 DownloadableProductPO 转换为可下载产品领域模型
func (p *DownloadableProductPO) toDownloadableProduct() *biz.DownloadableProduct {
	return &biz.DownloadableProduct{
		ID:          p.ID,
		UserID:      p.UserID,
		ProductID:   p.ProductID,
		ProductName: p.ProductName,
		DownloadURL: p.DownloadURL,
		ExpireAt:    p.ExpireAt,
	}
}

// toDownloadableProductPO 将可下载产品领域模型转换为 DownloadableProductPO
func toDownloadableProductPO(dp *biz.DownloadableProduct) *DownloadableProductPO {
	return &DownloadableProductPO{
		ID:          dp.ID,
		UserID:      dp.UserID,
		ProductID:   dp.ProductID,
		ProductName: dp.ProductName,
		DownloadURL: dp.DownloadURL,
		ExpireAt:    dp.ExpireAt,
	}
}

// ======================== 仓储实现 ========================

// userRepo 用户仓储实现
type userRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储实例
func NewUserRepo(db *gorm.DB) biz.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	po := toUserPO(user)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.toUser(), nil
}

func (r *userRepo) Update(ctx context.Context, user *biz.User) (*biz.User, error) {
	po := toUserPO(user)
	if err := r.db.WithContext(ctx).Model(po).Updates(po).Error; err != nil {
		return nil, err
	}
	return po.toUser(), nil
}

func (r *userRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&UserPO{}, id).Error
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*biz.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toUser(), nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.toUser(), nil
}

func (r *userRepo) List(ctx context.Context, offset, limit int) ([]*biz.User, int64, error) {
	var pos []*UserPO
	var total int64
	// 先统计总数
	if err := r.db.WithContext(ctx).Model(&UserPO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 再分页查询
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	users := make([]*biz.User, 0, len(pos))
	for _, po := range pos {
		users = append(users, po.toUser())
	}
	return users, total, nil
}

// addressRepo 地址仓储实现
type addressRepo struct {
	db *gorm.DB
}

// NewAddressRepo 创建地址仓储实例
func NewAddressRepo(db *gorm.DB) biz.AddressRepository {
	return &addressRepo{db: db}
}

func (r *addressRepo) Create(ctx context.Context, address *biz.Address) (*biz.Address, error) {
	po := toAddressPO(address)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.toAddress(), nil
}

func (r *addressRepo) Update(ctx context.Context, address *biz.Address) (*biz.Address, error) {
	po := toAddressPO(address)
	if err := r.db.WithContext(ctx).Model(po).Updates(po).Error; err != nil {
		return nil, err
	}
	return po.toAddress(), nil
}

func (r *addressRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AddressPO{}, id).Error
}

func (r *addressRepo) GetByID(ctx context.Context, id uint) (*biz.Address, error) {
	var po AddressPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toAddress(), nil
}

func (r *addressRepo) ListByUserID(ctx context.Context, userID uint) ([]*biz.Address, error) {
	var pos []*AddressPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	addrs := make([]*biz.Address, 0, len(pos))
	for _, po := range pos {
		addrs = append(addrs, po.toAddress())
	}
	return addrs, nil
}

// externalAssociationRepo 外部关联仓储实现
type externalAssociationRepo struct {
	db *gorm.DB
}

// NewExternalAssociationRepo 创建外部关联仓储实例
func NewExternalAssociationRepo(db *gorm.DB) biz.ExternalAssociationRepository {
	return &externalAssociationRepo{db: db}
}

func (r *externalAssociationRepo) Create(ctx context.Context, ea *biz.ExternalAssociation) (*biz.ExternalAssociation, error) {
	po := toExternalAssociationPO(ea)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.toExternalAssociation(), nil
}

func (r *externalAssociationRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ExternalAssociationPO{}, id).Error
}

func (r *externalAssociationRepo) GetByID(ctx context.Context, id uint) (*biz.ExternalAssociation, error) {
	var po ExternalAssociationPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toExternalAssociation(), nil
}

func (r *externalAssociationRepo) ListByUserID(ctx context.Context, userID uint) ([]*biz.ExternalAssociation, error) {
	var pos []*ExternalAssociationPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	eas := make([]*biz.ExternalAssociation, 0, len(pos))
	for _, po := range pos {
		eas = append(eas, po.toExternalAssociation())
	}
	return eas, nil
}

// downloadableProductRepo 可下载产品仓储实现
type downloadableProductRepo struct {
	db *gorm.DB
}

// NewDownloadableProductRepo 创建可下载产品仓储实例
func NewDownloadableProductRepo(db *gorm.DB) biz.DownloadableProductRepository {
	return &downloadableProductRepo{db: db}
}

func (r *downloadableProductRepo) Create(ctx context.Context, dp *biz.DownloadableProduct) (*biz.DownloadableProduct, error) {
	po := toDownloadableProductPO(dp)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.toDownloadableProduct(), nil
}

func (r *downloadableProductRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&DownloadableProductPO{}, id).Error
}

func (r *downloadableProductRepo) GetByID(ctx context.Context, id uint) (*biz.DownloadableProduct, error) {
	var po DownloadableProductPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toDownloadableProduct(), nil
}

func (r *downloadableProductRepo) ListByUserID(ctx context.Context, userID uint) ([]*biz.DownloadableProduct, error) {
	var pos []*DownloadableProductPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	dps := make([]*biz.DownloadableProduct, 0, len(pos))
	for _, po := range pos {
		dps = append(dps, po.toDownloadableProduct())
	}
	return dps, nil
}
