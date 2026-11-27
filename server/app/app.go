package app

import (
	"fmt"
	"log"
	"os"

	"opentools/api"
	"opentools/config"
	"opentools/editorial"
	"opentools/lab"
	"opentools/linkcheck"
	"opentools/models"
	"opentools/seed"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	if err := seed.EnsureSlugs(db); err != nil {
		return err
	}
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
		if port := os.Getenv("PORT"); port != "" {
			addr = "127.0.0.1:" + port
		}
	}
	log.Printf("open-tools api %s postgresql %s/%s", addr, cfg.PostgresHost+":"+cfg.PostgresPort, cfg.PostgresDB)
	api.Configure(cfg)
	r := api.Router(db)
	api.WarmIcons(db)
	api.MountAdmin(r, db)
	api.MountLab(r, db, cfg)
	if cfg.DailyArticles {
		editorial.StartDaily(db, lab.FromConfig(cfg))
	}
	linkcheck.Start(db)
	return r.Run(addr)
}

func openDB(cfg config.Config) (*gorm.DB, error) {
	db, err := OpenSchema(cfg)
	if err != nil {
		return nil, err
	}
	if err := seed.MigrateAccounts(db, cfg.AdminUser, cfg.AdminPassword); err != nil {
		return nil, fmt.Errorf("migrate accounts: %w", err)
	}
	return db, nil
}

// OpenSchema initializes PostgreSQL tables without creating or modifying a user account.
func OpenSchema(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(82422402)").Error; err != nil {
			return err
		}
		if err := tx.AutoMigrate(&models.Category{}, &models.CategoryGuide{}, &models.HistoryEvent{}, &models.Link{}, &models.Article{}, &models.ArticleComment{}, &models.Feedback{}, &models.Tag{}, &models.User{}, &models.UserLink{}, &models.PasswordReset{}, &models.InternetRankMonth{}, &models.InternetRankEntry{}, &models.InternetRankBlock{}, &models.SubmissionReview{}); err != nil {
			return err
		}
		if err := tx.Exec("UPDATE users SET email = NULL WHERE email = ''").Error; err != nil {
			return err
		}
		if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (lower(email))").Error; err != nil {
			return err
		}
		return tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_users_username_lower ON users (lower(username))").Error
	})
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}
