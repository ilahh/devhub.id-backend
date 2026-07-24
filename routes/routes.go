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

	publicGroup := api.Group("/public")
	{
		publicGroup.GET("/members", handlers.ListPublicMembers)
		publicGroup.GET("/members/:username", handlers.GetPublicProfile)
		publicGroup.GET("/members/:username/posts/:slug", handlers.GetPublicPost)
		publicGroup.GET("/posts", handlers.ListPublicPosts)
	}

	userGroup := api.Group("/user")
	userGroup.GET("/check-username", handlers.CheckUsername)

	protected := userGroup.Group("")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/profile", handlers.GetProfile)
		protected.PUT("/profile", handlers.UpdateProfile)
		protected.POST("/avatar", handlers.UploadAvatar)

		protected.GET("/professional-profile", handlers.GetProfessionalProfile)
		protected.PUT("/professional-profile", handlers.UpdateProfessionalProfile)

		protected.GET("/contact", handlers.GetContact)
		protected.PUT("/contact", handlers.UpdateContact)

		protected.GET("/portfolios", handlers.GetPortfolios)
		protected.POST("/portfolios", handlers.CreatePortfolio)
		protected.PUT("/portfolios/:id", handlers.UpdatePortfolio)
		protected.DELETE("/portfolios/:id", handlers.DeletePortfolio)

		protected.GET("/blog/posts", handlers.GetMyPosts)
		protected.POST("/blog/posts", handlers.CreatePost)
		protected.GET("/blog/posts/:id", handlers.GetMyPost)
		protected.PUT("/blog/posts/:id", handlers.UpdatePost)
		protected.DELETE("/blog/posts/:id", handlers.DeletePost)
		protected.POST("/blog/media", handlers.UploadBlogMedia)
	}

	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(models.RoleAdmin))
	{
		adminGroup.GET("/users", handlers.ListUsers)
		adminGroup.PUT("/users/:id/role", handlers.UpdateUserRole)
		adminGroup.PATCH("/users/:id/status", handlers.ToggleUserStatus)
		adminGroup.DELETE("/users/:id", handlers.DeleteUser)
	}

	moderatorGroup := api.Group("/moderator")
	moderatorGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(models.RoleAdmin, models.RoleModerator))
	{
		moderatorGroup.GET("/members", handlers.ListMembers)
		moderatorGroup.PATCH("/members/:id/status", handlers.ToggleMemberStatus)
	}
}
