package models

import "time"

// BlogPost adalah satu artikel blog milik seorang user. Isinya disusun dari
// kumpulan BlogBlock yang berurutan sehingga bisa berupa kombinasi teks, gambar,
// video, audio, model 3D, maupun embed dari layanan lain.
type BlogPost struct {
	Id          uint        `gorm:"primaryKey" json:"id"`
	UserId      uint        `gorm:"index;not null" json:"user_id"`
	Slug        string      `gorm:"index;not null" json:"slug"`
	Title       string      `gorm:"not null" json:"title"`
	Excerpt     string      `json:"excerpt"`
	CoverURL    *string     `json:"cover_url"`
	Status      string      `gorm:"type:varchar(20);index;not null;default:'draft'" json:"status"` // draft | published
	Blocks      []BlogBlock `gorm:"foreignKey:PostId" json:"blocks"`
	PublishedAt *time.Time  `json:"published_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// BlogBlock adalah satu blok konten di dalam sebuah BlogPost.
// Type menentukan cara blok ditampilkan:
//   - text    : paragraf teks (isi di Text)
//   - image   : gambar hasil upload (URL di MediaURL)
//   - video   : video hasil upload (URL di MediaURL)
//   - audio   : audio hasil upload (URL di MediaURL)
//   - model3d : model 3D .glb/.gltf hasil upload (URL di MediaURL)
//   - embed   : tautan eksternal (mis. YouTube) yang di-embed (URL di Text)
type BlogBlock struct {
	Id       uint   `gorm:"primaryKey" json:"id"`
	PostId   uint   `gorm:"index;not null" json:"post_id"`
	Type     string `gorm:"type:varchar(20);not null" json:"type"`
	Text     string `gorm:"type:text" json:"text"`
	MediaURL string `json:"media_url"`
	Caption  string `json:"caption"`
	Position int    `gorm:"not null;default:0" json:"position"`
}
