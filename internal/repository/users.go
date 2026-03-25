package repository

import (
	"context"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

type UsersRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.User, bool, error)
	GetByEmail(ctx context.Context, email string) (models.User, bool, error)
	GetList(ctx context.Context) ([]models.User, error)
}
