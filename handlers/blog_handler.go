package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/config"
	"backend/models"
	"backend/utils"
)

var validBlockTypes = map[string]bool{
	"text": true, "image": true, "video": true, "audio": true, "model3d": true, "embed": true,
}

type blockInput struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	MediaURL string `json:"media_url"`
	Caption  string `json:"caption"`
}

type postInput struct {
	Title    string       `json:"title"`
	Excerpt  string       `json:"excerpt"`
	CoverURL string       `json:"cover_url"`
	Status   string       `json:"status"`
	Blocks   []blockInput `json:"blocks"`
}

// normalizeStatus memastikan status hanya bernilai "draft" atau "published".
func normalizeStatus(s string) string {
	if s == "published" {
		return "published"
	}
	return "draft"
}

// uniqueSlug menghasilkan slug yang unik untuk seorang user. Bila slug dasar
// sudah dipakai, ditambahkan sufiks angka (-2, -3, ...). excludeID dipakai saat
// update agar post yang sedang diedit tidak dianggap bentrok dengan dirinya sendiri.
func uniqueSlug(userID uint, base string, excludeID uint) string {
	if base == "" {
		base = "post"
	}
	slug := base
	for i := 2; ; i++ {
		var count int64
		q := config.DB.Model(&models.BlogPost{}).Where("user_id = ? AND slug = ?", userID, slug)
		if excludeID != 0 {
			q = q.Where("id <> ?", excludeID)
		}
		q.Count(&count)
		if count == 0 {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// buildBlocks memvalidasi & menyusun blok dari payload menjadi model siap simpan,
// sekaligus menetapkan urutan (Position) dan membuang blok kosong.
func buildBlocks(postID, userID uint, inputs []blockInput) []models.BlogBlock {
	blocks := make([]models.BlogBlock, 0, len(inputs))
	pos := 0
	for _, b := range inputs {
		t := strings.TrimSpace(b.Type)
		if !validBlockTypes[t] {
			continue
		}
		// Buang blok yang tidak punya isi berarti.
		if (t == "text" || t == "embed") && strings.TrimSpace(b.Text) == "" {
			continue
		}
		if (t == "image" || t == "video" || t == "audio" || t == "model3d") && strings.TrimSpace(b.MediaURL) == "" {
			continue
		}
		blocks = append(blocks, models.BlogBlock{
			PostId:   postID,
			Type:     t,
			Text:     b.Text,
			MediaURL: strings.TrimSpace(b.MediaURL),
			Caption:  strings.TrimSpace(b.Caption),
			Position: pos,
		})
		pos++
	}
	return blocks
}

func orderedBlocks(db *gorm.DB) *gorm.DB {
	return db.Order("position asc")
}

// GetMyPosts mengembalikan seluruh blog milik user yang login (draft & publish).
func GetMyPosts(c *gin.Context) {
	userID := c.GetUint("userID")

	var posts []models.BlogPost
	config.DB.Where("user_id = ?", userID).
		Preload("Blocks", orderedBlocks).
		Order("updated_at desc").Find(&posts)
	if posts == nil {
		posts = []models.BlogPost{}
	}

	utils.SuccessResponse(c, 200, gin.H{"posts": posts})
}

// GetMyPost mengembalikan satu blog milik user yang login (untuk halaman editor).
func GetMyPost(c *gin.Context) {
	userID := c.GetUint("userID")
	id := c.Param("id")

	var post models.BlogPost
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("Blocks", orderedBlocks).First(&post).Error; err != nil {
		utils.ErrorResponse(c, 404, "Blog tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, 200, post)
}

// CreatePost membuat blog baru beserta blok-bloknya, dibungkus transaksi.
func CreatePost(c *gin.Context) {
	userID := c.GetUint("userID")

	var input postInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Data tidak valid")
		return
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		utils.ErrorResponse(c, 400, "Judul wajib diisi")
		return
	}

	status := normalizeStatus(input.Status)
	post := models.BlogPost{
		UserId:  userID,
		Slug:    uniqueSlug(userID, utils.Slugify(title), 0),
		Title:   title,
		Excerpt: strings.TrimSpace(input.Excerpt),
		Status:  status,
	}
	if cover := strings.TrimSpace(input.CoverURL); cover != "" {
		post.CoverURL = &cover
	}
	if status == "published" {
		now := time.Now()
		post.PublishedAt = &now
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		blocks := buildBlocks(post.Id, userID, input.Blocks)
		if len(blocks) > 0 {
			if err := tx.Create(&blocks).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.ErrorResponse(c, 500, "Gagal menyimpan blog")
		return
	}

	config.DB.Preload("Blocks", orderedBlocks).First(&post, post.Id)
	utils.SuccessResponse(c, 201, post)
}

// UpdatePost memperbarui blog milik user. Blok lama dihapus lalu ditulis ulang
// dari payload (pola replace) di dalam transaksi agar konsisten.
func UpdatePost(c *gin.Context) {
	userID := c.GetUint("userID")
	id := c.Param("id")

	var post models.BlogPost
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&post).Error; err != nil {
		utils.ErrorResponse(c, 404, "Blog tidak ditemukan")
		return
	}

	var input postInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, 400, "Data tidak valid")
		return
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		utils.ErrorResponse(c, 400, "Judul wajib diisi")
		return
	}

	status := normalizeStatus(input.Status)

	post.Title = title
	post.Excerpt = strings.TrimSpace(input.Excerpt)
	if cover := strings.TrimSpace(input.CoverURL); cover != "" {
		post.CoverURL = &cover
	} else {
		post.CoverURL = nil
	}

	// Set published_at saat pertama kali dipublikasikan; kosongkan bila kembali draft.
	if status == "published" && post.Status != "published" {
		now := time.Now()
		post.PublishedAt = &now
	} else if status != "published" {
		post.PublishedAt = nil
	}
	post.Status = status

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&post).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", post.Id).Delete(&models.BlogBlock{}).Error; err != nil {
			return err
		}
		blocks := buildBlocks(post.Id, userID, input.Blocks)
		if len(blocks) > 0 {
			if err := tx.Create(&blocks).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.ErrorResponse(c, 500, "Gagal memperbarui blog")
		return
	}

	config.DB.Preload("Blocks", orderedBlocks).First(&post, post.Id)
	utils.SuccessResponse(c, 200, post)
}

// DeletePost menghapus blog milik user beserta seluruh bloknya.
func DeletePost(c *gin.Context) {
	userID := c.GetUint("userID")
	id := c.Param("id")

	var post models.BlogPost
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&post).Error; err != nil {
		utils.ErrorResponse(c, 404, "Blog tidak ditemukan")
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.Id).Delete(&models.BlogBlock{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.BlogPost{}, post.Id).Error
	})
	if err != nil {
		utils.ErrorResponse(c, 500, "Gagal menghapus blog")
		return
	}

	utils.SuccessResponse(c, 200, gin.H{"message": "Blog dihapus"})
}

// blogMediaRules menentukan batas ukuran & format per jenis media blog.
var blogMediaRules = map[string]struct {
	MaxBytes int64
	Exts     map[string]bool
}{
	"image":   {5 * 1024 * 1024, map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}},
	"audio":   {20 * 1024 * 1024, map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".m4a": true}},
	"video":   {100 * 1024 * 1024, map[string]bool{".mp4": true, ".webm": true, ".mov": true}},
	"model3d": {30 * 1024 * 1024, map[string]bool{".glb": true, ".gltf": true}},
}

// UploadBlogMedia menerima satu file media (image/audio/video/model3d) dan
// mengembalikan URL-nya. Frontend memakai URL ini sebagai isi blok.
func UploadBlogMedia(c *gin.Context) {
	userID := c.GetUint("userID")

	mediaType := c.PostForm("type")
	rule, ok := blogMediaRules[mediaType]
	if !ok {
		utils.ErrorResponse(c, 400, "Tipe media tidak valid (image, audio, video, model3d)")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, 400, "File tidak ditemukan")
		return
	}

	if file.Size > rule.MaxBytes {
		utils.ErrorResponse(c, 400, fmt.Sprintf("Ukuran file maksimal %dMB", rule.MaxBytes/(1024*1024)))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !rule.Exts[ext] {
		utils.ErrorResponse(c, 400, "Format file tidak didukung untuk tipe "+mediaType)
		return
	}

	uploadDir := os.Getenv("BLOG_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads/blog"
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

	url := "/uploads/blog/" + filename
	utils.SuccessResponse(c, 200, gin.H{"url": url, "type": mediaType})
}
