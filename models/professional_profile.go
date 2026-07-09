package models

import "time"

// ProfessionalProfile menyimpan informasi "tempat tugas sekarang" milik seorang
// user (relasi one-to-one lewat UserId). Skill, riwayat pekerjaan, dan mata
// pelajaran disimpan di tabel terpisah karena jumlahnya bisa banyak.
type ProfessionalProfile struct {
	Id               uint      `gorm:"primaryKey" json:"id"`
	UserId           uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	CurrentWorkplace string    `json:"current_workplace"`
	CurrentPosition  string    `json:"current_position"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Skill adalah satu keahlian milik user, dengan tingkat opsional
// (Pemula/Menengah/Mahir).
type Skill struct {
	Id     uint   `gorm:"primaryKey" json:"id"`
	UserId uint   `gorm:"index;not null" json:"user_id"`
	Name   string `gorm:"not null" json:"name"`
	Level  string `json:"level"`
}

// WorkExperience adalah satu entri riwayat pekerjaan milik user.
type WorkExperience struct {
	Id          uint   `gorm:"primaryKey" json:"id"`
	UserId      uint   `gorm:"index;not null" json:"user_id"`
	Position    string `gorm:"not null" json:"position"`
	Institution string `json:"institution"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}

// Subject adalah satu mata pelajaran yang diampu user.
type Subject struct {
	Id     uint   `gorm:"primaryKey" json:"id"`
	UserId uint   `gorm:"index;not null" json:"user_id"`
	Name   string `gorm:"not null" json:"name"`
}
