package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/AlexKhomenko00/hotel-system/internal/reservation"
	"github.com/AlexKhomenko00/hotel-system/internal/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestInventorySingle     = 1
	TestInventoryMax        = 10
	TestInventoryOverbooked = 12
	TestInventoryReserved   = 2
	TestInventoryPartial    = 5
)

type MakeReservationBody struct {
	StartDate     shared.Date `json:"startDate"`
	EndDate       shared.Date `json:"endDate"`
	HotelID       uuid.UUID   `json:"hotelId"`
	RoomTypeID    uuid.UUID   `json:"roomTypeId"`
	ReservationId string      `json:"reservationId"`
}

type GenerateIdResponse struct {
	ReservationId string `json:"reservation_id"`
}

type GetAvailabilityResponse struct {
	Availability []interface{} `json:"availability"`
}

var reservationSuite *TestSuite

func init() {
	reservationSuite = GetTestSuite()
	config := reservationSuite.GetConfig()
	reservationSvc := reservation.New(reservationSuite.GetQueries(), reservationSuite.GetValidator(), &config, reservationSuite.GetDB())
	reservationSuite.RegisterPrivateHandlers(reservationSvc.RegisterHandlers, "/reservation")
}

func TestReservationOptimisticLocking(t *testing.T) {
	t.Parallel()

	t.Run("should_prevent_concurrent_reservations_for_same_room_and_dates", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 3).Truncate(24 * time.Hour)

		ctx := context.Background()
		dates := []time.Time{}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d)
		}
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          dates,
			TotalInventory: TestInventorySingle,
		})
		require.NoError(t, err)

		user1, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		user2, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		reservationId1 := uuid.New().String()
		reservationId2 := uuid.New().String()

		reservation1 := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: reservationId1,
		}

		reservation2 := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: reservationId2,
		}

		var wg sync.WaitGroup
		var resp1, resp2 *http.Response

		wg.Add(2)

		go func() {
			defer wg.Done()
			resp1, _ = reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation1, user1)
		}()

		go func() {
			defer wg.Done()
			resp2, _ = reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation2, user2)
		}()

		wg.Wait()

		successCount := 0
		if resp1 != nil && resp1.StatusCode == http.StatusCreated {
			successCount++
		}
		if resp2 != nil && resp2.StatusCode == http.StatusCreated {
			successCount++
		}

		if resp1 != nil {
			resp1.Body.Close()
		}
		if resp2 != nil {
			resp2.Body.Close()
		}

		assert.Equal(t, 1, successCount, "Only one reservation should succeed")

		var reservationCount int
		err = reservationSuite.GetDB().GetDB().QueryRowContext(ctx, `
      		SELECT COUNT(*) FROM booking.reservations 
      		WHERE hotel_id = $1 AND room_type_id = $2
		`, hotel.ID, roomType.ID).Scan(&reservationCount)

		require.NoError(t, err)
		assert.Equal(t, TestInventorySingle, reservationCount, "Only one reservation should exist in the database")
	})

	t.Run("should_prevent_duplicate_reservation_ids", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 3).Truncate(24 * time.Hour)

		ctx := context.Background()
		dates := []time.Time{}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d)
		}
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          dates,
			TotalInventory: TestInventoryMax,
		})
		require.NoError(t, err)

		resp, err := reservationSuite.MakeAuthenticatedRequest("GET", "/reservation/generate-id", nil, user)
		require.NoError(t, err)
		defer resp.Body.Close()

		var idResp GenerateIdResponse
		err = json.NewDecoder(resp.Body).Decode(&idResp)
		require.NoError(t, err)
		duplicateID := idResp.ReservationId

		reservation1 := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: duplicateID,
		}

		reservation2 := MakeReservationBody{
			StartDate:     shared.Date(startDate.AddDate(0, 0, 5)),
			EndDate:       shared.Date(endDate.AddDate(0, 0, 5)),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: duplicateID,
		}

		resp1, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation1, user)
		require.NoError(t, err)
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		resp2, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation2, user)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusConflict, resp2.StatusCode, "Second reservation with duplicate ID should fail")
	})

	t.Run("should_reject_second_reservation_when_capacity_reached", func(t *testing.T) {
		t.Parallel()
		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user1, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		user2, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := startDate.AddDate(0, 0, 1)

		ctx := context.Background()
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          []time.Time{startDate, endDate},
			TotalInventory: TestInventorySingle,
		})
		require.NoError(t, err)

		reservation1 := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: uuid.New().String(),
		}

		resp1, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation1, user1)
		require.NoError(t, err)
		defer resp1.Body.Close()
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		reservation2 := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: uuid.New().String(),
		}

		resp2, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation2, user2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusConflict, resp2.StatusCode, "Should fail due to inventory capacity reached")
	})
}

func TestReservationBusinessLogic(t *testing.T) {
	t.Parallel()

	t.Run("should_reject_reservation_when_capacity_reached", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 2).Truncate(24 * time.Hour)

		ctx := context.Background()
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          []time.Time{startDate, endDate},
			TotalInventory: TestInventorySingle,
		})
		require.NoError(t, err)
		_, err = reservationSuite.GetDB().GetDB().ExecContext(ctx, `
			UPDATE booking.room_type_inventory 
			SET total_reserved = $1
			WHERE hotel_id = $2 AND room_type_id = $3 AND date IN ($4, $5)
		`, TestInventorySingle, hotel.ID, roomType.ID, startDate, endDate)
		require.NoError(t, err)

		reservation := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: uuid.New().String(),
		}

		resp, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation, user)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should fail due to inventory capacity reached")
	})

	t.Run("should_succeed_within_overbooking_limits", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 2).Truncate(24 * time.Hour)

		ctx := context.Background()
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          []time.Time{startDate, endDate},
			TotalInventory: TestInventoryMax,
		})
		require.NoError(t, err)
		_, err = reservationSuite.GetDB().GetDB().ExecContext(ctx, `
			UPDATE booking.room_type_inventory 
			SET total_reserved = $1
			WHERE hotel_id = $2 AND room_type_id = $3 AND date IN ($4, $5)
		`, TestInventoryPartial, hotel.ID, roomType.ID, startDate, endDate)
		require.NoError(t, err)

		idResp, err := reservationSuite.MakeAuthenticatedRequest("GET", "/reservation/generate-id", nil, user)
		require.NoError(t, err)
		defer idResp.Body.Close()

		var idRespData GenerateIdResponse
		err = json.NewDecoder(idResp.Body).Decode(&idRespData)
		require.NoError(t, err)

		reservation := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: idRespData.ReservationId,
		}

		resp, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation, user)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("should_validate_reservation_dates", func(t *testing.T) {
		t.Parallel()

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		futureDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)

		reservation := MakeReservationBody{
			StartDate:     shared.Date(futureDate),
			EndDate:       shared.Date(futureDate.AddDate(0, 0, 1)),
			HotelID:       uuid.New(), // Non-existent hotel
			RoomTypeID:    uuid.New(), // Non-existent room type
			ReservationId: uuid.New().String(),
		}

		resp, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation, user)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Should fail when no inventory exists for non-existent hotel/room type")
	})

	t.Run("should_handle_invalid_reservation_id", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 2).Truncate(24 * time.Hour)

		ctx := context.Background()
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          []time.Time{startDate, endDate},
			TotalInventory: TestInventoryMax,
		})
		require.NoError(t, err)

		reservation := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: "invalid-uuid",
		}

		resp, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation, user)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should fail due to invalid reservation ID format")
	})

	t.Run("should_fail_when_inventory_missing_for_some_dates", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 4).Truncate(24 * time.Hour)

		ctx := context.Background()
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          []time.Time{startDate, startDate.AddDate(0, 0, 1)},
			TotalInventory: TestInventoryMax,
		})
		require.NoError(t, err)

		idResp, err := reservationSuite.MakeAuthenticatedRequest("GET", "/reservation/generate-id", nil, user)
		require.NoError(t, err)
		defer idResp.Body.Close()

		var idRespData GenerateIdResponse
		err = json.NewDecoder(idResp.Body).Decode(&idRespData)
		require.NoError(t, err)

		reservation := MakeReservationBody{
			StartDate:     shared.Date(startDate),
			EndDate:       shared.Date(endDate),
			HotelID:       hotel.ID,
			RoomTypeID:    roomType.ID,
			ReservationId: idRespData.ReservationId,
		}

		resp, err := reservationSuite.MakeAuthenticatedRequest("POST", "/reservation", reservation, user)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Should fail when inventory is missing for some dates")
	})
}

func TestRoomAvailability(t *testing.T) {
	t.Parallel()

	t.Run("should_return_available_rooms", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 3).Truncate(24 * time.Hour)

		ctx := context.Background()
		dates := []time.Time{}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d)
		}
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          dates,
			TotalInventory: TestInventoryMax,
		})
		require.NoError(t, err)
		_, err = reservationSuite.GetDB().GetDB().ExecContext(ctx, `
			UPDATE booking.room_type_inventory 
			SET total_reserved = $1, version = version + 1
			WHERE hotel_id = $2 AND room_type_id = $3 AND date = ANY($4)
		`, TestInventoryReserved, hotel.ID, roomType.ID, dates)
		require.NoError(t, err)

		url := fmt.Sprintf("/reservation/availability?hotelId=%s&checkIn=%s&checkOut=%s",
			hotel.ID.String(),
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))

		resp, err := reservationSuite.MakeAuthenticatedRequest("GET", url, nil, user)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var availResp GetAvailabilityResponse
		err = json.NewDecoder(resp.Body).Decode(&availResp)
		require.NoError(t, err)
		assert.NotEmpty(t, availResp.Availability)

		expectedDays := int(endDate.Sub(startDate).Hours()/24) + 1
		assert.Len(t, availResp.Availability, expectedDays)
	})

	t.Run("should_exclude_rooms_at_capacity", func(t *testing.T) {
		t.Parallel()

		hotel, err := reservationSuite.CreateTestHotel()
		require.NoError(t, err)

		roomType, err := reservationSuite.CreateTestRoomType(hotel.ID)
		require.NoError(t, err)

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		startDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
		endDate := time.Now().AddDate(0, 0, 2).Truncate(24 * time.Hour)

		ctx := context.Background()
		_, err = reservationSuite.GetQueries().BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
			HotelID:        hotel.ID,
			RoomTypeID:     roomType.ID,
			Dates:          []time.Time{startDate, endDate},
			TotalInventory: TestInventoryMax,
		})
		require.NoError(t, err)
		_, err = reservationSuite.GetDB().GetDB().ExecContext(ctx, `
			UPDATE booking.room_type_inventory 
			SET total_reserved = $1
			WHERE hotel_id = $2 AND room_type_id = $3 AND date IN ($4, $5)
		`, TestInventoryOverbooked, hotel.ID, roomType.ID, startDate, endDate)
		require.NoError(t, err)

		url := fmt.Sprintf("/reservation/availability?hotelId=%s&checkIn=%s&checkOut=%s",
			hotel.ID.String(),
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))

		resp, err := reservationSuite.MakeAuthenticatedRequest("GET", url, nil, user)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var availResp GetAvailabilityResponse
		err = json.NewDecoder(resp.Body).Decode(&availResp)
		require.NoError(t, err)
		assert.Empty(t, availResp.Availability, "Should return no availability when at capacity")
	})
}

func TestReservationIdGeneration(t *testing.T) {
	t.Parallel()

	t.Run("should_generate_unique_ids", func(t *testing.T) {
		t.Parallel()

		user, err := reservationSuite.CreateTestUser()
		require.NoError(t, err)

		ids := make(map[string]bool)

		for range 10 {
			resp, err := reservationSuite.MakeAuthenticatedRequest("GET", "/reservation/generate-id", nil, user)
			require.NoError(t, err)

			var idResp GenerateIdResponse
			err = json.NewDecoder(resp.Body).Decode(&idResp)
			require.NoError(t, err)
			resp.Body.Close()

			id := idResp.ReservationId
			assert.NotEmpty(t, id)
			assert.False(t, ids[id], "Generated duplicate ID: %s", id)

			_, err = uuid.Parse(id)
			assert.NoError(t, err, "Generated ID should be valid UUID: %s", id)

			ids[id] = true
		}
	})
}
