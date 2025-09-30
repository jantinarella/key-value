package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
)

type Config struct {
	APIKey        string `env:"API_KEY"`
	Port          string `env:"PORT"`
	Environment   string `env:"ENVIRONMENT"`
	KVServiceAddr string `env:"KV_SERVICE_ADDR"`
}

func Load() *Config {
	// Try to load .env file, but don't fail if it doesn't exist (for Docker)
	err := godotenv.Load()
	if err != nil {
		log.Infof("Not loading .env file")
	}

	kvServiceAddr := os.Getenv("KV_SERVICE_ADDR")
	if kvServiceAddr == "" {
		kvServiceAddr = "localhost:50051" // Default address
	}

	return &Config{
		APIKey:        os.Getenv("API_KEY"),
		Port:          os.Getenv("PORT"),
		Environment:   os.Getenv("ENVIRONMENT"),
		KVServiceAddr: kvServiceAddr,
	}
}
