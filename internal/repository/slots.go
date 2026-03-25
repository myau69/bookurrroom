package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

type SlotsRepository interface {
	UpsertMany(ctx context.Context, slots []models.Slot) error
	ListFreeByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error)
	GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error)
}
