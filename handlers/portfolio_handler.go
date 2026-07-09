package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend/config"
	"backend/models"
	"backend/utils"
)

// GetPortfolios mengembalikan seluruh item portofolio milik user yang login.
func GetPortfolios(c *gin.Context) {
	userID := c.GetUint("userID")

	var items []models.Portfolio
	config.DB.Where("user_id = ?", userID).Order("id desc").Find(&items)
	if items == nil {
		items = []models.Portfolio{}
	}

	utils.SuccessResponse(c, 200, gin.H{"portfolios": items})
}

// handlePortfolioImage memproses file gambar opsional dari form multipart.
// Mengembalikan URL gambar (jika ada), penanda apakah ada file diupload, dan
// pesan error (kosong bila tidak ada masalah).
func handlePortfolioImage(c *gin.Context, userID uint) (imageURL *string, uploaded bool, errMsg string) {
	file, err := c.FormFile("image")
	if err != nil {
		return nil, false, "" // tidak ada gambar diupload
	}

	if file.Size > 5*1024*1024 {
		return nil, false, "Ukuran gambar maksimal 5MB"
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		return nil, false, "Format gambar harus jpg, jpeg, png, atau webp"
	}

	uploadDir := os.Getenv("PORTFOLIO_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads/portfolios"
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, false, "Gagal menyiapkan folder upload"
	}

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		return nil, false, "Gagal menyimpan gambar"
	}

	url := "/uploads/portfolios/" + filename
	return &url, true, ""
}

// CreatePortfolio menambah satu item portofolio (form multipart agar bisa
// sekaligus mengunggah gambar untuk jenis projek).
func CreatePortfolio(c *gin.Context) {
	userID := c.GetUint("userID")

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		utils.ErrorResponse(c, 400, "Judul wajib diisi")
		return
	}

	imageURL, _, errMsg := handlePortfolioImage(c, userID)
	if errMsg != "" {
		utils.ErrorResponse(c, 400, errMsg)
		return
	}

	item := models.Portfolio{
		UserId:      userID,
		Category:    strings.TrimSpace(c.PostForm("category")),
		Title:       title,
		Description: strings.TrimSpace(c.PostForm("description")),
		Link:        strings.TrimSpace(c.PostForm("link")),
		Issuer:      strings.TrimSpace(c.PostForm("issuer")),
		IssuedDate:  strings.TrimSpace(c.PostForm("issued_date")),
		TechStack:   strings.TrimSpace(c.PostForm("tech_stack")),
		ImageURL:    imageURL,
	}

	if err := config.DB.Create(&item).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan portofolio")
		return
	}

	utils.SuccessResponse(c, 201, item)
}

// UpdatePortfolio memperbarui satu item portofolio milik user. Gambar hanya
// diganti bila ada file baru; bila form "remove_image" = "true" gambar dihapus.
func UpdatePortfolio(c *gin.Context) {
	userID := c.GetUint("userID")
	id := c.Param("id")

	var item models.Portfolio
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		utils.ErrorResponse(c, 404, "Portofolio tidak ditemukan")
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		utils.ErrorResponse(c, 400, "Judul wajib diisi")
		return
	}

	item.Category = strings.TrimSpace(c.PostForm("category"))
	item.Title = title
	item.Description = strings.TrimSpace(c.PostForm("description"))
	item.Link = strings.TrimSpace(c.PostForm("link"))
	item.Issuer = strings.TrimSpace(c.PostForm("issuer"))
	item.IssuedDate = strings.TrimSpace(c.PostForm("issued_date"))
	item.TechStack = strings.TrimSpace(c.PostForm("tech_stack"))

	imageURL, uploaded, errMsg := handlePortfolioImage(c, userID)
	if errMsg != "" {
		utils.ErrorResponse(c, 400, errMsg)
		return
	}
	if uploaded {
		item.ImageURL = imageURL
	} else if c.PostForm("remove_image") == "true" {
		item.ImageURL = nil
	}

	if err := config.DB.Save(&item).Error; err != nil {
		utils.ErrorResponse(c, 500, "Gagal memperbarui portofolio")
		return
	}

	utils.SuccessResponse(c, 200, item)
}

// DeletePortfolio menghapus satu item portofolio milik user.
func DeletePortfolio(c *gin.Context) {
	userID := c.GetUint("userID")
	id := c.Param("id")

	res := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Portfolio{})
	if res.Error != nil {
		utils.ErrorResponse(c, 500, "Gagal menghapus portofolio")
		return
	}
	if res.RowsAffected == 0 {
		utils.ErrorResponse(c, 404, "Portofolio tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"message": "Portofolio dihapus"})
}
