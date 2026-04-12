// Package service 客户服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/serviceauth/token"
	"nop-go/services/customer-service/internal/biz"
	"nop-go/services/customer-service/internal/models"
	shareModels "nop-go/shared/models"
)

// CustomerService 客户服务
type CustomerService struct {
	customerUC *biz.CustomerUseCase
	addressUC  *biz.AddressUseCase
	gdprUC     *biz.GdprUseCase
	jwtSvc     contract.JWTService
}

// NewCustomerService 创建客户服务。
//
// 中文说明：
// - 使用 framework 级 JWTService，替代项目层 jwtSecret；
// - 中间件改用 framework 提供的 AuthMiddleware。
func NewCustomerService(customerUC *biz.CustomerUseCase, addressUC *biz.AddressUseCase, gdprUC *biz.GdprUseCase, jwtSvc contract.JWTService) *CustomerService {
	return &CustomerService{
		customerUC: customerUC,
		addressUC:  addressUC,
		gdprUC:     gdprUC,
		jwtSvc:     jwtSvc,
	}
}

// RegisterRoutes 注册路由
func (s *CustomerService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	// 使用 framework JWT middleware
	customerAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "customer")
	{
		// 认证相关
		api.POST("/auth/register", s.Register)
		api.POST("/auth/login", s.Login)
		api.GET("/auth/me", customerAuth, s.GetCurrentUser)

		// 客户管理
		customers := api.Group("/customers")
		customers.Use(customerAuth)
		{
			customers.GET("/:id", s.GetCustomer)
			customers.GET("/:id/validate", s.ValidateCustomer)
			customers.PUT("/:id", s.UpdateProfile)
			customers.PUT("/:id/password", s.ChangePassword)
			customers.GET("", s.ListCustomers)
		}

		// 地址管理
		addresses := api.Group("/customers/:customer_id/addresses")
		addresses.Use(customerAuth)
		{
			addresses.GET("", s.GetAddresses)
			addresses.POST("", s.CreateAddress)
			addresses.GET("/:id", s.GetAddress)
			addresses.PUT("/:id", s.UpdateAddress)
			addresses.DELETE("/:id", s.DeleteAddress)
			addresses.PUT("/:id/default-billing", s.SetDefaultBilling)
			addresses.PUT("/:id/default-shipping", s.SetDefaultShipping)
		}

		// GDPR 合规管理
		gdpr := api.Group("/gdpr")
		{
			// GDPR 同意项管理（管理员）
			gdpr.GET("/consents", s.ListConsents)
			gdpr.POST("/consents", s.CreateConsent)
			gdpr.GET("/consents/:id", s.GetConsent)
			gdpr.PUT("/consents/:id", s.UpdateConsent)
			gdpr.DELETE("/consents/:id", s.DeleteConsent)

			// 客户同意操作
			gdpr.POST("/accept", customerAuth, s.AcceptConsent)
			gdpr.GET("/customers/:customer_id/consents", customerAuth, s.GetCustomerConsents)

			// 数据导出/删除请求
			gdpr.POST("/export", customerAuth, s.RequestDataExport)
			gdpr.POST("/delete", customerAuth, s.RequestDataDeletion)
			gdpr.GET("/customers/:customer_id/export", customerAuth, s.ExportCustomerData)

			// GDPR 请求管理（管理员）
			gdpr.GET("/requests", s.ListGdprRequests)
			gdpr.POST("/requests/:id/process", s.ProcessGdprRequest)

			// GDPR 日志
			gdpr.GET("/customers/:customer_id/logs", customerAuth, s.GetGdprLogs)
		}
	}
}

// Register 注册
func (s *CustomerService) Register(c *gin.Context) {
	var req biz.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := s.customerUC.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ToCustomerResponse(customer))
}

// Login 登录
func (s *CustomerService) Login(c *gin.Context) {
	var req biz.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.customerUC.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer": ToCustomerResponse(result.Customer),
		"token":    result.Token,
	})
}

// GetCurrentUser 获取当前用户
func (s *CustomerService) GetCurrentUser(c *gin.Context) {
	subjectID, ok := jwtmiddleware.SubjectIDFromContext(c)
	if !ok || subjectID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}
	customer, err := s.customerUC.GetByID(c.Request.Context(), uint64(subjectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ToCustomerResponse(customer))
}

// GetCustomer 获取客户
func (s *CustomerService) GetCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	customer, err := s.customerUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToCustomerResponse(customer))
}

// ValidateCustomer 验证客户（供其他服务调用）
func (s *CustomerService) ValidateCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.customerUC.ValidateCustomer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// UpdateProfile 更新资料
func (s *CustomerService) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		FirstName string     `json:"first_name"`
		LastName  string     `json:"last_name"`
		Gender    string     `json:"gender"`
		Birthday  *string    `json:"birthday"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := s.customerUC.UpdateProfile(c.Request.Context(), id, req.FirstName, req.LastName, req.Gender, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToCustomerResponse(customer))
}

// ChangePassword 修改密码
func (s *CustomerService) ChangePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.customerUC.ChangePassword(c.Request.Context(), id, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

// ListCustomers 客户列表
func (s *CustomerService) ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	customers, total, err := s.customerUC.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]CustomerResponse, len(customers))
	for i, c := range customers {
		items[i] = ToCustomerResponse(c)
	}

	c.JSON(http.StatusOK, shareModels.NewPagingResponse(items, total, page, pageSize))
}

// GetAddresses 获取地址列表
func (s *CustomerService) GetAddresses(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	addresses, err := s.addressUC.GetCustomerAddresses(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]AddressResponse, len(addresses))
	for i, a := range addresses {
		items[i] = ToAddressResponse(a)
	}

	c.JSON(http.StatusOK, items)
}

// CreateAddress 创建地址
func (s *CustomerService) CreateAddress(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	var req AddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address := &models.Address{
		CustomerID:  customerID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		Company:     req.Company,
		Country:     req.Country,
		State:       req.State,
		City:        req.City,
		Address1:    req.Address1,
		Address2:    req.Address2,
		ZipCode:     req.ZipCode,
	}

	if err := s.addressUC.CreateAddress(c.Request.Context(), address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ToAddressResponse(address))
}

// GetAddress 获取地址
func (s *CustomerService) GetAddress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	address, err := s.addressUC.GetAddressByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToAddressResponse(address))
}

// UpdateAddress 更新地址
func (s *CustomerService) UpdateAddress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req AddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address, err := s.addressUC.GetAddressByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	address.FirstName = req.FirstName
	address.LastName = req.LastName
	address.Email = req.Email
	address.Phone = req.Phone
	address.Company = req.Company
	address.Country = req.Country
	address.State = req.State
	address.City = req.City
	address.Address1 = req.Address1
	address.Address2 = req.Address2
	address.ZipCode = req.ZipCode

	if err := s.addressUC.UpdateAddress(c.Request.Context(), address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToAddressResponse(address))
}

// DeleteAddress 删除地址
func (s *CustomerService) DeleteAddress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.addressUC.DeleteAddress(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SetDefaultBilling 设置默认账单地址
func (s *CustomerService) SetDefaultBilling(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}
	addressID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.addressUC.SetDefaultBilling(c.Request.Context(), customerID, addressID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default billing address set"})
}

// SetDefaultShipping 设置默认配送地址
func (s *CustomerService) SetDefaultShipping(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}
	addressID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.addressUC.SetDefaultShipping(c.Request.Context(), customerID, addressID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default shipping address set"})
}

// DTO 定义

// CustomerResponse 客户响应
type CustomerResponse struct {
	ID            uint64 `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	FullName      string `json:"full_name"`
	Gender        string `json:"gender"`
	AvatarURL     string `json:"avatar_url"`
	IsActive      bool   `json:"is_active"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
	CreatedAt     string `json:"created_at"`
}

// ToCustomerResponse 转换为响应
func ToCustomerResponse(c *models.Customer) CustomerResponse {
	return CustomerResponse{
		ID:            c.ID,
		Username:      c.Username,
		Email:         c.Email,
		Phone:         c.Phone,
		FirstName:     c.FirstName,
		LastName:      c.LastName,
		FullName:      c.FullName(),
		Gender:        c.Gender,
		AvatarURL:     c.AvatarURL,
		IsActive:      c.IsActive,
		EmailVerified: c.EmailVerified,
		PhoneVerified: c.PhoneVerified,
		CreatedAt:     c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// AddressResponse 地址响应
type AddressResponse struct {
	ID                uint64 `json:"id"`
	CustomerID        uint64 `json:"customer_id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	FullName          string `json:"full_name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	Company           string `json:"company"`
	Country           string `json:"country"`
	State             string `json:"state"`
	City              string `json:"city"`
	Address1          string `json:"address1"`
	Address2          string `json:"address2"`
	ZipCode           string `json:"zip_code"`
	FullAddress       string `json:"full_address"`
	IsDefaultBilling  bool   `json:"is_default_billing"`
	IsDefaultShipping bool   `json:"is_default_shipping"`
}

// ToAddressResponse 转换为响应
func ToAddressResponse(a *models.Address) AddressResponse {
	return AddressResponse{
		ID:                a.ID,
		CustomerID:        a.CustomerID,
		FirstName:         a.FirstName,
		LastName:          a.LastName,
		FullName:          a.FirstName + " " + a.LastName,
		Email:             a.Email,
		Phone:             a.Phone,
		Company:           a.Company,
		Country:           a.Country,
		State:             a.State,
		City:              a.City,
		Address1:          a.Address1,
		Address2:          a.Address2,
		ZipCode:           a.ZipCode,
		FullAddress:       a.ToModel().FullAddress(),
		IsDefaultBilling:  a.IsDefaultBilling,
		IsDefaultShipping: a.IsDefaultShipping,
	}
}

// AddressCreateRequest 创建地址请求
type AddressCreateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Country   string `json:"country"`
	State     string `json:"state"`
	City      string `json:"city"`
	Address1  string `json:"address1" binding:"required"`
	Address2  string `json:"address2"`
	ZipCode   string `json:"zip_code"`
}

// ========== GDPR 处理器 ==========

func (s *CustomerService) ListConsents(c *gin.Context) {
	list, err := s.gdprUC.ListConsents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CustomerService) CreateConsent(c *gin.Context) {
	var req models.GdprConsentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	consent, err := s.gdprUC.CreateConsent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, consent)
}

func (s *CustomerService) GetConsent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	consent, err := s.gdprUC.GetConsent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consent)
}

func (s *CustomerService) UpdateConsent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.GdprConsentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	consent, err := s.gdprUC.UpdateConsent(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consent)
}

func (s *CustomerService) DeleteConsent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.gdprUC.DeleteConsent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CustomerService) AcceptConsent(c *gin.Context) {
	var req models.CustomerConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.IpAddress = c.ClientIP()
	if err := s.gdprUC.AcceptConsent(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "consent recorded"})
}

func (s *CustomerService) GetCustomerConsents(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	consents, err := s.gdprUC.GetCustomerConsents(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consents)
}

func (s *CustomerService) RequestDataExport(c *gin.Context) {
	var req models.GdprExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gdprReq, err := s.gdprUC.RequestDataExport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gdprReq)
}

func (s *CustomerService) RequestDataDeletion(c *gin.Context) {
	var req models.GdprDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gdprReq, err := s.gdprUC.RequestDataDeletion(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gdprReq)
}

func (s *CustomerService) ExportCustomerData(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	data, err := s.gdprUC.ExportCustomerData(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (s *CustomerService) ListGdprRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	requests, total, err := s.gdprUC.GetGdprRequests(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": requests, "total": total})
}

func (s *CustomerService) ProcessGdprRequest(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	approve := c.Query("approve") == "true"
	if err := s.gdprUC.ProcessGdprRequest(c.Request.Context(), id, approve); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "request processed"})
}

func (s *CustomerService) GetGdprLogs(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	logs, err := s.gdprUC.GetGdprLogs(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}