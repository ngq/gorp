// Package service 通知服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/notification-service/internal/biz"
	"nop-go/services/notification-service/internal/models"

	"github.com/gin-gonic/gin"
)

type NotificationService struct {
	notificationUC *biz.NotificationUseCase
	templateUC     *biz.TemplateUseCase
}

func NewNotificationService(notificationUC *biz.NotificationUseCase, templateUC *biz.TemplateUseCase) *NotificationService {
	return &NotificationService{notificationUC: notificationUC, templateUC: templateUC}
}

func (s *NotificationService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		notifications := api.Group("/notifications")
		{
			notifications.POST("/email", s.SendEmail)
			notifications.POST("/sms", s.SendSMS)
			notifications.POST("/template", s.SendTemplate)
			notifications.GET("/:id", s.GetNotification)
		}

		templates := api.Group("/templates")
		{
			templates.GET("", s.ListTemplates)
			templates.POST("", s.CreateTemplate)
			templates.GET("/:code", s.GetTemplate)
			templates.PUT("/:id", s.UpdateTemplate)
			templates.DELETE("/:id", s.DeleteTemplate)
		}
	}
}

func (s *NotificationService) SendEmail(c *gin.Context) {
	var req models.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.notificationUC.SendEmail(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "email queued"})
}

func (s *NotificationService) SendSMS(c *gin.Context) {
	var req models.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.notificationUC.SendSMS(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sms queued"})
}

func (s *NotificationService) SendTemplate(c *gin.Context) {
	var req models.SendTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.notificationUC.SendTemplate(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notification queued"})
}

func (s *NotificationService) GetNotification(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	n, err := s.notificationUC.GetNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToNotificationResponse(n))
}

func (s *NotificationService) ListTemplates(c *gin.Context) {
	list, err := s.templateUC.ListTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *NotificationService) CreateTemplate(c *gin.Context) {
	var tmpl models.NotificationTemplate
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.templateUC.CreateTemplate(c.Request.Context(), &tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tmpl)
}

func (s *NotificationService) GetTemplate(c *gin.Context) {
	code := c.Param("code")
	tmpl, err := s.templateUC.GetTemplate(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

func (s *NotificationService) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var tmpl models.NotificationTemplate
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tmpl.ID = id
	if err := s.templateUC.UpdateTemplate(c.Request.Context(), &tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

func (s *NotificationService) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.templateUC.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}