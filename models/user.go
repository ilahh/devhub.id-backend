package models

import "time"

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleModerator, RoleMember:
		return true
	}
	return false
}

type User struct {
	Id        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Username  *string   `gorm:"uniqueIndex" json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Role      Role      `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
