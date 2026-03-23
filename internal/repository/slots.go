package repository

import (
	"bookurrroom/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type SlotsRepository interface {
	UpsertMany(ctx context.Context, slots []models.Slot) error
	ListFreeByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error)
	GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error)
}
