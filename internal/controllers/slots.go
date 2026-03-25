package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"bookurrroom/internal/models"
	"bookurrroom/internal/services"
)

type SlotsHandler struct {
	svc *services.SlotsService
}

type slotResponse struct {
	ID     uuid.UUID `json:"id"`
	RoomID uuid.UUID `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

func NewSlotsHandler(svc *services.SlotsService) *SlotsHandler {
	return &SlotsHandler{svc: svc}
}

// @Summary List available room slots by date
// @Tags slots
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room ID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} slotsListResponse
// @Failure 400 {object} errorPayload
// @Failure 401 {object} errorPayload
// @Failure 403 {object} errorPayload
// @Failure 404 {object} errorPayload
// @Failure 500 {object} errorPayload
// @Router /rooms/{roomId}/slots/list [get]
func (h *SlotsHandler) ListByRoomAndDate(w http.ResponseWriter, r *http.Request) {
	roomID, err := parsePathUUID(chi.URLParam(r, "roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	dateRaw := r.URL.Query().Get("date")
	if dateRaw == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date is required")
		return
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, expected YYYY-MM-DD")
		return
	}

	slots, err := h.svc.ListAvailable(r.Context(), roomID, date.UTC())
	if err != nil {
		if errors.Is(err, services.ErrSlotsRoomNotFound) {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	out := make([]slotResponse, 0, len(slots))
	for _, slot := range slots {
		out = append(out, mapSlot(slot))
	}

	writeJSON(w, http.StatusOK, map[string][]slotResponse{
		"slots": out,
	})
}

func mapSlot(slot models.Slot) slotResponse {
	return slotResponse{
		ID:     slot.ID,
		RoomID: slot.RoomID,
		Start:  slot.StartUTC.UTC(),
		End:    slot.EndUTC.UTC(),
	}
}
