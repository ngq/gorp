// Package biz 瀹㈡埛鏈嶅姟涓氬姟閫昏緫灞?
package biz

import (
	"context"
	"errors"
	"time"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"nop-go/services/customer-service/internal/data"
	"nop-go/services/customer-service/internal/models"
	shareErrors "nop-go/shared/errors"

	"golang.org/x/crypto/bcrypt"
)

// CustomerUseCase 瀹㈡埛鐢ㄤ緥
type CustomerUseCase struct {
	customerRepo data.CustomerRepository
	addressRepo  data.AddressRepository
	roleRepo     data.CustomerRoleRepository
	jwtSvc       securitycontract.JWTService
}

// NewCustomerUseCase 鍒涘缓瀹㈡埛鐢ㄤ緥銆?
//
// 涓枃璇存槑锛?
// - 浣跨敤 framework 绾?JWTService锛屾浛浠ｉ」鐩眰 jwtSecret/jwtExpire锛?
// - JWTService 缁熶竴澶勭悊绛惧彂/楠岃瘉锛岄厤缃粠 auth.jwt.* 璇诲彇銆?
func NewCustomerUseCase(
	customerRepo data.CustomerRepository,
	addressRepo data.AddressRepository,
	roleRepo data.CustomerRoleRepository,
	jwtSvc securitycontract.JWTService,
) *CustomerUseCase {
	return &CustomerUseCase{
		customerRepo: customerRepo,
		addressRepo:  addressRepo,
		roleRepo:     roleRepo,
		jwtSvc:       jwtSvc,
	}
}

// RegisterRequest 娉ㄥ唽璇锋眰
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// Register 娉ㄥ唽瀹㈡埛
func (uc *CustomerUseCase) Register(ctx context.Context, req *RegisterRequest) (*models.Customer, error) {
	// 妫€鏌ョ敤鎴峰悕鏄惁宸插瓨鍦?
	if existing, _ := uc.customerRepo.GetByUsername(ctx, req.Username); existing != nil {
		return nil, shareErrors.ErrCustomerAlreadyExists
	}

	// 妫€鏌ラ偖绠辨槸鍚﹀凡瀛樺湪
	if existing, _ := uc.customerRepo.GetByEmail(ctx, req.Email); existing != nil {
		return nil, shareErrors.ErrCustomerAlreadyExists
	}

	// 瀵嗙爜鍝堝笇
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	customer := &models.Customer{
		Username:      req.Username,
		Email:         req.Email,
		Phone:         req.Phone,
		PasswordHash:  string(passwordHash),
		Gender:        "unknown",
		IsActive:      true,
		EmailVerified: false,
		PhoneVerified: false,
	}

	if err := uc.customerRepo.Create(ctx, customer); err != nil {
		return nil, err
	}

	// 娣诲姞榛樿瑙掕壊
	if role, _ := uc.roleRepo.GetBySystemName(ctx, models.RoleRegistered); role != nil {
		uc.customerRepo.AddRole(ctx, customer.ID, role.ID)
	}

	return customer, nil
}

// LoginRequest 鐧诲綍璇锋眰
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResult 鐧诲綍缁撴灉
type LoginResult struct {
	Customer *models.Customer
	Token    string
}

// Login 鐧诲綍
func (uc *CustomerUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResult, error) {
	// 灏濊瘯閫氳繃鐢ㄦ埛鍚嶆煡鎵?
	customer, err := uc.customerRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// 灏濊瘯閫氳繃閭鏌ユ壘
		customer, err = uc.customerRepo.GetByEmail(ctx, req.Username)
		if err != nil {
			return nil, shareErrors.ErrInvalidCredentials
		}
	}

	// 楠岃瘉瀵嗙爜
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password)); err != nil {
		return nil, shareErrors.ErrInvalidCredentials
	}

	// 妫€鏌ヨ处鎴风姸鎬?
	if !customer.IsActive {
		return nil, shareErrors.ErrCustomerDisabled
	}

	// 鏇存柊鏈€鍚庣櫥褰曚俊鎭?
	now := time.Now()
	customer.LastLoginAt = &now
	uc.customerRepo.Update(ctx, customer)

	roles := make([]string, 0, len(customer.Roles))
	for _, role := range customer.Roles {
		roles = append(roles, role.SystemName)
	}

	// 浣跨敤 framework JWTService 绛惧彂 token
	claims := uc.jwtSvc.NewClaims(int64(customer.ID), "customer", customer.Username, roles, 86400)
	token, err := uc.jwtSvc.Sign(claims)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Customer: customer,
		Token:    token,
	}, nil
}

// GetByID 鏍规嵁ID鑾峰彇瀹㈡埛
func (uc *CustomerUseCase) GetByID(ctx context.Context, id uint64) (*models.Customer, error) {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrCustomerNotFound
	}
	return customer, nil
}

// GetByEmail 鏍规嵁閭鑾峰彇瀹㈡埛
func (uc *CustomerUseCase) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	customer, err := uc.customerRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, shareErrors.ErrCustomerNotFound
	}
	return customer, nil
}

// UpdateProfile 鏇存柊瀹㈡埛璧勬枡
func (uc *CustomerUseCase) UpdateProfile(ctx context.Context, id uint64, firstName, lastName, gender string, birthday *time.Time) (*models.Customer, error) {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrCustomerNotFound
	}

	if firstName != "" {
		customer.FirstName = firstName
	}
	if lastName != "" {
		customer.LastName = lastName
	}
	if gender != "" {
		customer.Gender = gender
	}
	if birthday != nil {
		customer.Birthday = birthday
	}

	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// ChangePassword 淇敼瀵嗙爜
func (uc *CustomerUseCase) ChangePassword(ctx context.Context, id uint64, oldPassword, newPassword string) error {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrCustomerNotFound
	}

	// 楠岃瘉鏃у瘑鐮?
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(oldPassword)); err != nil {
		return shareErrors.ErrPasswordMismatch
	}

	// 鍝堝笇鏂板瘑鐮?
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	customer.PasswordHash = string(passwordHash)
	return uc.customerRepo.Update(ctx, customer)
}

// List 瀹㈡埛鍒楄〃
func (uc *CustomerUseCase) List(ctx context.Context, page, pageSize int) ([]*models.Customer, int64, error) {
	return uc.customerRepo.List(ctx, page, pageSize)
}

// ValidateCustomer 楠岃瘉瀹㈡埛锛堜緵鍏朵粬鏈嶅姟璋冪敤锛?
func (uc *CustomerUseCase) ValidateCustomer(ctx context.Context, id uint64) error {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrCustomerNotFound
	}

	if !customer.IsActive {
		return shareErrors.ErrCustomerDisabled
	}

	return nil
}

// AddressUseCase 鍦板潃鐢ㄤ緥
type AddressUseCase struct {
	addressRepo data.AddressRepository
}

// NewAddressUseCase 鍒涘缓鍦板潃鐢ㄤ緥
func NewAddressUseCase(addressRepo data.AddressRepository) *AddressUseCase {
	return &AddressUseCase{addressRepo: addressRepo}
}

// CreateAddress 鍒涘缓鍦板潃
func (uc *AddressUseCase) CreateAddress(ctx context.Context, address *models.Address) error {
	return uc.addressRepo.Create(ctx, address)
}

// GetAddressByID 鑾峰彇鍦板潃
func (uc *AddressUseCase) GetAddressByID(ctx context.Context, id uint64) (*models.Address, error) {
	address, err := uc.addressRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrAddressNotFound
	}
	return address, nil
}

// GetCustomerAddresses 鑾峰彇瀹㈡埛鎵€鏈夊湴鍧€
func (uc *AddressUseCase) GetCustomerAddresses(ctx context.Context, customerID uint64) ([]*models.Address, error) {
	return uc.addressRepo.GetByCustomerID(ctx, customerID)
}

// UpdateAddress 鏇存柊鍦板潃
func (uc *AddressUseCase) UpdateAddress(ctx context.Context, address *models.Address) error {
	_, err := uc.addressRepo.GetByID(ctx, address.ID)
	if err != nil {
		return shareErrors.ErrAddressNotFound
	}
	return uc.addressRepo.Update(ctx, address)
}

// DeleteAddress 鍒犻櫎鍦板潃
func (uc *AddressUseCase) DeleteAddress(ctx context.Context, id uint64) error {
	return uc.addressRepo.Delete(ctx, id)
}

// SetDefaultBilling 璁剧疆榛樿璐﹀崟鍦板潃
func (uc *AddressUseCase) SetDefaultBilling(ctx context.Context, customerID, addressID uint64) error {
	return uc.addressRepo.SetDefaultBilling(ctx, customerID, addressID)
}

// SetDefaultShipping 璁剧疆榛樿閰嶉€佸湴鍧€
func (uc *AddressUseCase) SetDefaultShipping(ctx context.Context, customerID, addressID uint64) error {
	return uc.addressRepo.SetDefaultShipping(ctx, customerID, addressID)
}

// IsErrCustomerNotFound 鍒ゆ柇鏄惁涓哄鎴蜂笉瀛樺湪閿欒
func IsErrCustomerNotFound(err error) bool {
	return errors.Is(err, shareErrors.ErrCustomerNotFound)
}

// ========== GDPR 鐢ㄤ緥 ==========

// GdprUseCase GDPR鐢ㄤ緥
type GdprUseCase struct {
	consentRepo     data.GdprConsentRepository
	logRepo         data.GdprLogRepository
	requestRepo     data.GdprRequestRepository
	customerConsent data.CustomerConsentRepository
	customerRepo    data.CustomerRepository
	addressRepo     data.AddressRepository
}

// NewGdprUseCase 鍒涘缓GDPR鐢ㄤ緥
func NewGdprUseCase(
	consentRepo data.GdprConsentRepository,
	logRepo data.GdprLogRepository,
	requestRepo data.GdprRequestRepository,
	customerConsent data.CustomerConsentRepository,
	customerRepo data.CustomerRepository,
	addressRepo data.AddressRepository,
) *GdprUseCase {
	return &GdprUseCase{
		consentRepo:     consentRepo,
		logRepo:         logRepo,
		requestRepo:     requestRepo,
		customerConsent: customerConsent,
		customerRepo:    customerRepo,
		addressRepo:     addressRepo,
	}
}

// CreateConsent 鍒涘缓GDPR鍚屾剰椤?
func (uc *GdprUseCase) CreateConsent(ctx context.Context, req *models.GdprConsentCreateRequest) (*models.GdprConsent, error) {
	consent := &models.GdprConsent{
		Message:         req.Message,
		IsRequired:      req.IsRequired,
		RequiredMessage: req.RequiredMessage,
		DisplayOrder:    req.DisplayOrder,
		IsActive:        true,
	}
	if err := uc.consentRepo.Create(ctx, consent); err != nil {
		return nil, err
	}
	return consent, nil
}

// GetConsent 鑾峰彇GDPR鍚屾剰椤?
func (uc *GdprUseCase) GetConsent(ctx context.Context, id uint64) (*models.GdprConsent, error) {
	return uc.consentRepo.GetByID(ctx, id)
}

// ListConsents GDPR鍚屾剰椤瑰垪琛?
func (uc *GdprUseCase) ListConsents(ctx context.Context) ([]*models.GdprConsent, error) {
	return uc.consentRepo.List(ctx)
}

// ListActiveConsents 鑾峰彇娲诲姩鐨凣DPR鍚屾剰椤?
func (uc *GdprUseCase) ListActiveConsents(ctx context.Context) ([]*models.GdprConsent, error) {
	return uc.consentRepo.ListActive(ctx)
}

// UpdateConsent 鏇存柊GDPR鍚屾剰椤?
func (uc *GdprUseCase) UpdateConsent(ctx context.Context, id uint64, req *models.GdprConsentUpdateRequest) (*models.GdprConsent, error) {
	consent, err := uc.consentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Message != "" {
		consent.Message = req.Message
	}
	consent.IsRequired = req.IsRequired
	consent.RequiredMessage = req.RequiredMessage
	consent.DisplayOrder = req.DisplayOrder
	consent.IsActive = req.IsActive
	if err := uc.consentRepo.Update(ctx, consent); err != nil {
		return nil, err
	}
	return consent, nil
}

// DeleteConsent 鍒犻櫎GDPR鍚屾剰椤?
func (uc *GdprUseCase) DeleteConsent(ctx context.Context, id uint64) error {
	return uc.consentRepo.Delete(ctx, id)
}

// AcceptConsent 瀹㈡埛鎺ュ彈鍚屾剰椤?
func (uc *GdprUseCase) AcceptConsent(ctx context.Context, req *models.CustomerConsentRequest) error {
	// 妫€鏌ュ悓鎰忛」鏄惁瀛樺湪
	consent, err := uc.consentRepo.GetByID(ctx, req.ConsentID)
	if err != nil {
		return err
	}

	// 妫€鏌ユ槸鍚﹀凡鏈夎褰?
	existing, err := uc.customerConsent.GetByCustomerAndConsent(ctx, req.CustomerID, req.ConsentID)
	if err == nil && existing != nil {
		// 鏇存柊鐜版湁璁板綍
		existing.IsAccepted = req.IsAccepted
		if req.IsAccepted {
			existing.AcceptedAt = time.Now()
		}
		return uc.customerConsent.Update(ctx, existing)
	}

	// 鍒涘缓鏂拌褰?
	cc := &models.CustomerConsent{
		CustomerID: req.CustomerID,
		ConsentID:  req.ConsentID,
		IsAccepted: req.IsAccepted,
	}
	if req.IsAccepted {
		cc.AcceptedAt = time.Now()
	}
	if err := uc.customerConsent.Create(ctx, cc); err != nil {
		return err
	}

	// 璁板綍鏃ュ織
	logType := 1 // 鍚屾剰
	if !req.IsAccepted {
		logType = 2 // 鎾ゅ洖
	}
	log := &models.GdprLog{
		CustomerID:   req.CustomerID,
		ConsentID:    req.ConsentID,
		RequestType:  logType,
		IpAddress:    req.IpAddress,
		CreatedOnUtc: time.Now().UTC(),
	}
	uc.logRepo.Create(ctx, log)

	_ = consent // 閬垮厤鏈娇鐢ㄨ鍛?
	return nil
}

// GetCustomerConsents 鑾峰彇瀹㈡埛鐨勫悓鎰忚褰?
func (uc *GdprUseCase) GetCustomerConsents(ctx context.Context, customerID uint64) ([]*models.CustomerConsent, error) {
	return uc.customerConsent.GetByCustomerID(ctx, customerID)
}

// RequestDataExport 璇锋眰鏁版嵁瀵煎嚭
func (uc *GdprUseCase) RequestDataExport(ctx context.Context, req *models.GdprExportRequest) (*models.GdprRequest, error) {
	gdprReq := &models.GdprRequest{
		CustomerID:   req.CustomerID,
		RequestType:  1, // 瀵煎嚭
		Status:       0, // 寰呭鐞?
		CreatedOnUtc: time.Now().UTC(),
	}
	if err := uc.requestRepo.Create(ctx, gdprReq); err != nil {
		return nil, err
	}
	return gdprReq, nil
}

// RequestDataDeletion 璇锋眰鏁版嵁鍒犻櫎
func (uc *GdprUseCase) RequestDataDeletion(ctx context.Context, req *models.GdprDeleteRequest) (*models.GdprRequest, error) {
	gdprReq := &models.GdprRequest{
		CustomerID:     req.CustomerID,
		RequestType:    2, // 鍒犻櫎
		RequestDetails: req.RequestDetails,
		Status:         0, // 寰呭鐞?
		CreatedOnUtc:   time.Now().UTC(),
	}
	if err := uc.requestRepo.Create(ctx, gdprReq); err != nil {
		return nil, err
	}
	return gdprReq, nil
}

// ExportCustomerData 瀵煎嚭瀹㈡埛鏁版嵁
func (uc *GdprUseCase) ExportCustomerData(ctx context.Context, customerID uint64) (*models.CustomerDataExport, error) {
	// 鑾峰彇瀹㈡埛淇℃伅
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 鑾峰彇鍦板潃
	addresses, err := uc.addressRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 鑾峰彇鍚屾剰璁板綍
	consents, err := uc.customerConsent.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 鑾峰彇鏃ュ織
	logs, err := uc.logRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return &models.CustomerDataExport{
		Customer:   customer,
		Addresses:  addresses,
		Consents:   consents,
		Logs:       logs,
		ExportedAt: time.Now(),
	}, nil
}

// DeleteCustomerData 鍒犻櫎瀹㈡埛鏁版嵁锛圙DPR鍚堣锛?
func (uc *GdprUseCase) DeleteCustomerData(ctx context.Context, customerID uint64) error {
	// 鍒犻櫎鍚屾剰璁板綍
	if err := uc.customerConsent.DeleteByCustomerID(ctx, customerID); err != nil {
		return err
	}

	// 鍒犻櫎瀹㈡埛锛堣蒋鍒犻櫎锛?
	if err := uc.customerRepo.Delete(ctx, customerID); err != nil {
		return err
	}

	return nil
}

// GetGdprRequests 鑾峰彇GDPR璇锋眰鍒楄〃
func (uc *GdprUseCase) GetGdprRequests(ctx context.Context, page, pageSize int) ([]*models.GdprRequest, int64, error) {
	return uc.requestRepo.List(ctx, page, pageSize)
}

// ProcessGdprRequest 澶勭悊GDPR璇锋眰
func (uc *GdprUseCase) ProcessGdprRequest(ctx context.Context, requestID uint64, approve bool) error {
	req, err := uc.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	if approve {
		req.Status = 1 // 宸插鐞?
		if req.RequestType == 2 {
			// 濡傛灉鏄垹闄よ姹傦紝鎵ц鍒犻櫎
			if err := uc.DeleteCustomerData(ctx, req.CustomerID); err != nil {
				return err
			}
		}
	} else {
		req.Status = 2 // 宸叉嫆缁?
	}

	req.ProcessedOnUtc = time.Now().UTC()
	return uc.requestRepo.Update(ctx, req)
}

// GetGdprLogs 鑾峰彇瀹㈡埛鐨凣DPR鏃ュ織
func (uc *GdprUseCase) GetGdprLogs(ctx context.Context, customerID uint64) ([]*models.GdprLog, error) {
	return uc.logRepo.GetByCustomerID(ctx, customerID)
}
