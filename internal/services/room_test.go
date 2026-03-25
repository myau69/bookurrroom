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

type roomsRepoMock struct {
	createFn func(ctx context.Context, room models.Room) (models.Room, error)
	listFn   func(ctx context.Context) ([]models.Room, error)
	existsFn func(ctx context.Context, roomID uuid.UUID) (bool, error)
}

var _ repository.RoomsRepository = (*roomsRepoMock)(nil)

func (m *roomsRepoMock) Create(ctx context.Context, room models.Room) (models.Room, error) {
	if m.createFn == nil {
		return room, nil
	}
	return m.createFn(ctx, room)
}

func (m *roomsRepoMock) GetList(ctx context.Context) ([]models.Room, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx)
}

func (m *roomsRepoMock) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	if m.existsFn == nil {
		return false, nil
	}
	return m.existsFn(ctx, roomID)
}

func TestRoomsServiceCreateSuccess(t *testing.T) {
	svc := NewRoomsService(&roomsRepoMock{})
	room, err := svc.Create(context.Background(), "Alpha", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "Alpha", room.Name)
	require.NotEqual(t, uuid.Nil, room.ID)
}

func TestRoomsServiceCreateInvalidName(t *testing.T) {
	svc := NewRoomsService(&roomsRepoMock{})
	_, err := svc.Create(context.Background(), "   ", nil, nil)
	require.ErrorIs(t, err, models.ErrRoomNameRequired)
}

func TestRoomsServiceGetListRepoError(t *testing.T) {
	repoErr := errors.New("list failed")
	svc := NewRoomsService(&roomsRepoMock{
		listFn: func(ctx context.Context) ([]models.Room, error) {
			return nil, repoErr
		},
	})
	_, err := svc.GetList(context.Background())
	require.ErrorIs(t, err, repoErr)
}
