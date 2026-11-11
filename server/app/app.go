package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"opentools/api"
	"opentools/config"
	"opentools/linkcheck"
	"opentools/models"
	"opentools/seed"

	"gorm.io/driver/mysql"
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
	if err := seed.Run(db); err != nil {
		return err
	}
	if err := seed.EnsureCatalog(db); err != nil {
		return err
	}
	if err := seed.EnsureSlugs(db); err != nil {
		return err
	}
	if err := seed.EnsureIntros(db); err != nil {
		return err
	}
	if err := seed.EnsureTags(db); err != nil {
		return err
	}
	if err := seed.EnsureProfile(db); err != nil {
		return err
	}
	if err := seed.EnsureAdmin(db, cfg.AdminUser, cfg.AdminPassword); err != nil {
		return err
	}
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("open-tools api %s mysql %s/%s", addr, cfg.MySQLHost+":"+cfg.MySQLPort, cfg.MySQLDB)
	api.Configure(cfg)
	r := api.Router(db)
	api.WarmIcons(db)
	api.MountAdmin(r, db, cfg)
	linkcheck.Start(db)
	return r.Run(addr)
}

func openDB(cfg config.Config) (*gorm.DB, error) {
	server, err := sql.Open("mysql", cfg.ServerDSN())
	if err != nil {
		return nil, err
	}
	defer server.Close()
	stmt := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.MySQLDB,
	)
	if _, err := server.Exec(stmt); err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	if err := db.AutoMigrate(&models.Category{}, &models.Link{}, &models.Article{}, &models.Tag{}, &models.Admin{}, &models.User{}, &models.UserLink{}, &models.PasswordReset{}); err != nil {
		return nil, err
	}
	if err := ensureUserEmail(db); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureUserEmail(db *gorm.DB) error {
	if err := db.Exec("ALTER TABLE users MODIFY COLUMN email VARCHAR(254) NULL").Error; err != nil {
		return err
	}
	if err := db.Exec("UPDATE users SET email = NULL WHERE email = ''").Error; err != nil {
		return err
	}
	var n int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'users' AND index_name = 'ux_users_email'").Scan(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return db.Exec("CREATE UNIQUE INDEX ux_users_email ON users (email)").Error
}
