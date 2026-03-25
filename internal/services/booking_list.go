package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"bookurrroom/internal/models"
)

var ErrBookingsInvalidPagination = errors.New("bookings: invalid pagination")

type BookingsPagination struct {
	Page     int
	PageSize int
	Total    int
}

func (s *BookingsService) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, BookingsPagination, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		return nil, BookingsPagination{}, ErrBookingsInvalidPagination
	}

	items, total, err := s.repo.ListAll(ctx, page, pageSize)
	if err != nil {
		return nil, BookingsPagination{}, err
	}

	return items, BookingsPagination{Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *BookingsService) ListMy(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error) {
	return s.repo.ListMyFuture(ctx, userID, now.UTC())
}
