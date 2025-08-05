package server

import (
	"encoding/json"
	"net/http"

	"github.com/AlexKhomenko00/hotel-system/internal/auth"
	"github.com/AlexKhomenko00/hotel-system/internal/hotel"
	"github.com/AlexKhomenko00/hotel-system/internal/reservation"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	hotelSvc := hotel.New(s.queries, s.validator)
	authSvc := auth.New(s.queries, s.validator, s.cfg)
	reservationSvc := reservation.New(s.queries, s.validator, s.cfg, s.db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authSvc.RegisterHandlers(r)

	r.Group(func(r chi.Router) {
		authSvc.SetupJWTAuthMiddleware(r)

		r.Route("/hotel", func(r chi.Router) {
			registerHotelRoutes(r, hotelSvc)
			registerRoomTypesRoutes(r, hotelSvc)
		})

		r.Route("/reservation", func(r chi.Router) {
			reservationSvc.RegisterHandlers(r)
		})
	})

	r.Get("/health", s.healthHandler)

	return r
}

func registerHotelRoutes(r chi.Router, hotelSvc *hotel.HotelService) {
	r.Post("/", hotelSvc.CreateHotelHandler)
	r.Get("/{id}", hotelSvc.GetHotelHandler)
	r.Put("/{id}", hotelSvc.UpdateHotelHandler)
	r.Delete("/{id}", hotelSvc.DeleteHotelHandler)
}

func registerRoomTypesRoutes(r chi.Router, hotelSvc *hotel.HotelService) {
	r.Route("/{hotelId}/rooms", func(r chi.Router) {
		r.Post("/", hotelSvc.AddRoomTypeHandler)
		r.Get("/{id}", hotelSvc.GetRoomTypeHandler)
		r.Put("/{id}", hotelSvc.UpdateRoomTypeHandler)
		r.Delete("/{id}", hotelSvc.DeleteRoomTypeHandler)
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
