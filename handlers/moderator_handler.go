package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"backend/config"
	"backend/models"
	"backend/utils"
)

func ListMembers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Where("role = ?", models.RoleMember).Order("id asc").Find(&users).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal mengambil data member")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"users": users})
}

func ToggleMemberStatus(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "ID user tidak valid")
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

	if user.Role != models.RoleMember {
		utils.ErrorResponse(c, 403, "Moderator hanya bisa memoderasi akun member")
		return
	}

	user.IsActive = input.IsActive
	if err := config.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan status")
		return
	}

	utils.SuccessResponse(c, 200, user)
}
