package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
	"bookurrroom/internal/services"
)

type RoomsHandler struct {
	svc *services.RoomsService
}

type createRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

type roomResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Capacity    *int       `json:"capacity"`
	CreatedAt   *time.Time `json:"createdAt"`
}

func NewRoomsHandler(svc *services.RoomsService) *RoomsHandler {
	return &RoomsHandler{svc: svc}
}

// @Summary Create room
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createRoomRequest true "room payload"
// @Success 201 {object} roomCreateResponse
// @Failure 400 {object} errorPayload
// @Failure 401 {object} errorPayload
// @Failure 403 {object} errorPayload
// @Failure 500 {object} errorPayload
// @Router /rooms/create [post]
func (h *RoomsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	room, err := h.svc.Create(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		if isRoomValidationError(err) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]roomResponse{
		"room": mapRoom(room),
	})
}

// @Summary List rooms
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Success 200 {object} roomsListResponse
// @Failure 401 {object} errorPayload
// @Failure 403 {object} errorPayload
// @Failure 500 {object} errorPayload
// @Router /rooms/list [get]
func (h *RoomsHandler) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.svc.GetList(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	out := make([]roomResponse, 0, len(rooms))
	for _, room := range rooms {
		out = append(out, mapRoom(room))
	}

	writeJSON(w, http.StatusOK, map[string][]roomResponse{
		"rooms": out,
	})
}

func isRoomValidationError(err error) bool {
	return errors.Is(err, models.ErrRoomInvalidID) ||
		errors.Is(err, models.ErrRoomNameRequired) ||
		errors.Is(err, models.ErrRoomInvalidCapacity)
}

func mapRoom(room models.Room) roomResponse {
	var createdAt *time.Time
	if room.CreatedAtUtc.IsZero() == false {
		v := room.CreatedAtUtc.UTC()
		createdAt = &v
	}

	return roomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
		CreatedAt:   createdAt,
	}
}
