package postgres

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type SchedulesPostgresRepository struct {
	db *sql.DB
}

var _ repository.SchedulesRepository = (*SchedulesPostgresRepository)(nil)

func NewSchedulesPostgresRepository(db *sql.DB) *SchedulesPostgresRepository {
	return &SchedulesPostgresRepository{db: db}
}

func (r *SchedulesPostgresRepository) Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	const q = `
INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, room_id, days_of_week, start_time, end_time, created_at`

	return r.scanSchedule(r.db.QueryRowContext(
		ctx,
		q,
		schedule.ID,
		schedule.RoomID,
		pq.Array(schedule.DaysOfWeek),
		schedule.StartTime,
		schedule.EndTime,
		schedule.CreatedAtUTC,
	))
}

func (r *SchedulesPostgresRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error) {
	const q = `
SELECT id, room_id, days_of_week, start_time, end_time, created_at
FROM schedules
WHERE room_id = $1`

	out, err := r.scanSchedule(r.db.QueryRowContext(ctx, q, roomID))
	if isNoRows(err) {
		return models.Schedule{}, false, nil
	}
	if err != nil {
		return models.Schedule{}, false, err
	}

	return out, true, nil
}

func (r *SchedulesPostgresRepository) scanSchedule(scan rowScanner) (models.Schedule, error) {
	var out models.Schedule
	var days []int

	if err := scan.Scan(
		&out.ID,
		&out.RoomID,
		pq.Array(&days),
		&out.StartTime,
		&out.EndTime,
		&out.CreatedAtUTC,
	); err != nil {
		return models.Schedule{}, err
	}

	out.DaysOfWeek = days
	return out, nil
}
