package repository

import (
	"bookurrroom/internal/models"
	"context"

	"github.com/google/uuid"
)

type SchedulesRepository interface {
	Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error)
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error)
}
