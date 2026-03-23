package repository

import (
	"bookurrroom/internal/models"
	"context"

	"github.com/google/uuid"
)

type RoomsRepository interface {
	Create(ctx context.Context, room models.Room) (models.Room, error)
	GetList(ctx context.Context) ([]models.Room, error)
	Exists(ctx context.Context, roomID uuid.UUID) (bool, error)
}
