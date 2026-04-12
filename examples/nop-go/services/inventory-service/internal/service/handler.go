// Package service 库存服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/inventory-service/internal/biz"
	"nop-go/services/inventory-service/internal/models"

	"github.com/gin-gonic/gin"
)

type InventoryService struct {
	inventoryUC *biz.InventoryUseCase
	warehouseUC *biz.WarehouseUseCase
}

func NewInventoryService(inventoryUC *biz.InventoryUseCase, warehouseUC *biz.WarehouseUseCase) *InventoryService {
	return &InventoryService{inventoryUC: inventoryUC, warehouseUC: warehouseUC}
}

func (s *InventoryService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/inventory/:id", s.GetInventory)
		api.GET("/inventory/product/:product_id", s.GetInventoryByProduct)
		api.POST("/inventory/reserve", s.ReserveStock)
		api.POST("/inventory/confirm/:order_id", s.ConfirmStock)
		api.POST("/inventory/release/:order_id", s.ReleaseStock)
		api.POST("/inventory/adjust", s.AdjustStock)

		warehouses := api.Group("/warehouses")
		{
			warehouses.GET("", s.ListWarehouses)
			warehouses.POST("", s.CreateWarehouse)
			warehouses.GET("/:id", s.GetWarehouse)
			warehouses.PUT("/:id", s.UpdateWarehouse)
			warehouses.DELETE("/:id", s.DeleteWarehouse)
		}
	}
}

func (s *InventoryService) GetInventory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	inv, err := s.inventoryUC.GetInventory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToInventoryResponse(inv))
}

func (s *InventoryService) GetInventoryByProduct(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)
	list, err := s.inventoryUC.GetByProductID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.InventoryResponse, len(list))
	for i, inv := range list {
		items[i] = models.ToInventoryResponse(inv)
	}
	c.JSON(http.StatusOK, items)
}

func (s *InventoryService) ReserveStock(c *gin.Context) {
	var req models.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 转换为 biz 的请求类型
	bizReq := &biz.ReserveStockRequest{
		OrderID:     req.OrderID,
		ProductID:   req.ProductID,
		WarehouseID: req.WarehouseID,
		Quantity:    req.Quantity,
	}
	if err := s.inventoryUC.ReserveStock(c.Request.Context(), bizReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stock reserved"})
}

func (s *InventoryService) ConfirmStock(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err := s.inventoryUC.ConfirmStock(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stock confirmed"})
}

func (s *InventoryService) ReleaseStock(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err := s.inventoryUC.ReleaseStock(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stock released"})
}

func (s *InventoryService) AdjustStock(c *gin.Context) {
	var req models.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.inventoryUC.AdjustStock(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stock adjusted"})
}

func (s *InventoryService) ListWarehouses(c *gin.Context) {
	list, err := s.warehouseUC.ListWarehouses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *InventoryService) CreateWarehouse(c *gin.Context) {
	var w models.Warehouse
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.warehouseUC.CreateWarehouse(c.Request.Context(), &w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (s *InventoryService) GetWarehouse(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	w, err := s.warehouseUC.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (s *InventoryService) UpdateWarehouse(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var w models.Warehouse
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w.ID = id
	if err := s.warehouseUC.UpdateWarehouse(c.Request.Context(), &w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (s *InventoryService) DeleteWarehouse(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.warehouseUC.DeleteWarehouse(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}