package services

import "gorm.io/gorm"

func NewService(db *gorm.DB) *gorm.DB {
	return db
}
