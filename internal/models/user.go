package models

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserInvalidID    = errors.New("users: invalid user id")
	ErrUserEmptyEmail   = errors.New("users: email is required")
	ErrUserInvalidEmail = errors.New("users: invalid user email")
	ErrUserInvalidRole  = errors.New("user: invalid user role")
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Role         Role
	CreatedAtUTC time.Time
	PasswordHash *string
}

func NewUser(id uuid.UUID, email string, role Role, now time.Time) (User, error) {
	user := User{
		ID:           id,
		Email:        strings.TrimSpace(strings.ToLower(email)),
		Role:         role,
		CreatedAtUTC: now.UTC(),
	}
	if err := user.ValidateForCreate(); err != nil {
		return User{}, err
	}
	return user, nil
}

func (u User) ValidateForCreate() error {
	if u.ID == uuid.Nil {
		return ErrUserInvalidID
	}
	if strings.TrimSpace(u.Email) == "" {
		return ErrUserEmptyEmail
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return ErrUserInvalidEmail
	}
	if _, err := ParseRole(string(u.Role)); err != nil {
		return ErrUserInvalidRole
	}
	return nil
}

func ParseRole(raw string) (Role, error) {
	role := Role(strings.TrimSpace(strings.ToLower(raw)))
	switch role {
	case RoleAdmin, RoleUser:
		return role, nil
	default:
		return "", ErrUserInvalidRole
	}
}
