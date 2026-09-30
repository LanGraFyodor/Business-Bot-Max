package main

import (
	"docspace/internal/app"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--check-bot" {
		if err := app.CheckBot(); err != nil {
			log.Fatal(err)
		}
		return
	}
	app.Run()
}
