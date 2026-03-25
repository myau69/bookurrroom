package services

import (
	"bookurrroom/internal/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrBookingsNotFound  = errors.New("bookings: booking not found")
	ErrBookingsForbidden = errors.New("bookings: cannot cancel another user's booking")
)

func (s *BookingsService) Cancel(ctx context.Context, userID, bookingID uuid.UUID) (models.Booking, error) {
	booking, found, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return models.Booking{}, err
	}
	if found == false {
		return models.Booking{}, ErrBookingsNotFound
	}
	if booking.IsOwnedBy(userID) == false {
		return models.Booking{}, ErrBookingsForbidden
	}
	if booking.Status == models.BookingStatusCancelled {
		return booking, nil
	}

	return s.repo.Cancel(ctx, bookingID)
}
