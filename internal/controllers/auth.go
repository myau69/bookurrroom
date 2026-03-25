package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"bookurrroom/internal/auth"
	"bookurrroom/internal/models"
	"bookurrroom/internal/services"
)

var (
	dummyAdminID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type AuthHandler struct {
	users  *services.UsersService
	tokens *auth.TokenManager
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type dummyLoginRequest struct {
	Role string `json:"role"`
}

type userResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"createdAt"`
}

func NewAuthHandler(users *services.UsersService, tokens *auth.TokenManager) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens}
}

// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "register payload"
// @Success 201 {object} registerResponse
// @Failure 400 {object} errorPayload
// @Failure 500 {object} errorPayload
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	role, err := models.ParseRole(req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "password is required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	hashString := string(hash)

	user, err := h.users.Create(r.Context(), req.Email, role, &hashString)
	if err != nil {
		if errors.Is(err, services.ErrUsersAlreadyExists) || isUserValidationError(err) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]userResponse{
		"user": mapUser(user),
	})
}

// @Summary Login by email/password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "login payload"
// @Success 200 {object} authTokenResponse
// @Failure 400 {object} errorPayload
// @Failure 401 {object} errorPayload
// @Failure 500 {object} errorPayload
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "password is required")
		return
	}

	user, found, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	if !found || user.PasswordHash == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	token, err := h.tokens.Issue(user.ID, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// @Summary Dummy login (test only)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dummyLoginRequest true "dummy login payload"
// @Success 200 {object} authTokenResponse
// @Failure 400 {object} errorPayload
// @Failure 500 {object} errorPayload
// @Router /dummyLogin [post]
func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	role, err := models.ParseRole(req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		return
	}

	userID := dummyUserID
	if role == models.RoleAdmin {
		userID = dummyAdminID
	}

	token, err := h.tokens.Issue(userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func isUserValidationError(err error) bool {
	return errors.Is(err, models.ErrUserInvalidID) ||
		errors.Is(err, models.ErrUserEmptyEmail) ||
		errors.Is(err, models.ErrUserInvalidEmail) ||
		errors.Is(err, models.ErrUserInvalidRole)
}

func mapUser(user models.User) userResponse {
	var createdAt *time.Time
	if user.CreatedAtUTC.IsZero() == false {
		v := user.CreatedAtUTC.UTC()
		createdAt = &v
	}

	return userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: createdAt,
	}
}
