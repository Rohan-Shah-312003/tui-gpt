package main

import (
	"log"

	"github.com/Rohan-Shah-312003/tui-gpt/ui"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := ui.NewApp()
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
