// Package service 管理后台服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/auth/jwt"
	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/models"
)

type AdminService struct {
	userUC    *biz.AdminUserUseCase
	roleUC    *biz.AdminRoleUseCase
	settingUC *biz.SettingUseCase
	logUC     *biz.ActivityLogUseCase
	jwtSvc    contract.JWTService
}

// NewAdminService 创建管理后台服务。
//
// 中文说明：
// - 使用 framework 级 JWTService，替代项目层 jwtSecret；
// - 中间件改用 framework 提供的 AuthMiddleware。
func NewAdminService(userUC *biz.AdminUserUseCase, roleUC *biz.AdminRoleUseCase, settingUC *biz.SettingUseCase, logUC *biz.ActivityLogUseCase, jwtSvc contract.JWTService) *AdminService {
	return &AdminService{userUC: userUC, roleUC: roleUC, settingUC: settingUC, logUC: logUC, jwtSvc: jwtSvc}
}

func (s *AdminService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/admin")
	// 使用 framework JWT middleware
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		api.POST("/login", s.Login)

		protected := api.Group("")
		protected.Use(adminAuth)
		protected.POST("/users", s.CreateAdmin)
		protected.GET("/users", s.ListAdmins)
		protected.GET("/users/:id", s.GetAdmin)
		protected.PUT("/users/:id", s.UpdateAdmin)
		protected.DELETE("/users/:id", s.DeleteAdmin)
		protected.POST("/users/:id/password", s.ChangePassword)

		protected.GET("/roles", s.ListRoles)
		protected.POST("/roles", s.CreateRole)
		protected.GET("/roles/:id", s.GetRole)
		protected.PUT("/roles/:id", s.UpdateRole)
		protected.DELETE("/roles/:id", s.DeleteRole)
		protected.GET("/permissions", s.ListPermissions)

		protected.GET("/settings", s.GetAllSettings)
		protected.GET("/settings/:name", s.GetSetting)
		protected.PUT("/settings", s.SetSetting)
		protected.DELETE("/settings/:name", s.DeleteSetting)

		protected.GET("/logs", s.ListLogs)
	}
}

func (s *AdminService) Login(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := s.userUC.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.AdminLoginResponse{
		Token: token,
		User:  models.ToAdminUserResponse(user),
	})
}

func (s *AdminService) CreateAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.AdminUser{
		Username: req.Username,
		Email:    req.Email,
	}

	if err := s.userUC.CreateAdmin(c.Request.Context(), user, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.ToAdminUserResponse(user))
}

func (s *AdminService) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	users, total, err := s.userUC.ListAdmins(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.AdminUserResponse, len(users))
	for i, u := range users {
		items[i] = models.ToAdminUserResponse(u)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (s *AdminService) GetAdmin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := s.userUC.GetAdmin(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToAdminUserResponse(user))
}

func (s *AdminService) UpdateAdmin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var user models.AdminUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.ID = id
	if err := s.userUC.UpdateAdmin(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToAdminUserResponse(&user))
}

func (s *AdminService) DeleteAdmin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.userUC.DeleteAdmin(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *AdminService) ChangePassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.userUC.ChangePassword(c.Request.Context(), id, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

func (s *AdminService) ListRoles(c *gin.Context) {
	roles, err := s.roleUC.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (s *AdminService) CreateRole(c *gin.Context) {
	var role models.AdminRole
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.roleUC.CreateRole(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (s *AdminService) GetRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, err := s.roleUC.GetRole(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (s *AdminService) UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var role models.AdminRole
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role.ID = id
	if err := s.roleUC.UpdateRole(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (s *AdminService) DeleteRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.roleUC.DeleteRole(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *AdminService) ListPermissions(c *gin.Context) {
	perms, err := s.roleUC.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

func (s *AdminService) GetAllSettings(c *gin.Context) {
	settings, err := s.settingUC.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (s *AdminService) GetSetting(c *gin.Context) {
	name := c.Param("name")
	value, err := s.settingUC.GetSetting(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": name, "value": value})
}

func (s *AdminService) SetSetting(c *gin.Context) {
	var req models.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.settingUC.SetSetting(c.Request.Context(), req.Name, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "setting updated"})
}

func (s *AdminService) DeleteSetting(c *gin.Context) {
	name := c.Param("name")
	if err := s.settingUC.DeleteSetting(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *AdminService) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	logs, total, err := s.logUC.ListLogs(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": logs, "total": total})
}