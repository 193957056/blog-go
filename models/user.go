package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Username     string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Email        string         `gorm:"size:100;uniqueIndex" json:"email"`
	Nickname     string         `gorm:"size:50" json:"nickname"`
	Avatar       string         `gorm:"size:500" json:"avatar"`
	Role         string         `gorm:"size:20;default:'admin'" json:"role"` // admin, editor
}
