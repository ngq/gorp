// Package service 鑱旂洘鎺ㄥ箍鏈嶅姟HTTP灞?
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/auth/jwt"
	"nop-go/services/affiliate-service/internal/biz"
	"nop-go/services/affiliate-service/internal/models"
)

// AffiliateService 鑱旂洘鎺ㄥ箍鏈嶅姟
type AffiliateService struct {
	affUC  *biz.AffiliateUseCase
	jwtSvc securitycontract.JWTService
}

// NewAffiliateService 鍒涘缓鑱旂洘鎺ㄥ箍鏈嶅姟
func NewAffiliateService(affUC *biz.AffiliateUseCase, jwtSvc securitycontract.JWTService) *AffiliateService {
	return &AffiliateService{affUC: affUC, jwtSvc: jwtSvc}
}

// RegisterRoutes 娉ㄥ唽璺敱
func (s *AffiliateService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/affiliate")
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		// 鑱旂洘浼氬憳绠＄悊
		api.POST("/affiliates", adminAuth, s.CreateAffiliate)
		api.GET("/affiliates", adminAuth, s.ListAffiliates)
		api.GET("/affiliates/search", adminAuth, s.SearchAffiliates)
		api.GET("/affiliates/:id", s.GetAffiliate)
		api.PUT("/affiliates/:id", adminAuth, s.UpdateAffiliate)
		api.DELETE("/affiliates/:id", adminAuth, s.DeleteAffiliate)
		api.POST("/affiliates/:id/activate", adminAuth, s.ActivateAffiliate)
		api.POST("/affiliates/:id/deactivate", adminAuth, s.DeactivateAffiliate)

		// 鑱旂洘鎺ㄨ崘杩借釜
		api.POST("/referrals/track", s.TrackReferral)
		api.POST("/referrals/convert", s.ConvertReferral)
		api.GET("/affiliates/:id/referrals", s.GetAffiliateReferrals)

		// 鑱旂洘璁㈠崟绠＄悊
		api.POST("/orders", s.CreateAffiliateOrder)
		api.GET("/affiliates/:id/orders", s.GetAffiliateOrders)

		// 浣ｉ噾绠＄悊
		api.POST("/commissions/calculate", s.CalculateCommission)
		api.GET("/affiliates/:id/commissions", s.GetAffiliateCommissions)
		api.GET("/affiliates/:id/balance", s.GetPendingBalance)

		// 鏀粯绠＄悊
		api.POST("/payouts", adminAuth, s.CreatePayout)
		api.POST("/payouts/:id/process", adminAuth, s.ProcessPayout)
		api.GET("/payouts/:id", s.GetPayout)
		api.GET("/affiliates/:id/payouts", s.GetAffiliatePayouts)

		// 缁熻淇℃伅
		api.GET("/affiliates/:id/stats", s.GetAffiliateStats)
	}
}

// ========== 鑱旂洘浼氬憳鎺ュ彛 ==========

func (s *AffiliateService) CreateAffiliate(c *gin.Context) {
	var req models.AffiliateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	affiliate, err := s.affUC.CreateAffiliate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, affiliate)
}

func (s *AffiliateService) GetAffiliate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	affiliate, err := s.affUC.GetAffiliate(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, affiliate)
}

func (s *AffiliateService) ListAffiliates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	affiliates, total, err := s.affUC.ListAffiliates(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": affiliates, "total": total})
}

func (s *AffiliateService) SearchAffiliates(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	affiliates, total, err := s.affUC.SearchAffiliates(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": affiliates, "total": total})
}

func (s *AffiliateService) UpdateAffiliate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.AffiliateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	affiliate, err := s.affUC.UpdateAffiliate(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, affiliate)
}

func (s *AffiliateService) DeleteAffiliate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.affUC.DeleteAffiliate(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *AffiliateService) ActivateAffiliate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.affUC.ActivateAffiliate(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "affiliate activated"})
}

func (s *AffiliateService) DeactivateAffiliate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.affUC.DeactivateAffiliate(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "affiliate deactivated"})
}

// ========== 鑱旂洘鎺ㄨ崘鎺ュ彛 ==========

func (s *AffiliateService) TrackReferral(c *gin.Context) {
	affiliateID, _ := strconv.ParseUint(c.Query("affiliate_id"), 10, 64)
	sessionID := c.Query("session_id")
	referrerURL := c.Query("referrer_url")
	ipAddress := c.Query("ip_address")
	customerID, _ := strconv.ParseUint(c.Query("customer_id"), 10, 64)

	if affiliateID == 0 || sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "affiliate_id and session_id are required"})
		return
	}

	referral, err := s.affUC.TrackReferral(c.Request.Context(), uint(affiliateID), sessionID, referrerURL, ipAddress, uint(customerID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, referral)
}

func (s *AffiliateService) ConvertReferral(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	if err := s.affUC.ConvertReferral(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "referral converted"})
}

func (s *AffiliateService) GetAffiliateReferrals(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	referrals, err := s.affUC.GetAffiliateReferrals(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, referrals)
}

// ========== 鑱旂洘璁㈠崟鎺ュ彛 ==========

func (s *AffiliateService) CreateAffiliateOrder(c *gin.Context) {
	var req models.AffiliateOrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := s.affUC.CreateAffiliateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (s *AffiliateService) GetAffiliateOrders(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	orders, err := s.affUC.GetAffiliateOrders(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// ========== 浣ｉ噾鎺ュ彛 ==========

func (s *AffiliateService) CalculateCommission(c *gin.Context) {
	var req models.CommissionCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commission, err := s.affUC.CalculateCommission(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if commission == nil {
		c.JSON(http.StatusOK, gin.H{"message": "commission amount is below minimum threshold"})
		return
	}

	c.JSON(http.StatusCreated, commission)
}

func (s *AffiliateService) GetAffiliateCommissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	commissions, err := s.affUC.GetAffiliateCommissions(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, commissions)
}

func (s *AffiliateService) GetPendingBalance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	balance, err := s.affUC.GetPendingBalance(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pending_balance": balance})
}

// ========== 鏀粯鎺ュ彛 ==========

func (s *AffiliateService) CreatePayout(c *gin.Context) {
	var req models.PayoutCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payout, err := s.affUC.CreatePayout(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payout)
}

func (s *AffiliateService) ProcessPayout(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.affUC.ProcessPayout(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "payout processed"})
}

func (s *AffiliateService) GetPayout(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	payout, err := s.affUC.GetPayout(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payout)
}

func (s *AffiliateService) GetAffiliatePayouts(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	payouts, err := s.affUC.GetAffiliatePayouts(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payouts)
}

// ========== 缁熻鎺ュ彛 ==========

func (s *AffiliateService) GetAffiliateStats(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	stats, err := s.affUC.GetAffiliateStats(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
