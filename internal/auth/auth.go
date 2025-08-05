package auth

import (
	"github.com/AlexKhomenko00/hotel-system/internal/auth/jwt"
	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/go-playground/validator/v10"
)

type AuthService struct {
	queries   *database.Queries
	validator *validator.Validate
	cfg       *config.Config
	jwt       jwt.Authenticator
}

type AuthBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=30"`
}

type UserResponse struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

func New(queries *database.Queries, validator *validator.Validate, cfg *config.Config) *AuthService {
	jwtAuth := jwt.NewAuthenticator(cfg.JWTSecret)

	return &AuthService{
		queries:   queries,
		validator: validator,
		cfg:       cfg,
		jwt:       jwtAuth,
	}
}
