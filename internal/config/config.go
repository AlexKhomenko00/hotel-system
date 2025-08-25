package config

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	JWTSecret          string `validate:"required"`
	PORT               string `validate:"required,numeric"`
	DB_HOST            string `validate:"required"`
	DB_PORT            string `validate:"required,numeric"`
	DB_DATABASE        string `validate:"required"`
	DB_USERNAME        string `validate:"required"`
	DB_PASSWORD        string `validate:"required"`
	DB_SCHEMA          string `validate:"required"`
	DB_SSLMODE         string `validate:"required"`
	OVERBOOKING_FACTOR string `validate:"required,numeric,min=1"`
}

var config *Config

func GetConfig(validator *validator.Validate) *Config {
	if config == nil {
		config = loadAndValidateConfig(validator)
	}
	return config
}

func loadAndValidateConfig(validator *validator.Validate) *Config {
	cfg := &Config{
		JWTSecret:          os.Getenv("JWT_SECRET"),
		PORT:               os.Getenv("PORT"),
		DB_HOST:            os.Getenv("DB_HOST"),
		DB_PORT:            os.Getenv("DB_PORT"),
		DB_DATABASE:        os.Getenv("DB_DATABASE"),
		DB_USERNAME:        os.Getenv("DB_USERNAME"),
		DB_PASSWORD:        os.Getenv("DB_PASSWORD"),
		DB_SCHEMA:          os.Getenv("DB_SCHEMA"),
		DB_SSLMODE:         os.Getenv("DB_SSLMODE"),
		OVERBOOKING_FACTOR: os.Getenv("OVERBOOKING_FACTOR"),
	}

	if err := validator.Struct(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return cfg
}
