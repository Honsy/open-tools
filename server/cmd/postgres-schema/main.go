package main

import (
	"log"

	"opentools/app"
	"opentools/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := app.OpenSchema(cfg)
	if err != nil {
		log.Fatal(err)
	}
	handle, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	if err := handle.Close(); err != nil {
		log.Fatal(err)
	}
}
