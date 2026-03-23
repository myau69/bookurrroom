package models

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrScheduleInvalidID          = errors.New("schedules: invalid schedule id")
	ErrScheduleInvalidRoomID      = errors.New("schedules: invalid room id")
	ErrScheduleEmptyDaysOfWeek    = errors.New("schedules: daysOfWeek is required")
	ErrScheduleInvalidDayOfWeek   = errors.New("schedules: dayOfWeek mist be in [1..7]")
	ErrScheduleDuplicateDayOfWeek = errors.New("schedules: duplicate dayOfWeek")
	ErrScheduleInvalidTimeFormat  = errors.New("schedules: invalid time format, expected HH:MM")
	ErrScheduleInvalideTimeRange  = errors.New("schedules: startTime must be before endTime")
)

type Schedule struct {
	ID           uuid.UUID
	RoomID       uuid.UUID
	DaysOfWeek   []int
	StartTime    string
	EndTime      string
	CreatedAtUTC time.Time
}

func NewSchedule(id, roomID uuid.UUID, days []int, startTime, endTime string, now time.Time) (Schedule, error) {
	s := Schedule{
		ID:           id,
		RoomID:       roomID,
		DaysOfWeek:   append([]int(nil), days...),
		StartTime:    strings.TrimSpace(startTime),
		EndTime:      strings.TrimSpace(endTime),
		CreatedAtUTC: now.UTC(),
	}
	if err := s.ValidateForCreate(); err != nil {
		return Schedule{}, err
	}
	s.DaysOfWeek = normalizeDays(s.DaysOfWeek)
	return s, nil
}

func (s Schedule) ValidateForCreate() error {
	if s.ID == uuid.Nil {
		return ErrScheduleInvalidID
	}
	if s.RoomID == uuid.Nil {
		return ErrScheduleInvalidRoomID
	}
	if len(s.DaysOfWeek) == 0 {
		return ErrScheduleEmptyDaysOfWeek
	}

	seen := make(map[int]struct{}, len(s.DaysOfWeek))
	for _, d := range s.DaysOfWeek {
		if d < 1 || d > 7 {
			return ErrScheduleInvalidDayOfWeek
		}
		if _, ok := seen[d]; ok {
			return ErrScheduleDuplicateDayOfWeek
		}
		seen[d] = struct{}{}
	}

	start, err := parseScheduleHHMM(s.StartTime)
	if err != nil {
		return err
	}
	end, err := parseScheduleHHMM(s.EndTime)
	if err != nil {
		return err
	}
	if start >= end {
		return ErrScheduleInvalideTimeRange
	}

	return nil
}

func parseScheduleHHMM(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, ErrScheduleInvalidTimeFormat
	}
	hh, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, ErrScheduleInvalidTimeFormat
	}
	mm, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, ErrScheduleInvalidTimeFormat
	}
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, ErrScheduleInvalidTimeFormat
	}
	return time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute, nil
}

func normalizeDays(days []int) []int {
	out := append([]int(nil), days...)
	sort.Ints(out)
	return out
}

func (s Schedule) TimeBounds() (time.Duration, time.Duration, error) {
	start, err := parseScheduleHHMM(s.StartTime)
	if err != nil {
		return 0, 0, fmt.Errorf("parse start: %w", err)
	}
	end, err := parseScheduleHHMM(s.EndTime)
	if err != nil {
		return 0, 0, fmt.Errorf("parse end: %w", err)
	}
	return start, end, nil
}
