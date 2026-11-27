package seed

import (
	"fmt"
	"log"
	"time"

	"opentools/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type accountMigration struct {
	Name string `gorm:"primaryKey;size:64"`
}

type legacyAdmin struct {
	ID           uint
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// MigrateAccounts copies administrator credentials once, keeping existing users and their IDs.
func MigrateAccounts(db *gorm.DB, username, password string) error {
	if err := db.AutoMigrate(&accountMigration{}); err != nil {
		return err
	}
	var renamed []string
	err := db.Transaction(func(tx *gorm.DB) error {
		marker := accountMigration{Name: "unify-admin-accounts"}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&marker)
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}

		var admins []legacyAdmin
		if tx.Migrator().HasTable("admins") {
			if err := tx.Table("admins").Order("id").Find(&admins).Error; err != nil {
				return err
			}
		}
		if len(admins) == 0 {
			var count int64
			if err := tx.Model(&models.User{}).Where("is_admin = ?", true).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return nil
			}
			if password == "" {
				return fmt.Errorf("config.yaml 里要有首次初始化用的 adminPassword")
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			admins = []legacyAdmin{{Username: username, PasswordHash: string(hash)}}
		}

		// Insert unchanged names first so an alias cannot take another administrator's name.
		var conflicts []legacyAdmin
		for _, admin := range admins {
			available, err := usernameAvailable(tx, admin.Username)
			if err != nil {
				return err
			}
			if !available {
				conflicts = append(conflicts, admin)
				continue
			}
			if err := createAdminUser(tx, admin, admin.Username); err != nil {
				return err
			}
		}
		for _, admin := range conflicts {
			alias, err := availableAdminAlias(tx, admin.ID)
			if err != nil {
				return err
			}
			if err := createAdminUser(tx, admin, alias); err != nil {
				return err
			}
			renamed = append(renamed, fmt.Sprintf("administrator %q migrated as %q", admin.Username, alias))
		}
		return nil
	})
	if err == nil {
		for _, message := range renamed {
			log.Print(message)
		}
	}
	return err
}

func usernameAvailable(db *gorm.DB, username string) (bool, error) {
	var count int64
	err := db.Model(&models.User{}).Where("lower(username) = lower(?)", username).Count(&count).Error
	return count == 0, err
}

func availableAdminAlias(db *gorm.DB, id uint) (string, error) {
	base := fmt.Sprintf("admin-%d", id)
	name := base
	for suffix := 2; ; suffix++ {
		available, err := usernameAvailable(db, name)
		if err != nil || available {
			return name, err
		}
		name = fmt.Sprintf("%s-%d", base, suffix)
	}
}

func createAdminUser(db *gorm.DB, admin legacyAdmin, username string) error {
	user := models.User{
		Username: username, PasswordHash: admin.PasswordHash, IsAdmin: true, CreatedAt: admin.CreatedAt,
	}
	return db.Omit("Email").Create(&user).Error
}
