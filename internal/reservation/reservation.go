package reservation

import (
	"errors"

	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/go-playground/validator/v10"
)

var (
	ErrInventoryCapacityReached = errors.New("reached maximum inventory capacity for date")
	ErrOptimisticLockMismatch   = errors.New("optimistic lock mismatch - inventory updated by another transaction")
	ErrInvalidReservationID     = errors.New("invalid reservation ID format")
	ErrDuplicateReservation     = errors.New("reservation with this ID already exists")
	ErrInvalidOverbookingFactor = errors.New("invalid overbooking factor")
	ErrInventoryNotFound        = errors.New("no inventory found for specified dates")
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
