package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend/config"
	"backend/models"
	"backend/utils"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._]{4,30}$`)

const usernameFormatError = "Username harus 4-30 karakter: huruf, angka, titik (.), atau underscore (_)"

func GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, 404, "User tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, 200, user)
}

type UpdateProfileInput struct {
	Username string `json:"username" binding:"required"`
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Username wajib diisi")
		return
	}

	if !usernameRegex.MatchString(input.Username) {
		utils.ErrorResponse(c, 400, usernameFormatError)
		return
	}

	// Huruf besar disimpan sebagai huruf kecil
	username := strings.ToLower(input.Username)

	var existing models.User
	if err := config.DB.Where("username = ? AND id != ?", username, userID).First(&existing).Error; err == nil {
		utils.ErrorResponse(c, 409, "Username sudah dipakai")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, 404, "User tidak ditemukan")
		return
	}

	user.Username = &username
	if err := config.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan profil")
		return
	}

	utils.SuccessResponse(c, 200, user)
}

func CheckUsername(c *gin.Context) {
	username := c.Query("username")

	if !usernameRegex.MatchString(username) {
		utils.ErrorResponse(c, 400, usernameFormatError)
		return
	}

	// Huruf besar disimpan sebagai huruf kecil, jadi pengecekan juga pakai versi lowercase
	username = strings.ToLower(username)

	var existing models.User
	if err := config.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		utils.ErrorResponse(c, 409, "Username sudah dipakai")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"available": true})
}

func UploadAvatar(c *gin.Context) {
	userID := c.GetUint("userID")

	file, err := c.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(c, 400, "File avatar tidak ditemukan")
		return
	}

	if file.Size > 5*1024*1024 {
		utils.ErrorResponse(c, 400, "Ukuran file maksimal 5MB")
		return
	}

	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		utils.ErrorResponse(c, 400, "Format file harus jpg, jpeg, png, atau webp")
		return
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads/avatars"
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyiapkan folder upload")
		return
	}

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan file")
		return
	}

	avatarURL := "/uploads/avatars/" + filename

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, 404, "User tidak ditemukan")
		return
	}
	user.AvatarURL = &avatarURL
	if err := config.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan avatar")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"avatar_url": avatarURL})
}
