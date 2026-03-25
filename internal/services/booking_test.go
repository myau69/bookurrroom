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

type bookingsRepoMock struct {
	createFn             func(ctx context.Context, booking models.Booking) (models.Booking, error)
	getByIDFn            func(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error)
	cancelFn             func(ctx context.Context, bookingID uuid.UUID) (models.Booking, error)
	listAllFn            func(ctx context.Context, page, pageSize int) ([]models.Booking, int, error)
	listMyFutureFn       func(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error)
	existsActiveBySlotFn func(ctx context.Context, slotID uuid.UUID) (bool, error)
}

var _ repository.BookingsRepository = (*bookingsRepoMock)(nil)

func (m *bookingsRepoMock) Create(ctx context.Context, booking models.Booking) (models.Booking, error) {
	if m.createFn == nil {
		return booking, nil
	}
	return m.createFn(ctx, booking)
}

func (m *bookingsRepoMock) GetByID(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error) {
	if m.getByIDFn == nil {
		return models.Booking{}, false, nil
	}
	return m.getByIDFn(ctx, bookingID)
}

func (m *bookingsRepoMock) Cancel(ctx context.Context, bookingID uuid.UUID) (models.Booking, error) {
	if m.cancelFn == nil {
		return models.Booking{ID: bookingID, Status: models.BookingStatusCancelled}, nil
	}
	return m.cancelFn(ctx, bookingID)
}

func (m *bookingsRepoMock) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	if m.listAllFn == nil {
		return []models.Booking{}, 0, nil
	}
	return m.listAllFn(ctx, page, pageSize)
}

func (m *bookingsRepoMock) ListMyFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error) {
	if m.listMyFutureFn == nil {
		return []models.Booking{}, nil
	}
	return m.listMyFutureFn(ctx, userID, now)
}

func (m *bookingsRepoMock) ExistsActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error) {
	if m.existsActiveBySlotFn == nil {
		return false, nil
	}
	return m.existsActiveBySlotFn(ctx, slotID)
}

type bookingsSlotReaderMock struct {
	getFn func(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error)
}

func (m bookingsSlotReaderMock) GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
	if m.getFn == nil {
		return models.Slot{}, false, nil
	}
	return m.getFn(ctx, slotID)
}

func TestBookingsServiceCreateSuccess(t *testing.T) {
	slotID := uuid.New()
	roomID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	svc := NewBookingsService(
		&bookingsRepoMock{},
		bookingsSlotReaderMock{getFn: func(ctx context.Context, gotSlotID uuid.UUID) (models.Slot, bool, error) {
			return models.Slot{ID: slotID, RoomID: roomID, StartUTC: now.Add(time.Hour), EndUTC: now.Add(90 * time.Minute)}, true, nil
		}},
	)
	svc.now = func() time.Time { return now }

	booking, err := svc.Create(context.Background(), userID, slotID, false)
	require.NoError(t, err)
	require.Equal(t, models.BookingStatusActive, booking.Status)
	require.Equal(t, userID, booking.UserID)
}

func TestBookingsServiceCreateSlotNotFound(t *testing.T) {
	svc := NewBookingsService(&bookingsRepoMock{}, bookingsSlotReaderMock{})
	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)
	require.ErrorIs(t, err, ErrBookingsSlotNotFound)
}

func TestBookingsServiceCreateSlotAlreadyBooked(t *testing.T) {
	now := time.Now().UTC()
	svc := NewBookingsService(
		&bookingsRepoMock{existsActiveBySlotFn: func(ctx context.Context, slotID uuid.UUID) (bool, error) { return true, nil }},
		bookingsSlotReaderMock{getFn: func(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
			return models.Slot{ID: slotID, RoomID: uuid.New(), StartUTC: now.Add(time.Hour), EndUTC: now.Add(90 * time.Minute)}, true, nil
		}},
	)
	svc.now = func() time.Time { return now }

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)
	require.ErrorIs(t, err, ErrBookingsSlotAlreadyBooked)
}

func TestBookingsServiceCreateSlotInPast(t *testing.T) {
	now := time.Now().UTC()
	svc := NewBookingsService(
		&bookingsRepoMock{},
		bookingsSlotReaderMock{getFn: func(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
			return models.Slot{ID: slotID, RoomID: uuid.New(), StartUTC: now.Add(-time.Hour), EndUTC: now.Add(-30 * time.Minute)}, true, nil
		}},
	)
	svc.now = func() time.Time { return now }

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)
	require.ErrorIs(t, err, ErrBookingsSlotInPast)
}

func TestBookingsServiceCancelIdempotent(t *testing.T) {
	bookingID := uuid.New()
	ownerID := uuid.New()
	svc := NewBookingsService(
		&bookingsRepoMock{getByIDFn: func(ctx context.Context, id uuid.UUID) (models.Booking, bool, error) {
			return models.Booking{ID: bookingID, UserID: ownerID, Status: models.BookingStatusCancelled}, true, nil
		}},
		bookingsSlotReaderMock{},
	)

	booking, err := svc.Cancel(context.Background(), ownerID, bookingID)
	require.NoError(t, err)
	require.Equal(t, models.BookingStatusCancelled, booking.Status)
}

func TestBookingsServiceCancelForbidden(t *testing.T) {
	bookingID := uuid.New()
	svc := NewBookingsService(
		&bookingsRepoMock{getByIDFn: func(ctx context.Context, id uuid.UUID) (models.Booking, bool, error) {
			return models.Booking{ID: bookingID, UserID: uuid.New(), Status: models.BookingStatusActive}, true, nil
		}},
		bookingsSlotReaderMock{},
	)

	_, err := svc.Cancel(context.Background(), uuid.New(), bookingID)
	require.ErrorIs(t, err, ErrBookingsForbidden)
}

func TestBookingsServiceListAllInvalidPageSize(t *testing.T) {
	svc := NewBookingsService(&bookingsRepoMock{}, bookingsSlotReaderMock{})
	_, _, err := svc.ListAll(context.Background(), 1, 101)
	require.ErrorIs(t, err, ErrBookingsInvalidPagination)
}

func TestBookingsServiceListAllRepoError(t *testing.T) {
	repoErr := errors.New("list failed")
	svc := NewBookingsService(&bookingsRepoMock{
		listAllFn: func(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
			return nil, 0, repoErr
		},
	}, bookingsSlotReaderMock{})

	_, _, err := svc.ListAll(context.Background(), 1, 20)
	require.ErrorIs(t, err, repoErr)
}
