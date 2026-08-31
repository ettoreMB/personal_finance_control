package auth

import "time"

type User struct {
	ID           uint `gorm:"primaryKey"`
	Email        string
	PasswordHash string
	CPF          string `gorm:"column:cpf"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}

type Session struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Session) TableName() string {
	return "sessions"
}
