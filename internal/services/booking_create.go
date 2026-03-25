package services

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrBookingsSlotNotFound      = errors.New("bookings: slot not found")
	ErrBookingsSlotAlreadyBooked = errors.New("bookings: slot already booked")
	ErrBookingsSlotInPast        = errors.New("bookings: slot is in the past")
)

func (s *BookingsService) Create(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (models.Booking, error) {
	slot, found, err := s.slots.GetByID(ctx, slotID)
	if err != nil {
		return models.Booking{}, err
	}
	if found == false {
		return models.Booking{}, ErrBookingsSlotNotFound
	}
	if slot.IsPast(s.now()) {
		return models.Booking{}, ErrBookingsSlotInPast
	}

	busy, err := s.repo.ExistsActiveBySlotID(ctx, slotID)
	if err != nil {
		return models.Booking{}, err
	}
	if busy {
		return models.Booking{}, ErrBookingsSlotAlreadyBooked
	}

	booking, err := models.NewBooking(uuid.New(), slotID, userID, s.now())
	if err != nil {
		return models.Booking{}, err
	}

	if createConferenceLink {
		link := fmt.Sprintf("https://conference.local/booking/%s", booking.ID.String())
		booking.ConferenceLink = &link
	}

	created, err := s.repo.Create(ctx, booking)
	if err != nil {
		if errors.Is(err, repository.ErrActiveBookingConflict) {
			return models.Booking{}, ErrBookingsSlotAlreadyBooked
		}
		return models.Booking{}, err
	}

	return created, nil
}
