package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

type BookingsSlotReader interface {
	GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error)
}

type BookingsService struct {
	repo  repository.BookingsRepository
	slots BookingsSlotReader
	now   func() time.Time
}

func NewBookingsService(repo repository.BookingsRepository, slots BookingsSlotReader) *BookingsService {
	return &BookingsService{repo: repo, slots: slots, now: time.Now}
}
