package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"

	"backend/models"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// IsEmailWhitelisted mengecek apakah email diizinkan (aktif) untuk role tertentu.
// Pengecekan dilakukan di backend (bukan hanya frontend) demi keamanan.
func IsEmailWhitelisted(role models.Role, email string) bool {
	email = normalizeEmail(email)
	if email == "" {
		return false
	}
	var count int64
	DB.Model(&models.RoleEmailWhitelist{}).
		Where("role = ? AND email = ? AND is_active = ?", role, email, true).
		Count(&count)
	return count > 0
}

// EnsureWhitelisted menambahkan email ke whitelist role secara idempotent dan
// memastikannya aktif. Dipakai saat seeding, saat user pertama menjadi admin,
// dan saat admin menetapkan role admin/moderator dari dashboard.
func EnsureWhitelisted(role models.Role, email string) error {
	email = normalizeEmail(email)
	if email == "" {
		return nil
	}

	var entry models.RoleEmailWhitelist
	err := DB.Where("role = ? AND email = ?", role, email).First(&entry).Error
	if err == nil {
		if !entry.IsActive {
			entry.IsActive = true
			return DB.Save(&entry).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	entry = models.RoleEmailWhitelist{Role: role, Email: email, IsActive: true}
	return DB.Create(&entry).Error
}

// SeedWhitelistFromEnv membaca ADMIN_EMAILS & MODERATOR_EMAILS (dipisah koma)
// lalu memasukkannya ke tabel whitelist sebagai bootstrap awal (Opsi B).
func SeedWhitelistFromEnv() {
	seed := func(role models.Role, envKey string) {
		raw := os.Getenv(envKey)
		for _, e := range strings.Split(raw, ",") {
			e = normalizeEmail(e)
			if e == "" {
				continue
			}
			if err := EnsureWhitelisted(role, e); err != nil {
				log.Printf("gagal seed whitelist %s (%s): %v", e, role, err)
			}
		}
	}

	seed(models.RoleAdmin, "ADMIN_EMAILS")
	seed(models.RoleModerator, "MODERATOR_EMAILS")
}

// AutoWhitelistExistingStaff memastikan email seluruh user yang sudah berrole
// admin/moderator ikut masuk whitelist supaya mereka tidak terkunci setelah
// aturan whitelist diberlakukan.
func AutoWhitelistExistingStaff() {
	var users []models.User
	if err := DB.Where("role IN ?", []models.Role{models.RoleAdmin, models.RoleModerator}).
		Find(&users).Error; err != nil {
		log.Printf("gagal memuat staff untuk auto-whitelist: %v", err)
		return
	}
	for _, u := range users {
		if err := EnsureWhitelisted(u.Role, u.Email); err != nil {
			log.Printf("gagal auto-whitelist %s: %v", u.Email, err)
		}
	}
}
