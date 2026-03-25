package services

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSchedulesRoomNotFound  = errors.New("schedules: room not found")
	ErrSchedulesAlreadyExists = errors.New("schedules: schedule already exists")
)

type SchedulesRoomChecker interface {
	Exists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type SchedulesCreateInput struct {
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

type SchedulesService struct {
	repo  repository.SchedulesRepository
	rooms SchedulesRoomChecker
	now   func() time.Time
}

func NewSchedulesService(repo repository.SchedulesRepository, rooms SchedulesRoomChecker) *SchedulesService {
	return &SchedulesService{repo: repo, rooms: rooms, now: time.Now}
}

func (s *SchedulesService) Create(ctx context.Context, in SchedulesCreateInput) (models.Schedule, error) {
	exists, err := s.rooms.Exists(ctx, in.RoomID)
	if err != nil {
		return models.Schedule{}, err
	}
	if exists == false {
		return models.Schedule{}, ErrSchedulesRoomNotFound
	}

	_, found, err := s.repo.GetByRoomID(ctx, in.RoomID)
	if err != nil {
		return models.Schedule{}, err
	}
	if found {
		return models.Schedule{}, ErrSchedulesAlreadyExists
	}

	schedule, err := models.NewSchedule(uuid.New(), in.RoomID, in.DaysOfWeek, in.StartTime, in.EndTime, s.now())
	if err != nil {
		return models.Schedule{}, err
	}

	return s.repo.Create(ctx, schedule)
}

func (s *SchedulesService) GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error) {
	return s.repo.GetByRoomID(ctx, roomID)
}

func (s *SchedulesService) GetRuleByRoomID(ctx context.Context, roomID uuid.UUID) (SlotsScheduleRule, bool, error) {
	schedule, found, err := s.repo.GetByRoomID(ctx, roomID)
	if err != nil {
		return SlotsScheduleRule{}, false, err
	}
	if found == false {
		return SlotsScheduleRule{}, false, nil
	}

	return SlotsScheduleRule{
		DaysOfWeek: append([]int(nil), schedule.DaysOfWeek...),
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}, true, nil
}
