package models

import "time"

type ProfessionalProfile struct {
	Id               uint      `gorm:"primaryKey" json:"id"`
	UserId           uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	CurrentWorkplace string    `json:"current_workplace"`
	CurrentPosition  string    `json:"current_position"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Skill struct {
	Id     uint   `gorm:"primaryKey" json:"id"`
	UserId uint   `gorm:"index;not null" json:"user_id"`
	Name   string `gorm:"not null" json:"name"`
	Level  string `json:"level"`
}

type WorkExperience struct {
	Id          uint   `gorm:"primaryKey" json:"id"`
	UserId      uint   `gorm:"index;not null" json:"user_id"`
	Position    string `gorm:"not null" json:"position"`
	Institution string `json:"institution"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description"`
}

type Subject struct {
	Id     uint   `gorm:"primaryKey" json:"id"`
	UserId uint   `gorm:"index;not null" json:"user_id"`
	Name   string `gorm:"not null" json:"name"`
}
