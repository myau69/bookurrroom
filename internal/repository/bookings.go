package repository

import (
	"bookurrroom/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type BookingsRepository interface {
	Create(ctx context.Context, boking models.Booking) (models.Booking, error)
	GetByID(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error)
	Cancel(ctx context.Context, bookingID uuid.UUID) (models.Booking, error)
	ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error)
	ListMyFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error)
	ExistsActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error)
}
