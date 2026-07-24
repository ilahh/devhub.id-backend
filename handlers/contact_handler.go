package handlers

import (
	"backend/config"
	"backend/models"
	"backend/utils"
	"github.com/gin-gonic/gin"
)

type ContactResponse struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Github   string `json:"github"`
	Linkedin string `json:"linkedin"`
	Website  string `json:"website"`
}

type UpdateContactInput struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Github   string `json:"github"`
	Linkedin string `json:"linkedin"`
	Website  string `json:"website"`
}

func GetContact(c *gin.Context) {
	userID := c.GetUint("userID")

	var contact models.Contact
	config.DB.Where("user_id = ?", userID).First(&contact)

	utils.SuccessResponse(c, 200, ContactResponse{
		Phone:    contact.Phone,
		Email:    contact.Email,
		Address:  contact.Address,
		Github:   contact.Github,
		Linkedin: contact.Linkedin,
		Website:  contact.Website,
	})
}

func UpdateContact(c *gin.Context) {
	userID := c.GetUint("userID")

	var input UpdateContactInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Data tidak valid")
		return
	}

	var contact models.Contact
	if err := config.DB.Where("user_id = ?", userID).First(&contact).Error; err != nil {
		contact = models.Contact{UserId: userID}
	}

	contact.Phone = input.Phone
	contact.Email = input.Email
	contact.Address = input.Address
	contact.Github = input.Github
	contact.Linkedin = input.Linkedin
	contact.Website = input.Website

	if err := config.DB.Save(&contact).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan kontak")
		return
	}

	GetContact(c)
}
