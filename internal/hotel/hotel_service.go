package hotel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/google/uuid"
)

type CreateHotelBody struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Location string `json:"location" validate:"required,min=1,max=255"`
}

type UpdateHotelBody struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Location string `json:"location" validate:"required,min=1,max=255"`
}

func (s *HotelService) createHotel(ctx context.Context, body CreateHotelBody) (database.BookingHotel, error) {

	_, err := s.queries.GetHotelByName(ctx,
		body.Name)
	if err == nil {
		return database.BookingHotel{},
			ErrDuplicateHotelByName
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return database.BookingHotel{},
			fmt.Errorf("failed to check hotel existence: %w", err)
	}

	return s.queries.CreateHotel(ctx, database.CreateHotelParams{
		ID:       uuid.New(),
		Name:     body.Name,
		Location: body.Location,
	})

}

func (s *HotelService) getHotelById(ctx context.Context, id uuid.UUID) (database.BookingHotel, error) {
	return s.queries.GetHotelById(ctx, id)
}

func (s *HotelService) updateHotel(ctx context.Context, id uuid.UUID, body UpdateHotelBody) (database.BookingHotel, error) {
	return s.queries.UpdateHotel(ctx, database.UpdateHotelParams{
		ID:       id,
		Name:     body.Name,
		Location: body.Location,
	})
}

func (s *HotelService) deleteHotel(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteHotel(ctx, id)
}
