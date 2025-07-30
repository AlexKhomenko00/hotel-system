package hotel

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlexKhomenko00/hotel-system/internal/shared"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *HotelService) CreateHotelHandler(w http.ResponseWriter, r *http.Request) {
	var body CreateHotelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body %v", err))
		return
	}

	if err := s.validator.Struct(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body: %v", err.Error()))
		return
	}

	hotel, err := s.createHotel(r.Context(), body)
	if err != nil {
		if errors.Is(err, ErrDuplicateHotelByName) {
			shared.WriteError(w, http.StatusBadRequest, ErrDuplicateHotelByName.Error())
			return
		}

		shared.WriteError(w, http.StatusInternalServerError, "Sorry, something went wrong")
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Envelope{"hotel": hotel})
}

func (s *HotelService) GetHotelHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	hotel, err := s.getHotelById(r.Context(), id)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, "Hotel not found")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Envelope{"hotel": hotel})
}

func (s *HotelService) UpdateHotelHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	var body UpdateHotelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body %v", err))
		return
	}

	if err := s.validator.Struct(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body: %v", err.Error()))
		return
	}

	hotel, err := s.updateHotel(r.Context(), id, body)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, "Hotel not found")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Envelope{"hotel": hotel})
}

func (s *HotelService) DeleteHotelHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	err = s.deleteHotel(r.Context(), id)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, "Hotel not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
