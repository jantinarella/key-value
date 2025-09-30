package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey      string `env:"API_KEY"`
	Port        string `env:"PORT"`
	Environment string `env:"ENVIRONMENT"`
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Not loading .env file")
	}

	return &Config{
		APIKey:      os.Getenv("API_KEY"),
		Port:        os.Getenv("PORT"),
		Environment: os.Getenv("ENVIRONMENT"),
	}
}
