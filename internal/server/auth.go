package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/AlexKhomenko00/hotel-system/internal/shared"
	"github.com/go-chi/jwtauth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userContextKey string

var UsrCtxKey userContextKey = "user"

type UserContext struct {
	Id    string
	Email string
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

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var body AuthBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.cfg.Validator().Struct(body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resEncoder := json.NewEncoder(w)

	usr, err := s.queries.GetUserByEmail(r.Context(), body.Email)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusUnauthorized)
		resEncoder.Encode(shared.ErrorRes{Message: "Invalid credentials"})
		return
	}

	if err != nil {
		slog.Error("database error during login", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(body.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Invalid credentials",
		})
		return
	}

	_, tokenString, err := s.jwt.Encode(map[string]interface{}{
		"userId": usr.ID,
		"email":  usr.Email,
	})
	if err != nil {
		slog.Error("Failed to generate JWT token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	resEncoder.Encode(LoginResponse{
		AccessToken: tokenString,
		User: UserResponse{
			Id:    usr.ID.String(),
			Email: usr.Email,
		},
	})
}

func (s *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	var body AuthBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resEncoder := json.NewEncoder(w)

	if err := s.cfg.Validator().Struct(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: fmt.Sprintf("Invalid body: %v", err.Error()),
		})
		return
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)

	if err != nil {
		slog.Error("unable to hash password", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = s.queries.GetUserByEmail(r.Context(), body.Email)
	if err == nil {
		// User exists - add timing delay to match bcrypt operation
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusBadRequest)
		err := resEncoder.Encode(shared.ErrorRes{
			Message: "Invalid credentials",
		})
		if err != nil {
			slog.Error("failed to encode response", "error", err)
		}
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	usr, err := s.queries.InsertUser(r.Context(), database.InsertUserParams{
		ID:           uuid.New(),
		Email:        body.Email,
		PasswordHash: string(hashedPasswordBytes),
	})
	if err != nil {
		slog.Error("unable to insert user into db", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(UserResponse{
		Id:    usr.ID.String(),
		Email: usr.Email,
	})

	if err != nil {
		slog.Error("failed to send user register response", "error", err)
	}
}

func (s *Server) jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if token == nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		usrID, ok := claims["userId"].(string)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		uid, err := uuid.Parse(usrID)
		if err != nil {
			http.Error(w, "invalid user ID in token", http.StatusUnauthorized)
			return
		}

		usr, err := s.queries.GetUserById(r.Context(), uid)
		if err != nil {
			slog.Error("Can't retrieve user from token", "error", err, "user_id", usrID)
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		usrCtx := context.WithValue(r.Context(), UsrCtxKey, UserContext{
			Id:    usr.ID.String(),
			Email: usr.Email,
		})

		next.ServeHTTP(w, r.WithContext(usrCtx))
	})
}
