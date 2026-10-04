package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"opentools/api"
	"opentools/config"
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
	return db, db.AutoMigrate(&models.Category{}, &models.Link{}, &models.Article{}, &models.Tag{}, &models.Admin{})
}
