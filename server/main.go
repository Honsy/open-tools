package main

import (
	"log"

	"opentools/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
