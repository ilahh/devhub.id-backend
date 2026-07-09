package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/config"
	"backend/models"
	"backend/utils"
)

type professionalProfileResponse struct {
	CurrentWorkplace string                  `json:"current_workplace"`
	CurrentPosition  string                  `json:"current_position"`
	Skills           []models.Skill          `json:"skills"`
	Experiences      []models.WorkExperience `json:"experiences"`
	Subjects         []models.Subject        `json:"subjects"`
}

// GetProfessionalProfile mengembalikan seluruh data profil profesional milik
// user yang sedang login (tempat tugas, skill, riwayat pekerjaan, mata pelajaran).
func GetProfessionalProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var profile models.ProfessionalProfile
	// Abaikan error "record not found" — user yang belum mengisi tetap dapat
	// respons dengan nilai kosong.
	config.DB.Where("user_id = ?", userID).First(&profile)

	var skills []models.Skill
	config.DB.Where("user_id = ?", userID).Order("id asc").Find(&skills)

	var experiences []models.WorkExperience
	config.DB.Where("user_id = ?", userID).Order("id desc").Find(&experiences)

	var subjects []models.Subject
	config.DB.Where("user_id = ?", userID).Order("id asc").Find(&subjects)

	if skills == nil {
		skills = []models.Skill{}
	}
	if experiences == nil {
		experiences = []models.WorkExperience{}
	}
	if subjects == nil {
		subjects = []models.Subject{}
	}

	utils.SuccessResponse(c, 200, professionalProfileResponse{
		CurrentWorkplace: profile.CurrentWorkplace,
		CurrentPosition:  profile.CurrentPosition,
		Skills:           skills,
		Experiences:      experiences,
		Subjects:         subjects,
	})
}

type skillInput struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type experienceInput struct {
	Position    string `json:"position"`
	Institution string `json:"institution"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}

type subjectInput struct {
	Name string `json:"name"`
}

type updateProfessionalProfileInput struct {
	CurrentWorkplace string            `json:"current_workplace"`
	CurrentPosition  string            `json:"current_position"`
	Skills           []skillInput      `json:"skills"`
	Experiences      []experienceInput `json:"experiences"`
	Subjects         []subjectInput    `json:"subjects"`
}

// UpdateProfessionalProfile menyimpan seluruh profil profesional sekaligus.
// Pendekatannya "replace": data skill/riwayat/mapel lama milik user dihapus lalu
// ditulis ulang dari payload, dibungkus transaksi agar konsisten.
func UpdateProfessionalProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var input updateProfessionalProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Data tidak valid")
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Upsert tempat tugas sekarang.
		var profile models.ProfessionalProfile
		if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			profile = models.ProfessionalProfile{UserId: userID}
		}
		profile.CurrentWorkplace = input.CurrentWorkplace
		profile.CurrentPosition = input.CurrentPosition
		if err := tx.Save(&profile).Error; err != nil {
			return err
		}

		// Ganti seluruh skill.
		if err := tx.Where("user_id = ?", userID).Delete(&models.Skill{}).Error; err != nil {
			return err
		}
		for _, s := range input.Skills {
			if s.Name == "" {
				continue
			}
			if err := tx.Create(&models.Skill{UserId: userID, Name: s.Name, Level: s.Level}).Error; err != nil {
				return err
			}
		}

		// Ganti seluruh riwayat pekerjaan.
		if err := tx.Where("user_id = ?", userID).Delete(&models.WorkExperience{}).Error; err != nil {
			return err
		}
		for _, e := range input.Experiences {
			if e.Position == "" && e.Institution == "" {
				continue
			}
			if err := tx.Create(&models.WorkExperience{
				UserId:      userID,
				Position:    e.Position,
				Institution: e.Institution,
				StartDate:   e.StartDate,
				EndDate:     e.EndDate,
				Description: e.Description,
			}).Error; err != nil {
				return err
			}
		}

		// Ganti seluruh mata pelajaran.
		if err := tx.Where("user_id = ?", userID).Delete(&models.Subject{}).Error; err != nil {
			return err
		}
		for _, sub := range input.Subjects {
			if sub.Name == "" {
				continue
			}
			if err := tx.Create(&models.Subject{UserId: userID, Name: sub.Name}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan profil profesional")
		return
	}

	// Balas dengan data terbaru.
	GetProfessionalProfile(c)
}
