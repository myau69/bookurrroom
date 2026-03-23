package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrBookingInvalidID     = errors.New("bookings: invalid id")
	ErrBookingInvalidSlotID = errors.New("bookings: invalid slot id")
	ErrBookingInvalidUserID = errors.New("bookings: invalid user id")
	ErrBookingInvalidStatus = errors.New("bookings: invalid status")
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

func ParseBookingStatus(raw string) (BookingStatus, error) {
	status := BookingStatus(strings.TrimSpace(strings.ToLower(raw)))
	switch status {
	case BookingStatusActive, BookingStatusCancelled:
		return status, nil
	default:
		return "", ErrBookingInvalidStatus
	}
}

type Booking struct {
	ID             uuid.UUID
	SlotID         uuid.UUID
	UserID         uuid.UUID
	Status         BookingStatus
	ConferenceLink *string
	CreatedatUTC   time.Time
	CancelledAtUTC *time.Time
}

func NewBooking(id, slotID, userID uuid.UUID, now time.Time) (Booking, error) {
	booking := Booking{
		ID:           id,
		SlotID:       slotID,
		UserID:       userID,
		Status:       BookingStatusActive,
		CreatedatUTC: now.UTC(),
	}
	if err := booking.ValidateForCreate(); err != nil {
		return Booking{}, err
	}
	return booking, nil
}

func (b Booking) ValidateForCreate() error {
	if b.ID == uuid.Nil {
		return ErrBookingInvalidID
	}
	if b.SlotID == uuid.Nil {
		return ErrBookingInvalidSlotID
	}
	if b.UserID == uuid.Nil {
		return ErrBookingInvalidUserID
	}
	return nil
}

func (b Booking) IsOwnedBy(userID uuid.UUID) bool {
	return b.UserID == userID
}

func (b *Booking) Cancel(now time.Time) {
	if b.Status == BookingStatusCancelled {
		return
	}
	b.Status = BookingStatusCancelled
	currentTime := now.UTC()
	b.CancelledAtUTC = &currentTime
}
