package handlers

import (
	"log"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/config"
	"backend/models"
	"backend/utils"
)

type RegisterInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Email dan password wajib diisi")
		return
	}

	if _, err := mail.ParseAddress(input.Email); err != nil {
		utils.ErrorResponse(c, 400, "Format email tidak valid")
		return
	}

	if len(input.Password) < 8 {
		utils.ErrorResponse(c, 400, "Password minimal 8 karakter")
		return
	}

	// Jangan mengisi username otomatis dari email.
	// Biarkan kosong (NULL) supaya user baru diarahkan ke halaman
	// "Lengkapi Profil" untuk memilih username mereka sendiri.
	var usernamePtr *string
	if strings.TrimSpace(input.Username) != "" {
		normalized := strings.ToLower(strings.TrimSpace(input.Username))
		if !usernameRegex.MatchString(normalized) {
			utils.ErrorResponse(c, 400, usernameFormatError)
			return
		}
		usernamePtr = &normalized
	}

	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		utils.ErrorResponse(c, 409, "Email sudah terdaftar")
		return
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Println("hash password error:", err)
		utils.ErrorResponse(c, 500, "Gagal memproses password")
		return
	}

	// User pertama yang mendaftar otomatis jadi admin.
	var count int64
	config.DB.Model(&models.User{}).Count(&count)

	role := models.RoleMember
	if count == 0 {
		role = models.RoleAdmin
	}

	user := models.User{Email: input.Email, Password: hashed, Username: usernamePtr, Role: role, IsActive: true}
	if err := config.DB.Create(&user).Error; err != nil {
		log.Println("create user error:", err)
		utils.ErrorResponse(c, 500, "Gagal membuat akun")
		return
	}

	// Register publik SELALU membuat akun member (role tidak pernah diambil dari input).
	// Satu-satunya pengecualian adalah user pertama yang otomatis menjadi admin; untuk
	// kasus itu, emailnya langsung dimasukkan ke whitelist agar tetap bisa login.
	if role == models.RoleAdmin || role == models.RoleModerator {
		_ = config.EnsureWhitelisted(role, user.Email)
	}

	utils.SuccessResponse(c, 201, gin.H{
		"message": "Registrasi berhasil",
		"user":    user,
	})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Email dan password wajib diisi")
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.ErrorResponse(c, 401, "Email atau password salah")
		return
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		utils.ErrorResponse(c, 401, "Email atau password salah")
		return
	}

	if !user.IsActive {
		utils.ErrorResponse(c, 403, "Akun kamu sedang dinonaktifkan")
		return
	}

	// Admin & moderator hanya boleh login jika email mereka ada di whitelist role terkait.
	// Pesan sengaja dibuat generik agar tidak membocorkan status whitelist ke publik.
	if user.Role == models.RoleAdmin || user.Role == models.RoleModerator {
		if !config.IsEmailWhitelisted(user.Role, user.Email) {
			utils.ErrorResponse(c, 403, "Email tidak memiliki akses untuk role ini")
			return
		}
	}

	token, err := utils.GenerateToken(user.Id, user.Role)
	if err != nil {
		utils.ErrorResponse(c, 500, "Gagal membuat token")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{
		"token": token,
		"user":  user,
	})
}
