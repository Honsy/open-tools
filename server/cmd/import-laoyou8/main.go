package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"opentools/config"
	"opentools/legacydata"

	mysqlDSN "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	apply := flag.Bool("apply", false, "write the snapshot into the opentools database")
	flag.Parse()
	if err := run(*apply); err != nil {
		log.Fatal(err)
	}
}

func run(apply bool) error {
	dsn, err := mysqlDSN.ParseDSN(os.Getenv("LAOYOU8_SOURCE_DSN"))
	if err != nil || dsn.DBName != "laoyouba" {
		return errors.New("set LAOYOU8_SOURCE_DSN to a read-only MySQL connection for laoyouba")
	}
	dsn.ParseTime, dsn.Timeout, dsn.ReadTimeout = true, 8*time.Second, 45*time.Second
	dbConfig := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	source, err := gorm.Open(mysql.Open(dsn.FormatDSN()), dbConfig)
	if err != nil {
		return errors.New("cannot connect to the LaoYou8 source database")
	}
	snapshot, err := legacydata.ReadSnapshot(source)
	if err != nil {
		return err
	}
	fmt.Printf("Source snapshot: %d sites, %d articles, %d article comments, %d site messages\n",
		len(snapshot.Sites), len(snapshot.Articles), len(snapshot.Comments), len(snapshot.Messages))
	if !apply {
		return nil
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.MySQLDB != "opentools" {
		return errors.New("destination database must be opentools")
	}
	destination, err := gorm.Open(mysql.Open(cfg.DSN()), dbConfig)
	if err != nil {
		return errors.New("cannot connect to the opentools destination database")
	}
	result, err := legacydata.Import(destination, snapshot)
	if err != nil {
		return fmt.Errorf("import rolled back: %w", err)
	}
	fmt.Printf("Imported %d sites, %d articles, %d comments and %d site messages. Skipped %d existing URLs, %d invalid URLs and %d orphan comments.\n",
		result.Sites, result.Articles, result.Comments, result.Messages, result.ExistingSites, result.InvalidSites, result.OrphanComments)
	return nil
}
