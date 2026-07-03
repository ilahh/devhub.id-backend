package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend/config"
	"backend/models"
	"backend/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak menemukan file .env, pakai environment variable yang ada")
	}

	config.ConnectDatabase()
	config.DB.AutoMigrate(&models.User{}, &models.RoleEmailWhitelist{})

	config.DB.Model(&models.User{}).
		Where("id = (SELECT MIN(id) FROM users) AND role <> ?", models.RoleAdmin).
		Update("role", models.RoleAdmin)

	config.SeedWhitelistFromEnv()
	config.AutoWhitelistExistingStaff()

	r := gin.Default()

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	origins := strings.Split(allowedOrigins, ",")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Static("/uploads", "./uploads")

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server jalan di port %s", port)
	r.Run(":" + port)
}
