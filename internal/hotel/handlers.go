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
	resEncoder := json.NewEncoder(w)

	var body CreateHotelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: fmt.Sprintf("Invalid body %v", err),
		})
		return
	}

	if err := s.validator.Struct(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: fmt.Sprintf("Invalid body: %v", err.Error()),
		})
		return
	}

	hotel, err := s.createHotel(r.Context(), body)
	if err != nil {
		if errors.Is(err, ErrDuplicateHotelByName) {
			w.WriteHeader(http.StatusBadRequest)
			resEncoder.Encode(shared.ErrorRes{
				Message: ErrDuplicateHotelByName.Error(),
			})
			return
		}

		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resEncoder.Encode(hotel)
}

func (s *HotelService) GetHotelHandler(w http.ResponseWriter, r *http.Request) {
	resEncoder := json.NewEncoder(w)
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Invalid hotel ID",
		})
		return
	}

	hotel, err := s.getHotelById(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Hotel not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	resEncoder.Encode(hotel)
}

func (s *HotelService) UpdateHotelHandler(w http.ResponseWriter, r *http.Request) {
	resEncoder := json.NewEncoder(w)
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Invalid hotel ID",
		})
		return
	}

	var body UpdateHotelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: fmt.Sprintf("Invalid body %v", err),
		})
		return
	}

	if err := s.validator.Struct(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: fmt.Sprintf("Invalid body: %v", err.Error()),
		})
		return
	}

	hotel, err := s.updateHotel(r.Context(), id, body)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Hotel not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	resEncoder.Encode(hotel)
}

func (s *HotelService) DeleteHotelHandler(w http.ResponseWriter, r *http.Request) {
	resEncoder := json.NewEncoder(w)
	
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Invalid hotel ID",
		})
		return
	}

	err = s.deleteHotel(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resEncoder.Encode(shared.ErrorRes{
			Message: "Hotel not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
