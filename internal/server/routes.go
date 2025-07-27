package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AlexKhomenko00/hotel-system/internal/hotel"
	"github.com/AlexKhomenko00/hotel-system/internal/shared"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	hotelSvc := hotel.New(&s.queries, s.validator)
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/hotel", func(r chi.Router) {
		r.Post("/", hotelSvc.CreateHotelHandler)
		r.Get("/{id}", hotelSvc.GetHotelHandler)
		r.Put("/{id}", hotelSvc.UpdateHotelHandler)
		r.Delete("/{id}", hotelSvc.DeleteHotelHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.jwt.Verifier())
		r.Use(s.jwtAuthMiddleware)

		r.Get("/verify", verifyTokenHandler)

		r.Route("/hotel", func(r chi.Router) {
			r.Post("/", hotelSvc.CreateHotelHandler)
			r.Get("/{id}", hotelSvc.GetHotelHandler)
			r.Put("/{id}", hotelSvc.UpdateHotelHandler)
			r.Delete("/{id}", hotelSvc.DeleteHotelHandler)
		})

	})

	r.Group(func(r chi.Router) {
		r.Post("/register", s.registerHandler)
		r.Post("/login", s.loginHandler)
	})

	r.Get("/health", s.healthHandler)

	return r
}

func verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	type TokenVerificationResponse struct {
		Message string `json:"message"`
	}
	e := json.NewEncoder(w)
	usr, ok := r.Context().Value(UsrCtxKey).(UserContext)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		e.Encode(shared.ErrorRes{
			Message: "Invalid token user claims",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	e.Encode(TokenVerificationResponse{
		Message: fmt.Sprintf("Hi! %v", usr),
	})
}
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
