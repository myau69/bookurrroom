package controllers

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/services"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookingsHandler struct {
	svc *services.BookingsService
	now func() time.Time
}

type createBookingRequest struct {
	SlotID               uuid.UUID `json:"slotId"`
	CreateConferenceLink bool      `json:"createConferenceLink"`
}

type bookingResponse struct {
	ID             uuid.UUID  `json:"id"`
	SlotID         uuid.UUID  `json:"slotId"`
	UserID         uuid.UUID  `json:"userId"`
	Status         string     `json:"status"`
	ConferenceLink *string    `json:"conferenceLink"`
	CreatedAt      *time.Time `json:"createdAt"`
}

type paginationResponse struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func NewBookingsHandler(svc *services.BookingsService) *BookingsHandler {
	return &BookingsHandler{
		svc: svc,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (h *BookingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if req.SlotID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "slotId is required")
		return
	}

	principal, ok := principalFromContext(r.Context())
	if ok == false {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	booking, err := h.svc.Create(r.Context(), principal.UserID, req.SlotID, req.CreateConferenceLink)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrBookingsSlotNotFound):
			writeError(w, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
			return
		case errors.Is(err, services.ErrBookingsSlotAlreadyBooked):
			writeError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", err.Error())
			return
		case errors.Is(err, services.ErrBookingsSlotInPast):
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]bookingResponse{
		"booking": mapBooking(booking),
	})
}

func (h *BookingsHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	page, err := optionalPositiveInt(r.URL.Query().Get("page"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid page")
		return
	}

	pageSize, err := optionalPositiveInt(r.URL.Query().Get("pageSize"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid pageSize")
		return
	}

	items, pagination, err := h.svc.ListAll(r.Context(), page, pageSize)
	if err != nil {
		if errors.Is(err, services.ErrBookingsInvalidPagination) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	out := make([]bookingResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapBooking(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bookings": out,
		"pagination": paginationResponse{
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
			Total:    pagination.Total,
		},
	})
}

func (h *BookingsHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if ok == false {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	items, err := h.svc.ListMy(r.Context(), principal.UserID, h.now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	out := make([]bookingResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapBooking(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bookings": out,
	})
}

func (h *BookingsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := parsePathUUID(chi.URLParam(r, "bookingId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId")
		return
	}

	principal, ok := principalFromContext(r.Context())
	if ok == false {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	booking, err := h.svc.Cancel(r.Context(), principal.UserID, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrBookingsNotFound):
			writeError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", err.Error())
			return
		case errors.Is(err, services.ErrBookingsForbidden):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]bookingResponse{
		"booking": mapBooking(booking),
	})
}

func mapBooking(booking models.Booking) bookingResponse {
	var createdAt *time.Time
	if booking.CreatedatUTC.IsZero() == false {
		v := booking.CreatedatUTC.UTC()
		createdAt = &v
	}

	return bookingResponse{
		ID:             booking.ID,
		SlotID:         booking.SlotID,
		UserID:         booking.UserID,
		Status:         string(booking.Status),
		ConferenceLink: booking.ConferenceLink,
		CreatedAt:      createdAt,
	}
}

func optionalPositiveInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, errors.New("must be positive")
	}
	return value, nil
}
