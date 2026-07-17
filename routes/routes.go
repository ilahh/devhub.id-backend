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

	// Endpoint publik (tanpa login): daftar member berportofolio + berblog,
	// profil publik, dan halaman baca blog.
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

		// Profil Profesional: skill, riwayat pekerjaan, tempat tugas, mata pelajaran.
		protected.GET("/professional-profile", handlers.GetProfessionalProfile)
		protected.PUT("/professional-profile", handlers.UpdateProfessionalProfile)

		// Portofolio: kumpulan karya/pencapaian milik user (CRUD + upload gambar).
		protected.GET("/portfolios", handlers.GetPortfolios)
		protected.POST("/portfolios", handlers.CreatePortfolio)
		protected.PUT("/portfolios/:id", handlers.UpdatePortfolio)
		protected.DELETE("/portfolios/:id", handlers.DeletePortfolio)

		// Blog pribadi: CRUD artikel berbasis blok (teks/gambar/video/audio/3D/embed)
		// plus endpoint upload media.
		protected.GET("/blog/posts", handlers.GetMyPosts)
		protected.POST("/blog/posts", handlers.CreatePost)
		protected.GET("/blog/posts/:id", handlers.GetMyPost)
		protected.PUT("/blog/posts/:id", handlers.UpdatePost)
		protected.DELETE("/blog/posts/:id", handlers.DeletePost)
		protected.POST("/blog/media", handlers.UploadBlogMedia)
	}

	// Khusus admin: kelola seluruh user (ubah role, suspend, hapus)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(models.RoleAdmin))
	{
		adminGroup.GET("/users", handlers.ListUsers)
		adminGroup.PUT("/users/:id/role", handlers.UpdateUserRole)
		adminGroup.PATCH("/users/:id/status", handlers.ToggleUserStatus)
		adminGroup.DELETE("/users/:id", handlers.DeleteUser)
	}

	// Khusus moderator (admin juga boleh akses): moderasi user dengan role member
	moderatorGroup := api.Group("/moderator")
	moderatorGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(models.RoleAdmin, models.RoleModerator))
	{
		moderatorGroup.GET("/members", handlers.ListMembers)
		moderatorGroup.PATCH("/members/:id/status", handlers.ToggleMemberStatus)
	}
}
