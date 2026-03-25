package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

var ErrUsersAlreadyExists = errors.New("users: user with this email already exists")

type UsersService struct {
	repo repository.UsersRepository
	now  func() time.Time
}

func NewUsersService(repo repository.UsersRepository) *UsersService {
	return &UsersService{repo: repo, now: time.Now}
}

func (s *UsersService) Create(ctx context.Context, email string, role models.Role, passwordHash *string) (models.User, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if _, found, err := s.repo.GetByEmail(ctx, normalizedEmail); err != nil {
		return models.User{}, err
	} else if found {
		return models.User{}, ErrUsersAlreadyExists
	}

	user, err := models.NewUser(uuid.New(), normalizedEmail, role, s.now())
	if err != nil {
		return models.User{}, err
	}
	user.PasswordHash = passwordHash

	return s.repo.Create(ctx, user)
}

func (s *UsersService) GetByID(ctx context.Context, id uuid.UUID) (models.User, bool, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UsersService) GetByEmail(ctx context.Context, email string) (models.User, bool, error) {
	return s.repo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
}

func (s *UsersService) GetList(ctx context.Context) ([]models.User, error) {
	return s.repo.GetList(ctx)
}
