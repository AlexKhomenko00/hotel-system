package reservation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AlexKhomenko00/hotel-system/internal/auth"
	"github.com/AlexKhomenko00/hotel-system/internal/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *ReservationService) RegisterHandlers(r chi.Router) {
	r.Get("/generate-id", s.GenerateReservationIdHandler)
	r.Post("/", s.MakeReservationHandler)
	r.Get("/availability", s.GetRoomAvailabilityHandler)
}

func (s *ReservationService) GenerateReservationIdHandler(w http.ResponseWriter, r *http.Request) {
	reservationId := s.generateReservationId()

	shared.WriteJSON(w, http.StatusOK, shared.Envelope{"reservation_id": reservationId})
}

func (s *ReservationService) MakeReservationHandler(w http.ResponseWriter, r *http.Request) {
	usr, ok := r.Context().Value(auth.UsrCtxKey).(auth.UserContext)
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "Guest ID not found in context")
		return
	}

	var body MakeReservationBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body %v", err))
		return
	}

	if err := s.validator.Struct(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body: %v", err.Error()))
		return
	}

	err := s.makeReservation(r.Context(), usr.GuestId, body)
	if err != nil {
		if errors.Is(err, ErrInventoryCapacityReached) {
			shared.WriteError(w, http.StatusConflict, "No availability for selected dates")
			return
		}
		if errors.Is(err, ErrInvalidReservationID) {
			shared.WriteError(w, http.StatusBadRequest, "Invalid reservation ID")
			return
		}
		if errors.Is(err, ErrOptimisticLockMismatch) {
			shared.WriteError(w, http.StatusConflict, "Reservation conflict - please try again")
			return
		}
		if errors.Is(err, ErrDuplicateReservation) {
			shared.WriteError(w, http.StatusConflict, "Reservation ID already exists")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "Sorry, something went wrong")
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Envelope{"message": "Reservation created successfully", "reservation_id": body.ReservationId})
}
func (s *ReservationService) GetRoomAvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := r.URL.Query().Get("hotelId")
	checkInStr := r.URL.Query().Get("checkIn")
	checkOutStr := r.URL.Query().Get("checkOut")

	hotelID, err := uuid.Parse(hotelIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotelId")
		return
	}

	checkIn, err := time.Parse(time.DateOnly, checkInStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid checkIn")
		return
	}

	checkOut, err := time.Parse(time.DateOnly, checkOutStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid checkOut")
		return
	}

	if !checkOut.After(checkIn) {
		shared.WriteError(w, http.StatusBadRequest, "checkOut should be after checkIn date")
		return
	}

	availability, err := s.getRoomAvailability(r.Context(), GetAvailabilityQuery{
		HotelID:  hotelID,
		CheckIn:  checkIn,
		CheckOut: checkOut,
	})

	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "Failed to get room availability")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Envelope{"availability": availability})
}
