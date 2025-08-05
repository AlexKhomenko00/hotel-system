package reservation

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AlexKhomenko00/hotel-system/internal/auth"
	"github.com/AlexKhomenko00/hotel-system/internal/shared"
	"github.com/go-chi/chi/v5"
)

func (s *ReservationService) RegisterHandlers(r chi.Router) {
	r.Get("/generate-id", s.GenerateReservationIdHandler)
	r.Post("/", s.MakeReservationHandler)
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

	err := s.makeReservation(r.Context(), usr.Id, body)
	if err != nil {
		if err.Error() == ErrInventoryCapacityReached.Error() {
			shared.WriteError(w, http.StatusConflict, "No availability for selected dates")
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "Sorry, something went wrong")
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Envelope{"message": "Reservation created successfully", "reservation_id": body.ReservationId})
}
