package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

type schedulesRepoMock struct {
	createFn      func(ctx context.Context, schedule models.Schedule) (models.Schedule, error)
	getByRoomIDFn func(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error)
}

var _ repository.SchedulesRepository = (*schedulesRepoMock)(nil)

func (m *schedulesRepoMock) Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	if m.createFn == nil {
		return schedule, nil
	}
	return m.createFn(ctx, schedule)
}

func (m *schedulesRepoMock) GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error) {
	if m.getByRoomIDFn == nil {
		return models.Schedule{}, false, nil
	}
	return m.getByRoomIDFn(ctx, roomID)
}

type schedulesRoomCheckerMock struct {
	existsFn func(ctx context.Context, roomID uuid.UUID) (bool, error)
}

func (m schedulesRoomCheckerMock) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	if m.existsFn == nil {
		return true, nil
	}
	return m.existsFn(ctx, roomID)
}

func TestSchedulesServiceCreateSuccess(t *testing.T) {
	svc := NewSchedulesService(&schedulesRepoMock{}, schedulesRoomCheckerMock{})

	schedule, err := svc.Create(context.Background(), SchedulesCreateInput{
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, schedule.DaysOfWeek)
}

func TestSchedulesServiceCreateRoomNotFound(t *testing.T) {
	svc := NewSchedulesService(&schedulesRepoMock{}, schedulesRoomCheckerMock{
		existsFn: func(ctx context.Context, roomID uuid.UUID) (bool, error) {
			return false, nil
		},
	})

	_, err := svc.Create(context.Background(), SchedulesCreateInput{RoomID: uuid.New(), DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00"})
	require.ErrorIs(t, err, ErrSchedulesRoomNotFound)
}

func TestSchedulesServiceCreateAlreadyExists(t *testing.T) {
	svc := NewSchedulesService(&schedulesRepoMock{
		getByRoomIDFn: func(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error) {
			return models.Schedule{ID: uuid.New()}, true, nil
		},
	}, schedulesRoomCheckerMock{})

	_, err := svc.Create(context.Background(), SchedulesCreateInput{RoomID: uuid.New(), DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00"})
	require.ErrorIs(t, err, ErrSchedulesAlreadyExists)
}

func TestSchedulesServiceCreateRoomCheckerError(t *testing.T) {
	repoErr := errors.New("checker fail")
	svc := NewSchedulesService(&schedulesRepoMock{}, schedulesRoomCheckerMock{
		existsFn: func(ctx context.Context, roomID uuid.UUID) (bool, error) {
			return false, repoErr
		},
	})

	_, err := svc.Create(context.Background(), SchedulesCreateInput{RoomID: uuid.New(), DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00"})
	require.ErrorIs(t, err, repoErr)
}
