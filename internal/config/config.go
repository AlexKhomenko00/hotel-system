package config

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	JWTSecret   string `validate:"required"`
	PORT        string `validate:"required,numeric"`
	DB_HOST     string `validate:"required"`
	DB_PORT     string `validate:"required,numeric"`
	DB_DATABASE string `validate:"required"`
	DB_USERNAME string `validate:"required"`
	DB_PASSWORD string `validate:"required"`
	DB_SCHEMA   string `validate:"required"`
	DB_SSLMODE  string `validate:"required"`

	validator *validator.Validate
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = loadAndValidateConfig()
	}
	return config
}

func loadAndValidateConfig() *Config {
	cfg := &Config{
		JWTSecret:   os.Getenv("JWT_SECRET"),
		PORT:        os.Getenv("PORT"),
		DB_HOST:     os.Getenv("BLUEPRINT_DB_HOST"),
		DB_PORT:     os.Getenv("BLUEPRINT_DB_PORT"),
		DB_DATABASE: os.Getenv("BLUEPRINT_DB_DATABASE"),
		DB_USERNAME: os.Getenv("BLUEPRINT_DB_USERNAME"),
		DB_PASSWORD: os.Getenv("BLUEPRINT_DB_PASSWORD"),
		DB_SCHEMA:   os.Getenv("BLUEPRINT_DB_SCHEMA"),
		DB_SSLMODE:  os.Getenv("BLUEPRINT_DB_SSLMODE"),
		validator:   validator.New(),
	}

	if err := cfg.validator.Struct(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return cfg
}

func (c *Config) Validator() *validator.Validate {
	return c.validator
}
