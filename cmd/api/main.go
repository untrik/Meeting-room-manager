package main

import (
	"log"

	"github.com/avito-internships/test-backend-1-untrik/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
