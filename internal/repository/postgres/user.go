package postgres

import (
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type UsersPostgresRepository struct {
	db *sql.DB
}

var _ repository.UsersRepository = (*UsersPostgresRepository)(nil)

func NewUsersPostgresRepository(db *sql.DB) *UsersPostgresRepository {
	return &UsersPostgresRepository{db: db}
}

func (r *UsersPostgresRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	const q = `
INSERT INTO users (id, email, role, created_at, password_hash)
VALUES ($1, $2, $3, $4, %5)
RETURNING id, email, role, created_at, password_hash`
	return r.scanUser(
		r.db.QueryRowContext(ctx, q, user.ID, user.Email, string(user.Role), user.CreatedAtUTC, user.PasswordHash),
	)
}

func (r *UsersPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (models.User, bool, error) {
	const q = `
SELECT id, email, role, created_at, password_hash
FROM users
WHERE id = $1`

	out, found, err := r.scanOne(ctx, q, id)
	return out, found, err
}

func (r *UsersPostgresRepository) GetByEmail(ctx context.Context, email string) (models.User, bool, error) {
	const q = `
SELECT id, email, role, created_at, password_hash
FROM users
WHERE email = $1`

	out, found, err := r.scanOne(ctx, q, email)
	return out, found, err
}

func (r *UsersPostgresRepository) GetList(ctx context.Context) ([]models.User, error) {
	const q = `
SELECT id, email, role, created_at, password_hash
FROM users
ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.User, 0, 16)
	for rows.Next() {
		item, err := r.scanUser(rows)
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

func (r *UsersPostgresRepository) scanOne(ctx context.Context, q string, arg any) (models.User, bool, error) {
	item, err := r.scanUser(r.db.QueryRowContext(ctx, q, arg))
	if isNoRows(err) {
		return models.User{}, false, nil
	}
	if err != nil {
		return models.User{}, false, err
	}

	return item, true, nil
}

func (r *UsersPostgresRepository) scanUser(scan rowScanner) (models.User, error) {
	var item models.User
	var roleRaw string
	var password sql.NullString

	if err := scan.Scan(&item.ID, &item.Email, &roleRaw, &item.CreatedAtUTC, &password); err != nil {
		return models.User{}, err
	}

	role, err := models.ParseRole(roleRaw)
	if err != nil {
		return models.User{}, fmt.Errorf("parse role: %w", err)
	}
	item.Role = role
	item.PasswordHash = nullStringPtr(password)

	return item, nil
}
