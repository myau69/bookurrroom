package controllers

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"bookurrroom/internal/services"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SchedulesHandler struct {
	svc      *services.SchedulesService
	slots    repository.SlotsRepository
	planner  services.SlotsPlanner
	now      func() time.Time
	windowTo func(time.Time) time.Time
}

type createScheduleRequest struct {
	RoomID     *uuid.UUID `json:"roomId,omitempty"`
	DaysOfWeek []int      `json:"daysOfWeek"`
	StartTime  string     `json:"startTime"`
	EndTime    string     `json:"endTime"`
}

type scheduleResponse struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

func NewSchedulesHandler(svc *services.SchedulesService, slots repository.SlotsRepository, planner services.SlotsPlanner) *SchedulesHandler {
	return &SchedulesHandler{
		svc:      svc,
		slots:    slots,
		planner:  planner,
		now:      func() time.Time { return time.Now().UTC() },
		windowTo: func(from time.Time) time.Time { return from.AddDate(0, 0, 6) },
	}
}

func (h *SchedulesHandler) Create(w http.ResponseWriter, r *http.Request) {
	roomID, err := parsePathUUID(chi.URLParam(r, "roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	var req createScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if req.RoomID == nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "roomId is required")
		return
	}
	if *req.RoomID != roomID {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "roomId in path and body must match")
		return
	}

	schedule, err := h.svc.Create(r.Context(), services.SchedulesCreateInput{
		RoomID:     roomID,
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSchedulesRoomNotFound):
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
			return
		case errors.Is(err, services.ErrSchedulesAlreadyExists):
			writeError(w, http.StatusConflict, "SCHEDULE_EXISTS", err.Error())
			return
		case isScheduleValidationError(err):
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
	}

	from := h.now()
	to := h.windowTo(from)
	generatedSlots, err := h.planner.BuildWindow(schedule.RoomID, services.SlotsScheduleRule{
		DaysOfWeek: schedule.DaysOfWeek,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if err := h.slots.UpsertMany(r.Context(), generatedSlots); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]scheduleResponse{
		"schedule": mapSchedule(schedule),
	})
}

func isScheduleValidationError(err error) bool {
	return errors.Is(err, models.ErrScheduleInvalidID) ||
		errors.Is(err, models.ErrScheduleInvalidRoomID) ||
		errors.Is(err, models.ErrScheduleEmptyDaysOfWeek) ||
		errors.Is(err, models.ErrScheduleInvalidDayOfWeek) ||
		errors.Is(err, models.ErrScheduleDuplicateDayOfWeek) ||
		errors.Is(err, models.ErrScheduleInvalidTimeFormat) ||
		errors.Is(err, models.ErrScheduleInvalideTimeRange)
}

func mapSchedule(schedule models.Schedule) scheduleResponse {
	return scheduleResponse{
		ID:         schedule.ID,
		RoomID:     schedule.RoomID,
		DaysOfWeek: append([]int(nil), schedule.DaysOfWeek...),
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}
}
