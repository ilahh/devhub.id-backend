package models

import "time"

type Contact struct {
	Id        uint      `gorm:"primaryKey" json:"id"`
	UserId    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	Github    string    `json:"github"`
	Linkedin  string    `json:"linkedin"`
	Website   string    `json:"website"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
