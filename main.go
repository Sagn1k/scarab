package main

import (
	"log"

	"github.com/Sagn1k/scarab/api"
	"github.com/Sagn1k/scarab/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	cfg := config.NewConfig()

	server := api.NewServer(cfg)
	server.Start()
}
