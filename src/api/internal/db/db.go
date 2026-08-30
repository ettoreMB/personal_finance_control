package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}
