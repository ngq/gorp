// Package service 物流服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/shipping-service/internal/biz"
	"nop-go/services/shipping-service/internal/models"

	"github.com/gin-gonic/gin"
)

type ShippingService struct {
	shipmentUC *biz.ShipmentUseCase
	methodUC   *biz.ShippingMethodUseCase
}

func NewShippingService(shipmentUC *biz.ShipmentUseCase, methodUC *biz.ShippingMethodUseCase) *ShippingService {
	return &ShippingService{shipmentUC: shipmentUC, methodUC: methodUC}
}

func (s *ShippingService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		shipments := api.Group("/shipments")
		{
			shipments.GET("", s.ListShipments)
			shipments.POST("", s.CreateShipment)
			shipments.GET("/:id", s.GetShipment)
			shipments.GET("/order/:order_id", s.GetShipmentByOrderID)
			shipments.PUT("/:id/tracking", s.UpdateTracking)
		}

		methods := api.Group("/shipping-methods")
		{
			methods.GET("", s.ListMethods)
			methods.POST("", s.CreateMethod)
			methods.GET("/:id", s.GetMethod)
			methods.PUT("/:id", s.UpdateMethod)
			methods.DELETE("/:id", s.DeleteMethod)
		}
	}
}

func (s *ShippingService) CreateShipment(c *gin.Context) {
	var req models.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shipment, err := s.shipmentUC.CreateShipment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.ToShipmentResponse(shipment))
}

func (s *ShippingService) GetShipment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	shipment, err := s.shipmentUC.GetShipment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToShipmentResponse(shipment))
}

func (s *ShippingService) GetShipmentByOrderID(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("order_id"), 10, 64)
	shipment, err := s.shipmentUC.GetShipmentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToShipmentResponse(shipment))
}

func (s *ShippingService) UpdateTracking(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.UpdateTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.shipmentUC.UpdateTracking(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tracking updated"})
}

func (s *ShippingService) ListShipments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := s.shipmentUC.ListShipments(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.ShipmentResponse, len(list))
	for i, s := range list {
		items[i] = models.ToShipmentResponse(s)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (s *ShippingService) ListMethods(c *gin.Context) {
	list, err := s.methodUC.ListMethods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *ShippingService) CreateMethod(c *gin.Context) {
	var m models.ShippingMethod
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.methodUC.CreateMethod(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (s *ShippingService) GetMethod(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	m, err := s.methodUC.GetMethod(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (s *ShippingService) UpdateMethod(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var m models.ShippingMethod
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.ID = id
	if err := s.methodUC.UpdateMethod(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (s *ShippingService) DeleteMethod(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.methodUC.DeleteMethod(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}