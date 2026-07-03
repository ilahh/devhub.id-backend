package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"backend/models"
	"backend/utils"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.ErrorResponse(c, 401, "Token tidak ditemukan")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.ErrorResponse(c, 401, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleRequired(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, 403, "Akses ditolak")
			c.Abort()
			return
		}

		userRole, _ := roleVal.(models.Role)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, 403, "Kamu tidak punya akses ke fitur ini")
		c.Abort()
	}
}
