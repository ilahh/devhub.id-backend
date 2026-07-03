package models

import "time"

// RoleEmailWhitelist menyimpan daftar email yang diizinkan memiliki/masuk
// sebagai role admin atau moderator (Opsi A). Kombinasi role+email unik.
//
// Catatan: skema saat ini memakai Role sebagai string (bukan role_id/foreign
// key ke tabel roles) agar konsisten dengan model User yang sudah ada.
type RoleEmailWhitelist struct {
	Id        uint      `gorm:"primaryKey" json:"id"`
	Role      Role      `gorm:"type:varchar(20);not null;uniqueIndex:idx_role_email" json:"role"`
	Email     string    `gorm:"not null;uniqueIndex:idx_role_email" json:"email"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RoleEmailWhitelist) TableName() string {
	return "role_email_whitelists"
}
