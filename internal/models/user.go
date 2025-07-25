package models

import (
	"time"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Password  string    `json:"password" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Role      string    `json:"role" gorm:"default:'user'"`
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
