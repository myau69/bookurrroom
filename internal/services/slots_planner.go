package services

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

var ErrSlotsPlannerInvalidTime = errors.New("slots: planner invalid time")

type SlotsPlanner interface {
	BuildForDate(roomID uuid.UUID, rule SlotsScheduleRule, date time.Time) ([]models.Slot, error)
	BuildWindow(roomID uuid.UUID, rule SlotsScheduleRule, from, to time.Time) ([]models.Slot, error)
}

type DefaultSlotsPlanner struct{}

func NewDefaultSlotsPlanner() *DefaultSlotsPlanner {
	return &DefaultSlotsPlanner{}
}

func (p *DefaultSlotsPlanner) BuildForDate(roomID uuid.UUID, rule SlotsScheduleRule, date time.Time) ([]models.Slot, error) {
	dayStart := time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	isoDay := weekdayISO(dayStart.Weekday())
	if containsDay(rule.DaysOfWeek, isoDay) == false {
		return []models.Slot{}, nil
	}

	startOffset, err := parseSlotsHHMM(rule.StartTime)
	if err != nil {
		return nil, err
	}
	endOffset, err := parseSlotsHHMM(rule.EndTime)
	if err != nil {
		return nil, err
	}
	if endOffset <= startOffset {
		return nil, ErrSlotsPlannerInvalidTime
	}

	slots := make([]models.Slot, 0, 16)
	cursor := dayStart.Add(startOffset)
	windowEnd := dayStart.Add(endOffset)
	for cursor.Add(models.SlotDuration).After(windowEnd) == false {
		s, err := models.NewSlot(uuid.New(), roomID, cursor, cursor.Add(models.SlotDuration))
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
		cursor = cursor.Add(models.SlotDuration)
	}
	return slots, nil
}

func (p *DefaultSlotsPlanner) BuildWindow(roomID uuid.UUID, rule SlotsScheduleRule, from, to time.Time) ([]models.Slot, error) {
	start := time.Date(from.UTC().Year(), from.UTC().Month(), from.UTC().Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(to.UTC().Year(), to.UTC().Month(), to.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if end.Before(start) {
		return []models.Slot{}, nil
	}

	all := make([]models.Slot, 0, 64)
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		daily, err := p.BuildForDate(roomID, rule, d)
		if err != nil {
			return nil, err
		}
		all = append(all, daily...)
	}
	return all, nil
}

func parseSlotsHHMM(v string) (time.Duration, error) {
	parts := strings.Split(strings.TrimSpace(v), ":")
	if len(parts) != 2 {
		return 0, ErrSlotsPlannerInvalidTime
	}
	hh, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, ErrSlotsPlannerInvalidTime
	}
	mm, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, ErrSlotsPlannerInvalidTime
	}
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, ErrSlotsPlannerInvalidTime
	}
	return time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute, nil
}

func containsDay(days []int, needle int) bool {
	copyDays := append([]int(nil), days...)
	sort.Ints(copyDays)
	idx := sort.SearchInts(copyDays, needle)
	return idx < len(copyDays) && copyDays[idx] == needle
}

func weekdayISO(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}
