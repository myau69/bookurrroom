package postgres

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SlotsPostgresRepository struct {
	db *sql.DB
}

var _ repository.SlotsRepository = (*SlotsPostgresRepository)(nil)

func NewSlotsPostgresRepository(db *sql.DB) *SlotsPostgresRepository {
	return &SlotsPostgresRepository{db: db}
}

func (r *SlotsPostgresRepository) UpsertMany(ctx context.Context, values []models.Slot) error {
	if len(values) == 0 {
		return nil
	}

	const q = `
INSERT INTO slots (id, room_id, start_utc, end_utc)
VALUES ($1, $2, $3, $4)
ON CONFLICT ON CONSTRAINT slots_room_time_uniq DO NOTHING`

	for _, item := range values {
		if _, err := r.db.ExecContext(ctx, q, item.ID, item.RoomID, item.StartUTC, item.EndUTC); err != nil {
			return err
		}
	}

	return nil
}

func (r *SlotsPostgresRepository) ListFreeByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	dayStart := time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	const q = `
SELECT s.id, s.room_id, s.start_utc, s.end_utc
FROM slots s
LEFT JOIN bookings b
  ON b.slot_id = s.id
 AND b.status = 'active'
WHERE s.room_id = $1
  AND s.start_utc >= $2
  AND s.start_utc < $3
  AND b.id IS NULL
ORDER BY s.start_utc ASC`

	rows, err := r.db.QueryContext(ctx, q, roomID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	out := make([]models.Slot, 0, 32)
	for rows.Next() {
		item, err := r.scanSlot(rows)
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

func (r *SlotsPostgresRepository) GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
	const q = `
SELECT id, room_id, start_utc, end_utc
FROM slots
WHERE id = $1`

	item, err := r.scanSlot(r.db.QueryRowContext(ctx, q, slotID))
	if isNoRows(err) {
		return models.Slot{}, false, nil
	}
	if err != nil {
		return models.Slot{}, false, err
	}
	return item, true, nil
}

func (r *SlotsPostgresRepository) scanSlot(scan rowScanner) (models.Slot, error) {
	var item models.Slot
	if err := scan.Scan(&item.ID, &item.RoomID, &item.StartUTC, &item.EndUTC); err != nil {
		return models.Slot{}, err
	}
	return item, nil
}
