package models

import "time"

type Portfolio struct {
	Id          uint      `gorm:"primaryKey" json:"id"`
	UserId      uint      `gorm:"index;not null" json:"user_id"`
	Category    string    `gorm:"index" json:"category"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Issuer      string    `json:"issuer"`
	IssuedDate  string    `json:"issued_date"`
	TechStack   string    `json:"tech_stack"`
	ImageURL    *string   `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
