// backend/cmd/main.go

package main

import (
	"log"

	"github.com/Culturae-org/culturae/internal/app"
)

func main() {
	application, err := app.SetupApp()
	if err != nil {
		log.Fatalf("Failed to setup app: %v", err)
	}
	application.Run()
}
