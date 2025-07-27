package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/AlexKhomenko00/hotel-system/internal/server/jwt"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port    int
	jwt     jwt.Authenticator
	db      database.Service
	queries database.Queries
	cfg     *config.Config
}

func NewServer() *http.Server {
	cfg := config.GetConfig()
	dbService, err := database.Create(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database %w", err)
	}

	port, _ := strconv.Atoi(cfg.PORT)
	jwtAuth := jwt.NewAuthenticator(cfg.JWTSecret)
	NewServer := &Server{
		port:    port,
		jwt:     jwtAuth,
		db:      dbService,
		queries: *database.New(dbService.GetDB()),
		cfg:     cfg,
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
