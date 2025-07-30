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

func (s *HotelService) AddRoomTypeHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := chi.URLParam(r, "hotelId")
	hotelID, err := uuid.Parse(hotelIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	var body AddRoomType
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body %v", err))
		return
	}

	body.HotelID = hotelID.String()

	if err := s.validator.Struct(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body: %v", err.Error()))
		return
	}

	roomType, err := s.addRoomType(r.Context(), body)
	if err != nil {
		if errors.Is(err, ErrDuplicateRoomByName) {
			shared.WriteError(w, http.StatusConflict, "Room type with such name already exists")
			return
		}
		shared.WriteError(w, http.StatusInternalServerError, "Failed to add room type")
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Envelope{"roomType": roomType})
}

func (s *HotelService) GetRoomTypeHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := chi.URLParam(r, "hotelId")
	hotelID, err := uuid.Parse(hotelIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid room type ID")
		return
	}

	roomType, err := s.getRoomTypeById(r.Context(), hotelID, roomID)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, "Room type not found")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Envelope{"roomType": roomType})
}

func (s *HotelService) UpdateRoomTypeHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := chi.URLParam(r, "hotelId")
	hotelID, err := uuid.Parse(hotelIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid room type ID")
		return
	}

	var body UpdateRoomType
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body %v", err))
		return
	}

	if err := s.validator.Struct(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid body: %v", err.Error()))
		return
	}

	roomType, err := s.updateRoomType(r.Context(), hotelID, roomID, body)
	if err != nil {
		if errors.Is(err, ErrDuplicateRoomByName) {
			shared.WriteError(w, http.StatusConflict, "Room type with such name already exists")
			return
		}
		shared.WriteError(w, http.StatusNotFound, "Room type not found")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Envelope{"roomType": roomType})
}

func (s *HotelService) DeleteRoomTypeHandler(w http.ResponseWriter, r *http.Request) {
	hotelIDStr := chi.URLParam(r, "hotelId")
	hotelID, err := uuid.Parse(hotelIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid hotel ID")
		return
	}

	roomIDStr := chi.URLParam(r, "id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "Invalid room type ID")
		return
	}

	err = s.deleteRoomType(r.Context(), hotelID, roomID)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, "Room type not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
