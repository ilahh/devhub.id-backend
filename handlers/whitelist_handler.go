package handlers

import (
	"net/mail"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"backend/config"
	"backend/models"
	"backend/utils"
)

// ListWhitelists mengembalikan seluruh entri whitelist email. Khusus admin.
func ListWhitelists(c *gin.Context) {
	var items []models.RoleEmailWhitelist
	if err := config.DB.Order("id asc").Find(&items).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal mengambil data whitelist")
		return
	}
	utils.SuccessResponse(c, 200, gin.H{"whitelists": items})
}

type CreateWhitelistInput struct {
	Role  models.Role `json:"role" binding:"required"`
	Email string      `json:"email" binding:"required"`
}

// CreateWhitelist menambahkan email ke whitelist untuk role admin/moderator.
func CreateWhitelist(c *gin.Context) {
	var input CreateWhitelistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Role dan email wajib diisi")
		return
	}

	// Whitelist hanya relevan untuk role admin & moderator.
	if input.Role != models.RoleAdmin && input.Role != models.RoleModerator {
		utils.ErrorResponse(c, 400, "Whitelist hanya untuk role admin atau moderator")
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		utils.ErrorResponse(c, 400, "Format email tidak valid")
		return
	}

	var existing models.RoleEmailWhitelist
	if err := config.DB.Where("role = ? AND email = ?", input.Role, email).First(&existing).Error; err == nil {
		utils.ErrorResponse(c, 409, "Email sudah ada di whitelist role ini")
		return
	}

	item := models.RoleEmailWhitelist{Role: input.Role, Email: email, IsActive: true}
	if err := config.DB.Create(&item).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menambahkan whitelist")
		return
	}

	utils.SuccessResponse(c, 201, item)
}

type ToggleWhitelistInput struct {
	IsActive bool `json:"is_active"`
}

// ToggleWhitelistStatus mengaktifkan/menonaktifkan sebuah entri whitelist.
func ToggleWhitelistStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "ID tidak valid")
		return
	}

	var input ToggleWhitelistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Status tidak valid")
		return
	}

	var item models.RoleEmailWhitelist
	if err := config.DB.First(&item, id).Error; err != nil {
		utils.ErrorResponse(c, 404, "Whitelist tidak ditemukan")
		return
	}

	item.IsActive = input.IsActive
	if err := config.DB.Save(&item).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan status")
		return
	}

	utils.SuccessResponse(c, 200, item)
}

// DeleteWhitelist menghapus sebuah entri whitelist.
func DeleteWhitelist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "ID tidak valid")
		return
	}

	if err := config.DB.Delete(&models.RoleEmailWhitelist{}, id).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menghapus whitelist")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"message": "Whitelist berhasil dihapus"})
}
