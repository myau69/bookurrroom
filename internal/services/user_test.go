package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
)

type usersRepoMock struct {
	createFn     func(ctx context.Context, user models.User) (models.User, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (models.User, bool, error)
	getByEmailFn func(ctx context.Context, email string) (models.User, bool, error)
	listFn       func(ctx context.Context) ([]models.User, error)
}

var _ repository.UsersRepository = (*usersRepoMock)(nil)

func (m *usersRepoMock) Create(ctx context.Context, user models.User) (models.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return user, nil
}

func (m *usersRepoMock) GetByID(ctx context.Context, id uuid.UUID) (models.User, bool, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return models.User{}, false, nil
}

func (m *usersRepoMock) GetByEmail(ctx context.Context, email string) (models.User, bool, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return models.User{}, false, nil
}

func (m *usersRepoMock) GetList(ctx context.Context) ([]models.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func TestUsersServiceCreateSuccess(t *testing.T) {
	repo := &usersRepoMock{}
	svc := NewUsersService(repo)

	user, err := svc.Create(context.Background(), "USER@Example.com", models.RoleUser, nil)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", user.Email)
	require.Equal(t, models.RoleUser, user.Role)
	require.NotEqual(t, uuid.Nil, user.ID)
}

func TestUsersServiceCreateDuplicateEmail(t *testing.T) {
	repo := &usersRepoMock{
		getByEmailFn: func(ctx context.Context, email string) (models.User, bool, error) {
			return models.User{ID: uuid.New()}, true, nil
		},
	}
	svc := NewUsersService(repo)

	_, err := svc.Create(context.Background(), "user@example.com", models.RoleUser, nil)
	require.ErrorIs(t, err, ErrUsersAlreadyExists)
}

func TestUsersServiceCreateRepoError(t *testing.T) {
	repoErr := errors.New("db exploded")
	repo := &usersRepoMock{
		getByEmailFn: func(ctx context.Context, email string) (models.User, bool, error) {
			return models.User{}, false, repoErr
		},
	}
	svc := NewUsersService(repo)

	_, err := svc.Create(context.Background(), "user@example.com", models.RoleUser, nil)
	require.ErrorIs(t, err, repoErr)
}
