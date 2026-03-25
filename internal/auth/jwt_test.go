package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"bookurrroom/internal/models"
)

func TestTokenManagerIssueParse(t *testing.T) {
	manager := NewTokenManager("secret", time.Hour)
	userID := uuid.New()

	token, err := manager.Issue(userID, models.RoleUser)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	principal, err := manager.Parse(token)
	require.NoError(t, err)
	require.Equal(t, userID, principal.UserID)
	require.Equal(t, models.RoleUser, principal.Role)
}

func TestTokenManagerParseInvalid(t *testing.T) {
	manager := NewTokenManager("secret", time.Hour)

	_, err := manager.Parse("not-a-token")
	require.ErrorIs(t, err, ErrInvalidToken)

	token, err := manager.Issue(uuid.New(), models.RoleAdmin)
	require.NoError(t, err)

	other := NewTokenManager("another-secret", time.Hour)
	_, err = other.Parse(token)
	require.ErrorIs(t, err, ErrInvalidToken)
}
