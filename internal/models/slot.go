package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const SlotDuration = 30 * time.Minute

var (
	ErrSlotInvalidID        = errors.New("slots: invalid slots id")
	ErrSlotInvalidRoomId    = errors.New("slots: invalid room id")
	ErrSlotInvalidTimeRange = errors.New("slots: end time must be after start time")
	ErrSlotInvalidDuration  = errors.New("slots: slot duration bust be 30 minutes")
	ErrSlotTimeUTC          = errors.New("slots: time must be in UTC")
)

type Slot struct {
	ID       uuid.UUID
	RoomID   uuid.UUID
	StartUTC time.Time
	EndUTC   time.Time
}

func NewSlot(id, roomID uuid.UUID, startUTC time.Time, endUTC time.Time) (Slot, error) {
	slot := Slot{
		ID:       id,
		RoomID:   roomID,
		StartUTC: startUTC,
		EndUTC:   endUTC,
	}
	if err := slot.Validate(); err != nil {
		return Slot{}, err
	}
	return slot, nil
}

func (s Slot) Validate() error {
	if s.ID == uuid.Nil {
		return ErrSlotInvalidID
	}
	if s.RoomID == uuid.Nil {
		return ErrRoomInvalidID
	}
	if s.StartUTC.Location() != time.UTC || s.EndUTC.Location() != time.UTC {
		return ErrSlotTimeUTC
	}
	if s.EndUTC.Sub(s.StartUTC) != SlotDuration {
		return ErrSlotInvalidDuration
	}
	return nil
}

func (s Slot) IsPast(now time.Time) bool {
	return s.StartUTC.Before(now.UTC())
}
