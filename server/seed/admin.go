package seed

import (
	"opentools/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func EnsureAdmin(db *gorm.DB, username, password string) error {
	var n int64
	if err := db.Model(&models.Admin{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Create(&models.Admin{Username: username, PasswordHash: string(hash)}).Error
}
