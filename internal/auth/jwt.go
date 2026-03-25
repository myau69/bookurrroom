package auth

import (
	"bookurrroom/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("auth: invalid token")

type Principal struct {
	UserID uuid.UUID
	Role   models.Role
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

type claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *TokenManager) Issue(userID uuid.UUID, role models.Role) (string, error) {
	now := time.Now().UTC()
	payload := claims{
		UserID: userID.String(),
		Role:   string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(now),
		},
	}
	if m.ttl > 0 {
		payload.ExpiresAt = jwt.NewNumericDate(now.Add(m.ttl))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString(m.secret)
}

func (m *TokenManager) Parse(tokenRaw string) (Principal, error) {
	var parsedClaims claims
	token, err := jwt.ParseWithClaims(tokenRaw, &parsedClaims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok == false {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || token.Valid == false {
		return Principal{}, ErrInvalidToken
	}
	userID, err := uuid.Parse(parsedClaims.UserID)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	role, err := models.ParseRole(parsedClaims.Role)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}

	return Principal{
		UserID: userID,
		Role:   role,
	}, nil
}
