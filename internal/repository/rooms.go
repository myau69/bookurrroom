package repository

import (
	"context"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

type RoomsRepository interface {
	Create(ctx context.Context, room models.Room) (models.Room, error)
	GetList(ctx context.Context) ([]models.Room, error)
	Exists(ctx context.Context, roomID uuid.UUID) (bool, error)
}
