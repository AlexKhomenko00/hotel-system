package hotel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/google/uuid"
)

type AddRoomType struct {
	RoomType    string `json:"roomType" validate:"min=0,max=255"`
	HotelID     string
	Description string `json:"description" validate:"omitempty"`
}

type UpdateRoomType struct {
	RoomType    string `json:"roomType" validate:"min=0,max=255"`
	Description string `json:"description" validate:"omitempty"`
}

func (s *HotelService) addRoomType(ctx context.Context, body AddRoomType) (database.BookingRoomType, error) {
	_, err := s.queries.FindRoomTypesByHotelIdAndName(ctx, database.FindRoomTypesByHotelIdAndNameParams{
		HotelID: uuid.MustParse(body.HotelID),
		Name:    body.RoomType,
	})

	if err == nil {
		return database.BookingRoomType{}, ErrDuplicateRoomByName
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return database.BookingRoomType{},
			fmt.Errorf("failed to check room type existence: %w", err)
	}

	roomType, err := s.queries.CreateRoomType(ctx, database.CreateRoomTypeParams{
		HotelID: uuid.MustParse(body.HotelID),
		Name:    body.RoomType,
		Description: sql.NullString{
			String: body.Description,
			Valid:  body.Description != "",
		},
	})

	if err != nil {
		return database.BookingRoomType{}, err
	}

	return roomType, nil
}

func (s *HotelService) getRoomTypeById(ctx context.Context, hotelID, roomTypeID uuid.UUID) (database.BookingRoomType, error) {
	roomType, err := s.queries.GetRoomTypeByIdAndHotelId(ctx, database.GetRoomTypeByIdAndHotelIdParams{
		ID:      roomTypeID,
		HotelID: hotelID,
	})

	if err != nil {
		return database.BookingRoomType{}, err
	}

	return roomType, nil
}

func (s *HotelService) updateRoomType(ctx context.Context, hotelID, roomTypeID uuid.UUID, body UpdateRoomType) (database.BookingRoomType, error) {
	existingRoomType, err := s.queries.FindRoomTypesByHotelIdAndName(ctx, database.FindRoomTypesByHotelIdAndNameParams{
		HotelID: hotelID,
		Name:    body.RoomType,
	})

	if err == nil && existingRoomType.ID != roomTypeID {
		return database.BookingRoomType{}, ErrDuplicateRoomByName
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return database.BookingRoomType{}, fmt.Errorf("failed to check room type existence: %w", err)
	}

	roomType, err := s.queries.UpdateRoomType(ctx, database.UpdateRoomTypeParams{
		ID:      roomTypeID,
		HotelID: hotelID,
		Name:    body.RoomType,
		Description: sql.NullString{
			String: body.Description,
			Valid:  body.Description != "",
		},
	})

	if err != nil {
		return database.BookingRoomType{}, err
	}

	return roomType, nil
}

func (s *HotelService) deleteRoomType(ctx context.Context, hotelID, roomTypeID uuid.UUID) error {
	err := s.queries.DeleteRoomType(ctx, database.DeleteRoomTypeParams{
		ID:      roomTypeID,
		HotelID: hotelID,
	})

	if err != nil {
		return err
	}

	return nil
}
