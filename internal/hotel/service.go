package hotel

import (
	"errors"

	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/go-playground/validator/v10"
)

var (
	ErrDuplicateHotelByName = errors.New("hotelService: Hotel with such name already exists")
	ErrDuplicateRoomByName  = errors.New("hotelService: Room type with such name already exists")
)

type HotelService struct {
	queries   *database.Queries
	validator validator.Validate
}

func New(queries *database.Queries, validator validator.Validate) *HotelService {
	return &HotelService{
		queries:   queries,
		validator: validator,
	}
}
