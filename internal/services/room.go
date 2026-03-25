package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

type RoomsService struct {
	repo repository.RoomsRepository
	now  func() time.Time
}

func NewRoomsService(repo repository.RoomsRepository) *RoomsService {
	return &RoomsService{repo: repo, now: time.Now}
}

func (s *RoomsService) Create(ctx context.Context, name string, description *string, capacity *int) (models.Room, error) {
	room, err := models.NewRoom(uuid.New(), name, description, capacity, s.now())
	if err != nil {
		return models.Room{}, err
	}
	return s.repo.Create(ctx, room)
}

func (s *RoomsService) GetList(ctx context.Context) ([]models.Room, error) {
	return s.repo.GetList(ctx)
}

func (s *RoomsService) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	return s.repo.Exists(ctx, roomID)
}
