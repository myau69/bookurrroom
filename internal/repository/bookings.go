package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

var ErrActiveBookingConflict = errors.New("bookings: active booking conflict")

type BookingsRepository interface {
	Create(ctx context.Context, booking models.Booking) (models.Booking, error)
	GetByID(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error)
	Cancel(ctx context.Context, bookingID uuid.UUID) (models.Booking, error)
	ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error)
	ListMyFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error)
	ExistsActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error)
}
