package reservation

import (
	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/go-playground/validator/v10"
)

type ReservationState string

const (
	ReservationStatusPending  ReservationState = "pending"
	ReservationStatusPaid     ReservationState = "paid"
	ReservationStatusRejected ReservationState = "rejected"
	ReservationStatusRefunded ReservationState = "refunded"
)

type ReservationService struct {
	queries   *database.Queries
	validator *validator.Validate
	cfg       *config.Config
	db        database.Service
}

func New(queries *database.Queries, validator *validator.Validate, cfg *config.Config, db database.Service) *ReservationService {
	return &ReservationService{
		queries:   queries,
		validator: validator,
		cfg:       cfg,
		db:        db,
	}
}
