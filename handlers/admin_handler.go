package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"backend/config"
	"backend/models"
	"backend/utils"
)

// ListUsers mengembalikan seluruh user beserta role dan statusnya. Khusus admin.
func ListUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Order("id asc").Find(&users).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal mengambil data user")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"users": users})
}

type UpdateRoleInput struct {
	Role models.Role `json:"role" binding:"required"`
}

// UpdateUserRole mengubah role seorang user. Admin tidak bisa mengubah role dirinya sendiri
// supaya tidak ada kondisi tanpa admin yang tersisa.
func UpdateUserRole(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "ID user tidak valid")
		return
	}

	requesterID := c.GetUint("userID")
	if uint(targetID) == requesterID {
		utils.ErrorResponse(c, 400, "Tidak bisa mengubah role akun sendiri")
		return
	}

	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil || !input.Role.IsValid() {
		utils.ErrorResponse(c, 400, "Role tidak valid. Gunakan admin, moderator, atau member")
		return
	}

	var user models.User
	if err := config.DB.First(&user, targetID).Error; err != nil {
		utils.ErrorResponse(c, 404, "User tidak ditemukan")
		return
	}

	user.Role = input.Role
	if err := config.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan role")
		return
	}

	utils.SuccessResponse(c, 200, user)
}

type UpdateStatusInput struct {
	IsActive bool `json:"is_active"`
}

// ToggleUserStatus mengaktifkan/menonaktifkan akun user mana pun. Admin tidak bisa
// menonaktifkan akunnya sendiri.
func ToggleUserStatus(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "ID user tidak valid")
		return
	}

	requesterID := c.GetUint("userID")
	if uint(targetID) == requesterID {
		utils.ErrorResponse(c, 400, "Tidak bisa menonaktifkan akun sendiri")
		return
	}

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Status tidak valid")
		return
	}

	var user models.User
	if err := config.DB.First(&user, targetID).Error; err != nil {
		utils.ErrorResponse(c, 404, "User tidak ditemukan")
		return
	}

	user.IsActive = input.IsActive
	if err := config.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan status")
		return
	}

	utils.SuccessResponse(c, 200, user)
}

// DeleteUser menghapus akun user. Admin tidak bisa menghapus akunnya sendiri.
func DeleteUser(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "ID user tidak valid")
		return
	}

	requesterID := c.GetUint("userID")
	if uint(targetID) == requesterID {
		utils.ErrorResponse(c, 400, "Tidak bisa menghapus akun sendiri")
		return
	}

	if err := config.DB.Delete(&models.User{}, targetID).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menghapus user")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"message": "User berhasil dihapus"})
}
