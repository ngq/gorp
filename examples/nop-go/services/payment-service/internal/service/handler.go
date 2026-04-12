// Package service 支付服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/payment-service/internal/biz"
	"nop-go/services/payment-service/internal/models"

	"github.com/gin-gonic/gin"
)

type PaymentService struct {
	paymentUC *biz.PaymentUseCase
}

func NewPaymentService(paymentUC *biz.PaymentUseCase) *PaymentService {
	return &PaymentService{paymentUC: paymentUC}
}

func (s *PaymentService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/payments/:id", s.GetPayment)
		api.GET("/payments/order/:order_id", s.GetPaymentByOrderID)
		api.POST("/payments", s.CreatePayment)
		api.POST("/payments/:id/pay", s.MarkAsPaid)
		api.POST("/payments/:id/fail", s.MarkAsFailed)
		api.POST("/payments/refund", s.Refund)
		api.POST("/refunds/:id/process", s.ProcessRefund)
	}
}

func (s *PaymentService) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, err := s.paymentUC.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.ToPaymentResponse(payment))
}

func (s *PaymentService) GetPayment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	payment, err := s.paymentUC.GetPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToPaymentResponse(payment))
}

func (s *PaymentService) GetPaymentByOrderID(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("order_id"), 10, 64)
	payment, err := s.paymentUC.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToPaymentResponse(payment))
}

func (s *PaymentService) MarkAsPaid(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		TransactionID string `json:"transaction_id"`
	}
	c.ShouldBindJSON(&req)
	if err := s.paymentUC.MarkAsPaid(c.Request.Context(), id, req.TransactionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "payment marked as paid"})
}

func (s *PaymentService) MarkAsFailed(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		ErrorCode    string `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}
	c.ShouldBindJSON(&req)
	if err := s.paymentUC.MarkAsFailed(c.Request.Context(), id, req.ErrorCode, req.ErrorMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "payment marked as failed"})
}

func (s *PaymentService) Refund(c *gin.Context) {
	var req models.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refund, err := s.paymentUC.Refund(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": refund.ID, "status": refund.Status})
}

func (s *PaymentService) ProcessRefund(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Success bool `json:"success"`
	}
	c.ShouldBindJSON(&req)
	if err := s.paymentUC.ProcessRefund(c.Request.Context(), id, req.Success); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "refund processed"})
}