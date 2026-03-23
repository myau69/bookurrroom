package postgres

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type RoomsPostgresRepository struct {
	db *sql.DB
}

var _ repository.RoomsRepository = (*RoomsPostgresRepository)(nil)

func NewRoomsPostgresRepository(db *sql.DB) *RoomsPostgresRepository {
	return &RoomsPostgresRepository{db: db}
}

func (r *RoomsPostgresRepository) Create(ctx context.Context, room models.Room) (models.Room, error) {
	const q = `
INSERT INTO room (id, name, description, capacity, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, capacity, created_at`

	return r.scanRoom(
		r.db.QueryRowContext(ctx, q, room.ID, room.Name, room.Description, room.Capacity, room.CreatedAtUtc),
	)
}

func (r *RoomsPostgresRepository) GetList(ctx context.Context) ([]models.Room, error) {
	const q = `
SELECT id, name, description, capacity, created_at
FROM rooms
ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Room, 0, 16)
	for rows.Next() {
		item, err := r.scanRoom(rows)
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

func (r *RoomsPostgresRepository) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, roomID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *RoomsPostgresRepository) scanRoom(scan rowScanner) (models.Room, error) {
	var item models.Room
	var description sql.NullString
	var capacity sql.NullInt64

	if err := scan.Scan(&item.ID, &item.Name, &description, &capacity, &item.CreatedAtUtc); err != nil {
		return models.Room{}, err
	}

	item.Description = nullStringPtr(description)
	item.Capacity = nullIntPtr(capacity)

	return item, nil
}
