package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRoomInvalidID       = errors.New("rooms: invalid room id")
	ErrRoomNameRequired    = errors.New("rooms: name is required")
	ErrRoomInvalidCapacity = errors.New("rooms: capacity must be positive")
)

type Room struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	Capacity     *int
	CreatedAtUtc time.Time
}

func NewRoom(id uuid.UUID, name string, description *string, capacity *int, now time.Time) (Room, error) {
	room := Room{
		ID:           id,
		Name:         strings.TrimSpace(name),
		Description:  description,
		Capacity:     capacity,
		CreatedAtUtc: now.UTC(),
	}
	if err := room.ValidateForCreate(); err != nil {
		return Room{}, err
	}
	return room, nil
}

func (r Room) ValidateForCreate() error {
	if r.ID == uuid.Nil {
		return ErrRoomInvalidID
	}
	if strings.TrimSpace(r.Name) == "" {
		return ErrRoomNameRequired
	}
	if r.Capacity != nil && *r.Capacity <= 0 {
		return ErrRoomInvalidCapacity
	}
	return nil
}
