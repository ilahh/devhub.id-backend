package routes

import (
	"github.com/gin-gonic/gin"

	"backend/handlers"
	"backend/middleware"
	"backend/models"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	authGroup := api.Group("/auth")
	authGroup.POST("/register", handlers.Register)
	authGroup.POST("/login", handlers.Login)

	userGroup := api.Group("/user")
	userGroup.GET("/check-username", handlers.CheckUsername)

	protected := userGroup.Group("")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/profile", handlers.GetProfile)
		protected.PUT("/profile", handlers.UpdateProfile)
		protected.POST("/avatar", handlers.UploadAvatar)
	}

	// Khusus admin: kelola seluruh user (ubah role, suspend, hapus)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(models.RoleAdmin))
	{
		adminGroup.GET("/users", handlers.ListUsers)
		adminGroup.PUT("/users/:id/role", handlers.UpdateUserRole)
		adminGroup.PATCH("/users/:id/status", handlers.ToggleUserStatus)
		adminGroup.DELETE("/users/:id", handlers.DeleteUser)

		// Kelola whitelist email admin/moderator dari dashboard admin.
		adminGroup.GET("/whitelists", handlers.ListWhitelists)
		adminGroup.POST("/whitelists", handlers.CreateWhitelist)
		adminGroup.PATCH("/whitelists/:id/status", handlers.ToggleWhitelistStatus)
		adminGroup.DELETE("/whitelists/:id", handlers.DeleteWhitelist)
	}

	// Khusus moderator (admin juga boleh akses): moderasi user dengan role member
	moderatorGroup := api.Group("/moderator")
	moderatorGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(models.RoleAdmin, models.RoleModerator))
	{
		moderatorGroup.GET("/members", handlers.ListMembers)
		moderatorGroup.PATCH("/members/:id/status", handlers.ToggleMemberStatus)
	}
}
