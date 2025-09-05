package reservation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type MakeReservationBody struct {
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	HotelID       uuid.UUID `json:"hotelId" validate:"uuid4"`
	RoomTypeID    uuid.UUID `json:"roomTypeId" validate:"uuid4"`
	ReservationId string    `json:"reservationId" validate:"uuid4"`
}

func (s *ReservationService) makeReservation(ctx context.Context, guestID uuid.UUID, body MakeReservationBody) error {
	tx, err := s.db.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start make reservation transaction %w", err)
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	inventory, err := qtx.GetHotelInventoryForRange(ctx, database.GetHotelInventoryForRangeParams{
		RoomTypeID: body.RoomTypeID,
		HotelID:    body.HotelID,
		Date:       body.StartDate,
		Date_2:     body.EndDate,
	})

	if err != nil {
		return fmt.Errorf("failed to get hotel %q inventory: %w", body.HotelID, err)
	}

	expectedDays := int(body.EndDate.Sub(body.StartDate).Hours()/24) + 1
	if len(inventory) != expectedDays {
		return ErrInventoryNotFound
	}

	overbookingFactor, err := strconv.ParseFloat(s.cfg.OVERBOOKING_FACTOR, 64)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidOverbookingFactor, err)
	}

	for _, inventoryDate := range inventory {
		maxCapacity := int32(float64(inventoryDate.TotalInventory) * overbookingFactor)
		if (inventoryDate.TotalReserved + 1) > maxCapacity {
			return fmt.Errorf("%w: %s", ErrInventoryCapacityReached, inventoryDate.Date)
		}
	}

	reservationUUID, err := uuid.Parse(body.ReservationId)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidReservationID, err)
	}

	_, err = qtx.InsertReservation(ctx, database.InsertReservationParams{
		ID:         reservationUUID,
		HotelID:    body.HotelID,
		RoomTypeID: body.RoomTypeID,
		StartDate:  body.StartDate,
		EndDate:    body.EndDate,
		Status:     string(ReservationStatusPending),
		GuestID:    guestID,
	})

	if err != nil {
		// Check for duplicate key violation (primary key constraint on reservation ID)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
		   strings.Contains(err.Error(), "reservations_pkey") {
			return fmt.Errorf("%w: %v", ErrDuplicateReservation, err)
		}
		return fmt.Errorf("failed to insert reservation %w", err)
	}

	g := new(errgroup.Group)

	for _, inventoryDate := range inventory {
		g.Go(func() error {
			rowsCount, err := qtx.UpdateRoomTypeInventoryForDate(ctx, database.UpdateRoomTypeInventoryForDateParams{
				RoomTypeID:    inventoryDate.RoomTypeID,
				TotalReserved: 1,
				HotelID:       inventoryDate.HotelID,
				Date:          inventoryDate.Date,
				Version:       inventoryDate.Version,
			})

			if err != nil {
				return err
			}

			if rowsCount == 0 {
				return fmt.Errorf("%w: date %q, hotel %q, room type %q", ErrOptimisticLockMismatch, inventoryDate.Date, inventoryDate.HotelID, inventoryDate.RoomTypeID)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return tx.Commit()

}

func (s *ReservationService) generateReservationId() string {
	/*
		This is oversimplification even though okay-ish one.
		Probably for real production system distributed unique id generator like twitter's snowflake may by used.
	**/
	return uuid.NewString()
}

type GetAvailabilityQuery struct {
	HotelID  uuid.UUID
	CheckIn  time.Time
	CheckOut time.Time
}

func (s *ReservationService) getRoomAvailability(ctx context.Context, payload GetAvailabilityQuery) ([]database.GetRoomAvailabilityByDatesRow, error) {
	overbookingFactor, err := strconv.ParseFloat(s.cfg.OVERBOOKING_FACTOR, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOverbookingFactor, err)
	}

	availability, err := s.queries.GetRoomAvailabilityByDates(ctx, database.GetRoomAvailabilityByDatesParams{
		HotelID:     payload.HotelID,
		CheckIn:     payload.CheckIn,
		CheckOut:    payload.CheckOut,
		Overbooking: float64(overbookingFactor),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get room availability for hotel %q: %w", payload.HotelID, err)
	}

	return availability, nil
}
