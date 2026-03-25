package repository

import (
	"context"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

type SchedulesRepository interface {
	Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error)
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error)
}
