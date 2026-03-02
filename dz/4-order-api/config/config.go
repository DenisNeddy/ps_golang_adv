package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Db         DbConfig
	ServerPort string
}

type DbConfig struct {
	Dsn string
}

func NewConfig() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file, using default config")
	}
	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		ServerPort: ":8082",
	}
}
