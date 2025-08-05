package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port      int
	db        database.Service
	queries   *database.Queries
	cfg       *config.Config
	validator *validator.Validate
}

func NewServer() *http.Server {
	validator := validator.New()
	cfg := config.GetConfig(validator)
	dbService, err := database.Create(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database %w", err)
	}

	port, _ := strconv.Atoi(cfg.PORT)
	NewServer := &Server{
		port:      port,
		db:        dbService,
		queries:   database.New(dbService.GetDB()),
		cfg:       cfg,
		validator: validator,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
