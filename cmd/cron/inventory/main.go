package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/config"
	"github.com/AlexKhomenko00/hotel-system/internal/database"
	"github.com/AlexKhomenko00/hotel-system/internal/hotel"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	validator := validator.New()
	cfg := config.GetConfig(validator)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Create(cfg)
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}
	queries := database.New(db.GetDB())

	hotelSvc := hotel.New(queries, validator)

	/* More complex partition based job processing should be applied.
	Probably hotels may be partitioned by country or timezones so CRON schedule should account for it.
	For presentation purpose that we don't have thousands of hotels and they all in the same time zone so we can safely CRON once per day.
	**/
	hotels, err := hotelSvc.GetActiveHotels(ctx)
	if err != nil {
		log.Fatalf("Failed to retrieve active hotels: %v", err)
	}

	if len(hotels) == 0 {
		log.Fatal("No active hotels found!")
	}

	slog.Info("Starting inventory population", "hotel_count", len(hotels))

	var wg sync.WaitGroup

	wg.Add(len(hotels))
	for _, hotel := range hotels {
		go func() {
			defer wg.Done()
			err := fetchAndPopulateHotelInventory(ctx, queries, hotel)
			if err != nil {
				slog.Error("Failed to update inventory for hotel!",
					"hotel", hotel.ID,
					"error", err)
				return
			}

			slog.Info("Successfully updated hotel", "hotel_id", hotel.ID)

		}()
	}

	wg.Wait()
}

func fetchAndPopulateHotelInventory(ctx context.Context, queries *database.Queries, hotel database.BookingHotel) error {
	today := time.Now().Truncate(24 * time.Hour)
	windowEnd := today.AddDate(2, 0, 0) // 2 years ahead

	capacityPlans, err := getHotelInventoryFromTo(ctx, queries, hotel.ID, today, windowEnd)
	if err != nil {
		return fmt.Errorf("failed to get capacity plans: %w", err)
	}

	if len(capacityPlans) == 0 {
		return fmt.Errorf("capacity plans not found for hotel %s", hotel.ID)
	}

	var wg sync.WaitGroup

	wg.Add(len(capacityPlans))
	for rTypeId, capacity := range capacityPlans {
		go func() {
			defer wg.Done()

			var dates []time.Time
			for d := capacity.fromDate; d.Before(capacity.toDate.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
				dates = append(dates, d)
			}

			_, err := queries.BatchUpdateRoomTypeInventory(ctx, database.BatchUpdateRoomTypeInventoryParams{
				Dates:          dates,
				HotelID:        hotel.ID,
				RoomTypeID:     rTypeId,
				TotalInventory: capacity.totalInventory,
			})

			if err != nil {
				slog.Error("failed to insert capacity for room",
					"room_type_id", rTypeId,
					"error", err.Error())
				return
			}

			slog.Info("successfully updated room type inventory",
				"hotel_id", hotel.ID,
				"room_type_id", rTypeId,
				"date_count", len(dates))

		}()
	}

	wg.Wait()

	return nil
}

type RoomCapacityPerType map[uuid.UUID]PlannedRoomTypeCapacity
type PlannedRoomTypeCapacity struct {
	fromDate       time.Time
	toDate         time.Time
	totalInventory int32
}

// Mocked approach for inventory management. Usually calculated and provided by business
func getHotelInventoryFromTo(ctx context.Context, queries *database.Queries, hotelID uuid.UUID, dateFrom, dateTo time.Time) (RoomCapacityPerType, error) {
	roomTypes, err := queries.GetHotelUniqueRoomTypes(ctx, hotelID)
	if err != nil {
		return nil, err
	}

	capacity := make(RoomCapacityPerType)
	for _, rTypeID := range roomTypes {
		capacity[rTypeID] = PlannedRoomTypeCapacity{
			fromDate:       dateFrom,
			toDate:         dateFrom.AddDate(2, 0, 0), // 2 years
			totalInventory: 30,
		}
	}

	return capacity, nil
}

func maxTime(date_1, date_2 time.Time) time.Time {
	if date_1.After(date_2) {
		return date_1
	}
	return date_2
}

func minTime(date_1, date_2 time.Time) time.Time {
	if date_1.Before(date_2) {
		return date_1
	}

	return date_2
}
