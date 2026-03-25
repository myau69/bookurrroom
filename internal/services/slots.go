package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

var ErrSlotsRoomNotFound = errors.New("slots: room not found")

type SlotsRoomChecker interface {
	Exists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type SlotsScheduleRule struct {
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

type SlotsScheduleReader interface {
	GetRuleByRoomID(ctx context.Context, roomID uuid.UUID) (SlotsScheduleRule, bool, error)
}

type SlotsService struct {
	rooms     SlotsRoomChecker
	schedules SlotsScheduleReader
	repo      repository.SlotsRepository
}

func NewSlotsService(rooms SlotsRoomChecker, schedules SlotsScheduleReader, repo repository.SlotsRepository) *SlotsService {
	return &SlotsService{rooms: rooms, schedules: schedules, repo: repo}
}

func (s *SlotsService) ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	exists, err := s.rooms.Exists(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if exists == false {
		return nil, ErrSlotsRoomNotFound
	}

	_, found, err := s.schedules.GetRuleByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if found == false {
		return []models.Slot{}, nil
	}

	return s.repo.ListFreeByRoomAndDate(ctx, roomID, date.UTC())
}

func (s *SlotsService) GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
	return s.repo.GetByID(ctx, slotID)
}
