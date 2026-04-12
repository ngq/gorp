// Package service 订单服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/order-service/internal/biz"
	"nop-go/services/order-service/internal/models"

	"github.com/gin-gonic/gin"
)

type OrderService struct {
	orderUC    *biz.OrderUseCase
	giftCardUC *biz.GiftCardUseCase
	returnUC   *biz.ReturnRequestUseCase
}

func NewOrderService(orderUC *biz.OrderUseCase, giftCardUC *biz.GiftCardUseCase, returnUC *biz.ReturnRequestUseCase) *OrderService {
	return &OrderService{orderUC: orderUC, giftCardUC: giftCardUC, returnUC: returnUC}
}

func (s *OrderService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.GET("", s.ListOrders)
			orders.POST("", s.CreateOrder)
			orders.GET("/:id", s.GetOrder)
			orders.PUT("/:id/status", s.UpdateOrderStatus)
			orders.POST("/:id/cancel", s.CancelOrder)
		}

		api.GET("/customers/:customer_id/orders", s.GetCustomerOrders)

		giftcards := api.Group("/giftcards")
		{
			giftcards.GET("/:code", s.GetGiftCard)
			giftcards.POST("/:code/redeem", s.RedeemGiftCard)
		}

		returns := api.Group("/returns")
		{
			returns.POST("", s.CreateReturnRequest)
			returns.GET("/:id", s.GetReturnRequest)
			returns.PUT("/:id/approve", s.ApproveReturnRequest)
		}
	}
}

func (s *OrderService) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := s.orderUC.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.ToOrderResponse(order))
}

func (s *OrderService) GetOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := s.orderUC.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToOrderResponse(order))
}

func (s *OrderService) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	orders, total, err := s.orderUC.ListOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.OrderResponse, len(orders))
	for i, o := range orders {
		items[i] = models.ToOrderResponse(o)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (s *OrderService) GetCustomerOrders(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	orders, total, err := s.orderUC.GetCustomerOrders(c.Request.Context(), customerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.OrderResponse, len(orders))
	for i, o := range orders {
		items[i] = models.ToOrderResponse(o)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (s *OrderService) UpdateOrderStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.orderUC.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (s *OrderService) CancelOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.orderUC.CancelOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
}

func (s *OrderService) GetGiftCard(c *gin.Context) {
	code := c.Param("code")
	card, err := s.giftCardUC.ValidateGiftCard(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": card.Code, "amount": card.Amount, "remaining": card.RemainingAmount})
}

func (s *OrderService) RedeemGiftCard(c *gin.Context) {
	code := c.Param("code")
	var req struct {
		CustomerID uint64 `json:"customer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.giftCardUC.RedeemGiftCard(c.Request.Context(), code, req.CustomerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "gift card redeemed"})
}

func (s *OrderService) CreateReturnRequest(c *gin.Context) {
	var req models.ReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.returnUC.CreateReturnRequest(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (s *OrderService) GetReturnRequest(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	req, err := s.returnUC.GetReturnRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (s *OrderService) ApproveReturnRequest(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		AdminID uint64 `json:"admin_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.returnUC.ApproveReturnRequest(c.Request.Context(), id, req.AdminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "return approved"})
}