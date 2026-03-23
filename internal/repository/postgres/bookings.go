package postgres

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type BookingsPostgresRepository struct {
	db *sql.DB
}

var _ repository.BookingsRepository = (*BookingsPostgresRepository)(nil)

func NewBookingsPostgresRepository(db *sql.DB) *BookingsPostgresRepository {
	return &BookingsPostgresRepository{db: db}
}

func (r *BookingsPostgresRepository) Create(ctx context.Context, booking models.Booking) (models.Booking, error) {
	const q = `
INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at, cancelled_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, slot_id, user_id, status, conference_link, created_at, cancelled_at`

	return r.scanBooking(r.db.QueryRowContext(
		ctx,
		q,
		booking.ID,
		booking.SlotID,
		booking.UserID,
		string(booking.Status),
		booking.ConferenceLink,
		booking.CreatedatUTC,
		booking.CancelledAtUTC,
	))
}

func (r *BookingsPostgresRepository) GetByID(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error) {
	const q = `
SELECT id, slot_id, user_id, status, conference_link, created_at, cancelled_at
FROM bookings
WHERE id = $1`

	out, found, err := r.scanOne(ctx, q, bookingID)
	return out, found, err
}

func (r *BookingsPostgresRepository) Cancel(ctx context.Context, bookingID uuid.UUID) (models.Booking, error) {
	const q = `
UPDATE bookings
SET status = $2,
    cancelled_at = COALESCE(cancelled_at, $3)
WHERE id = $1
RETURNING id, slot_id, user_id, status, conference_link, created_at, cancelled_at`

	now := time.Now().UTC()

	return r.scanBooking(
		r.db.QueryRowContext(ctx, q, bookingID, string(models.BookingStatusCancelled), now),
	)
}

func (r *BookingsPostgresRepository) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	const countQ = `SELECT COUNT(*) FROM bookings`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	const listQ = `
SELECT id, slot_id, user_id, status, conference_link, created_at, cancelled_at
FROM bookings
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, listQ, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]models.Booking, 0, pageSize)
	for rows.Next() {
		item, err := r.scanBooking(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *BookingsPostgresRepository) ListMyFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error) {
	const q = `
SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at, b.cancelled_at
FROM bookings b
JOIN slots s ON s.id = b.slot_id
WHERE b.user_id = $1
  AND b.status = 'active'
  AND s.start_utc >= $2
ORDER BY s.start_utc ASC`

	rows, err := r.db.QueryContext(ctx, q, userID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Booking, 0, 16)
	for rows.Next() {
		item, err := r.scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BookingsPostgresRepository) ExistsActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM bookings WHERE slot_id = $1 AND status = 'active')`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, slotID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *BookingsPostgresRepository) scanOne(ctx context.Context, q string, arg any) (models.Booking, bool, error) {
	item, err := r.scanBooking(r.db.QueryRowContext(ctx, q, arg))
	if isNoRows(err) {
		return models.Booking{}, false, nil
	}
	if err != nil {
		return models.Booking{}, false, err
	}

	return item, true, nil
}

func (r *BookingsPostgresRepository) scanBooking(scan rowScanner) (models.Booking, error) {
	var item models.Booking
	var statusRaw string
	var conferenceLink sql.NullString
	var cancelledAt sql.NullTime

	if err := scan.Scan(
		&item.ID,
		&item.SlotID,
		&item.UserID,
		&statusRaw,
		&conferenceLink,
		&item.CreatedatUTC,
		&cancelledAt,
	); err != nil {
		return models.Booking{}, err
	}

	status, err := models.ParseBookingStatus(statusRaw)
	if err != nil {
		return models.Booking{}, fmt.Errorf("parse status: %w", err)
	}
	item.Status = status
	item.ConferenceLink = nullStringPtr(conferenceLink)
	item.CancelledAtUTC = nullTimeUTCPtr(cancelledAt)

	return item, nil
}
