package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

type slotsRepoMock struct {
	listFreeFn func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error)
	getByIDFn  func(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error)
}

var _ repository.SlotsRepository = (*slotsRepoMock)(nil)

func (m *slotsRepoMock) UpsertMany(ctx context.Context, slots []models.Slot) error {
	return nil
}

func (m *slotsRepoMock) ListFreeByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	if m.listFreeFn == nil {
		return []models.Slot{}, nil
	}
	return m.listFreeFn(ctx, roomID, date)
}

func (m *slotsRepoMock) GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
	if m.getByIDFn == nil {
		return models.Slot{}, false, nil
	}
	return m.getByIDFn(ctx, slotID)
}

type slotsRoomCheckerMock struct {
	existsFn func(ctx context.Context, roomID uuid.UUID) (bool, error)
}

func (m slotsRoomCheckerMock) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	if m.existsFn == nil {
		return true, nil
	}
	return m.existsFn(ctx, roomID)
}

type slotsScheduleReaderMock struct {
	getFn func(ctx context.Context, roomID uuid.UUID) (SlotsScheduleRule, bool, error)
}

func (m slotsScheduleReaderMock) GetRuleByRoomID(ctx context.Context, roomID uuid.UUID) (SlotsScheduleRule, bool, error) {
	if m.getFn == nil {
		return SlotsScheduleRule{}, true, nil
	}
	return m.getFn(ctx, roomID)
}

func TestSlotsServiceListAvailableSuccess(t *testing.T) {
	now := time.Now().UTC()
	expected := []models.Slot{{ID: uuid.New(), RoomID: uuid.New(), StartUTC: now, EndUTC: now.Add(models.SlotDuration)}}
	svc := NewSlotsService(
		slotsRoomCheckerMock{},
		slotsScheduleReaderMock{},
		&slotsRepoMock{listFreeFn: func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
			return expected, nil
		}},
	)

	actual, err := svc.ListAvailable(context.Background(), uuid.New(), now)
	require.NoError(t, err)
	require.Len(t, actual, 1)
}

func TestSlotsServiceListAvailableRoomNotFound(t *testing.T) {
	svc := NewSlotsService(
		slotsRoomCheckerMock{existsFn: func(ctx context.Context, roomID uuid.UUID) (bool, error) { return false, nil }},
		slotsScheduleReaderMock{},
		&slotsRepoMock{},
	)

	_, err := svc.ListAvailable(context.Background(), uuid.New(), time.Now().UTC())
	require.ErrorIs(t, err, ErrSlotsRoomNotFound)
}

func TestSlotsServiceListAvailableRepoError(t *testing.T) {
	repoErr := errors.New("list fail")
	svc := NewSlotsService(
		slotsRoomCheckerMock{},
		slotsScheduleReaderMock{},
		&slotsRepoMock{listFreeFn: func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
			return nil, repoErr
		}},
	)

	_, err := svc.ListAvailable(context.Background(), uuid.New(), time.Now().UTC())
	require.ErrorIs(t, err, repoErr)
}
