// Package service 店铺服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/serviceauth/token"
	"nop-go/services/store-service/internal/biz"
	"nop-go/services/store-service/internal/models"
)

// StoreService 店铺服务
type StoreService struct {
	storeUC  *biz.StoreUseCase
	vendorUC *biz.VendorUseCase
	jwtSvc   contract.JWTService
}

// NewStoreService 创建店铺服务
func NewStoreService(storeUC *biz.StoreUseCase, vendorUC *biz.VendorUseCase, jwtSvc contract.JWTService) *StoreService {
	return &StoreService{storeUC: storeUC, vendorUC: vendorUC, jwtSvc: jwtSvc}
}

// RegisterRoutes 注册路由
func (s *StoreService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		// 店铺管理
		stores := api.Group("/stores")
		stores.Use(adminAuth)
		{
			stores.GET("", s.ListStores)
			stores.POST("", s.CreateStore)
			stores.GET("/:id", s.GetStore)
			stores.PUT("/:id", s.UpdateStore)
			stores.DELETE("/:id", s.DeleteStore)
			stores.GET("/:id/vendors", s.GetStoreVendors)
		}

		// 供应商管理
		vendors := api.Group("/vendors")
		vendors.Use(adminAuth)
		{
			vendors.GET("", s.ListVendors)
			vendors.POST("", s.CreateVendor)
			vendors.GET("/:id", s.GetVendor)
			vendors.PUT("/:id", s.UpdateVendor)
			vendors.DELETE("/:id", s.DeleteVendor)
			vendors.POST("/:id/stores/:store_id", s.AddVendorToStore)
			vendors.DELETE("/:id/stores/:store_id", s.RemoveVendorFromStore)
			vendors.POST("/:id/notes", s.AddVendorNote)
			vendors.GET("/:id/notes", s.GetVendorNotes)
		}
	}
}

// ================== 店铺接口 ==================

func (s *StoreService) CreateStore(c *gin.Context) {
	var req models.StoreCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store, err := s.storeUC.CreateStore(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, store)
}

func (s *StoreService) GetStore(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	store, err := s.storeUC.GetStore(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, store)
}

func (s *StoreService) ListStores(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	stores, total, err := s.storeUC.ListStores(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": stores, "total": total})
}

func (s *StoreService) UpdateStore(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.StoreUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store, err := s.storeUC.UpdateStore(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, store)
}

func (s *StoreService) DeleteStore(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.storeUC.DeleteStore(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *StoreService) GetStoreVendors(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	vendors, err := s.storeUC.GetStoreVendors(c.Request.Context(), uint(storeID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vendors)
}

// ================== 供应商接口 ==================

func (s *StoreService) CreateVendor(c *gin.Context) {
	var req models.VendorCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendor, err := s.vendorUC.CreateVendor(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, vendor)
}

func (s *StoreService) GetVendor(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	vendor, err := s.vendorUC.GetVendor(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vendor)
}

func (s *StoreService) ListVendors(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	vendors, total, err := s.vendorUC.ListVendors(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": vendors, "total": total})
}

func (s *StoreService) UpdateVendor(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.VendorUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendor, err := s.vendorUC.UpdateVendor(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vendor)
}

func (s *StoreService) DeleteVendor(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.vendorUC.DeleteVendor(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *StoreService) AddVendorToStore(c *gin.Context) {
	vendorID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	storeID, _ := strconv.ParseUint(c.Param("store_id"), 10, 64)

	var req struct {
		IsDefault bool `json:"is_default"`
	}
	c.ShouldBindJSON(&req)

	if err := s.vendorUC.AddVendorToStore(c.Request.Context(), uint(storeID), uint(vendorID), req.IsDefault); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vendor added to store"})
}

func (s *StoreService) RemoveVendorFromStore(c *gin.Context) {
	vendorID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	storeID, _ := strconv.ParseUint(c.Param("store_id"), 10, 64)

	if err := s.vendorUC.RemoveVendorFromStore(c.Request.Context(), uint(storeID), uint(vendorID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *StoreService) AddVendorNote(c *gin.Context) {
	vendorID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.vendorUC.AddVendorNote(c.Request.Context(), uint(vendorID), req.Note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note added"})
}

func (s *StoreService) GetVendorNotes(c *gin.Context) {
	vendorID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	notes, err := s.vendorUC.GetVendorNotes(c.Request.Context(), uint(vendorID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notes)
}