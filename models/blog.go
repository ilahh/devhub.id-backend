package models

import "time"

type BlogPost struct {
	Id          uint        `gorm:"primaryKey" json:"id"`
	UserId      uint        `gorm:"index;not null" json:"user_id"`
	Slug        string      `gorm:"index;not null" json:"slug"`
	Title       string      `gorm:"not null" json:"title"`
	Excerpt     string      `json:"excerpt"`
	CoverURL    *string     `json:"cover_url"`
	Content     string      `gorm:"type:text" json:"content"`
	Status      string      `gorm:"type:varchar(20);index;not null;default:'draft'" json:"status"`
	Blocks      []BlogBlock `gorm:"foreignKey:PostId" json:"blocks"`
	PublishedAt *time.Time  `json:"published_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type BlogBlock struct {
	Id       uint   `gorm:"primaryKey" json:"id"`
	PostId   uint   `gorm:"index;not null" json:"post_id"`
	Type     string `gorm:"type:varchar(20);not null" json:"type"`
	Text     string `gorm:"type:text" json:"text"`
	MediaURL string `json:"media_url"`
	Caption  string `json:"caption"`
	Position int    `gorm:"not null;default:0" json:"position"`
}
