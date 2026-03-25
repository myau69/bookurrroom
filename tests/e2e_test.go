package tests

import (
	"bookurrroom/internal/auth"
	"bookurrroom/internal/bootstrap"
	"bookurrroom/internal/controllers"
	"bookurrroom/internal/repository/postgres"
	"bookurrroom/internal/services"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testDummyAdminID = "00000000-0000-0000-0000-000000000001"
	testDummyUserID  = "00000000-0000-0000-0000-000000000002"
)

type e2eEnv struct {
	db     *sql.DB
	server *httptest.Server
}

type e2ePreparedFlow struct {
	adminToken string
	userToken  string
	roomID     uuid.UUID
	slots      []e2eSlot
}

func TestE2ECreateFlow(t *testing.T) {
	env := setupE2E(t)
	t.Cleanup(env.close)

	flow := mustPrepareFlow(t, env.server.URL, "E2E Alpha")

	bookingID := mustCreateBooking(t, env.server.URL, flow.userToken, flow.slots[0].ID, false)
	require.NotEqual(t, uuid.Nil, bookingID)
}

func TestE2ECancelFlow(t *testing.T) {
	env := setupE2E(t)
	t.Cleanup(env.close)

	flow := mustPrepareFlow(t, env.server.URL, "E2E Beta")

	bookingID := mustCreateBooking(t, env.server.URL, flow.userToken, flow.slots[0].ID, false)

	status := mustCancelBooking(t, env.server.URL, flow.userToken, bookingID)
	require.Equal(t, "cancelled", status)

	status = mustCancelBooking(t, env.server.URL, flow.userToken, bookingID)
	require.Equal(t, "cancelled", status)
}

func TestE2EAdminCannotCreateBooking(t *testing.T) {
	env := setupE2E(t)
	t.Cleanup(env.close)

	flow := mustPrepareFlow(t, env.server.URL, "E2E Gamma")

	status := createBookingStatus(t, env.server.URL, flow.adminToken, flow.slots[0].ID, false)
	require.Equal(t, http.StatusForbidden, status)
}

func TestE2EMyBookingsReturnsOnlyActiveFuture(t *testing.T) {
	env := setupE2E(t)
	t.Cleanup(env.close)

	flow := mustPrepareFlow(t, env.server.URL, "E2E Delta")

	bookingID := mustCreateBooking(t, env.server.URL, flow.userToken, flow.slots[0].ID, false)

	myBookings := mustListMyBookings(t, env.server.URL, flow.userToken)
	require.Len(t, myBookings, 1)
	require.Equal(t, bookingID, myBookings[0].ID)
	require.Equal(t, "active", myBookings[0].Status)

	_ = mustCancelBooking(t, env.server.URL, flow.userToken, bookingID)

	myBookings = mustListMyBookings(t, env.server.URL, flow.userToken)
	require.Len(t, myBookings, 0)
}

func mustPrepareFlow(t *testing.T, baseURL, roomName string) e2ePreparedFlow {
	t.Helper()

	adminToken := mustDummyLogin(t, baseURL, "admin")
	roomID := mustCreateRoom(t, baseURL, adminToken, roomName)
	mustCreateSchedule(t, baseURL, adminToken, roomID)

	userToken := mustDummyLogin(t, baseURL, "user")
	slots := mustListSlots(t, baseURL, userToken, roomID, time.Now().UTC().AddDate(0, 0, 1))
	require.NotEmpty(t, slots)

	return e2ePreparedFlow{
		adminToken: adminToken,
		userToken:  userToken,
		roomID:     roomID,
		slots:      slots,
	}
}

func setupE2E(t *testing.T) *e2eEnv {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL is required")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../migrations"
	}

	if err := bootstrap.RunMigrations(dbURL, migrationsPath); err != nil {
		t.Skipf("skip e2e, cannot run migrations: %v", err)
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Skipf("skip e2e, cannot connect database: %v", err)
	}

	mustCleanDatabase(t, db)

	usersRepo := postgres.NewUsersPostgresRepository(db)
	roomsRepo := postgres.NewRoomsPostgresRepository(db)
	schedulesRepo := postgres.NewSchedulesPostgresRepository(db)
	slotsRepo := postgres.NewSlotsPostgresRepository(db)
	bookingsRepo := postgres.NewBookingsPostgresRepository(db)

	usersSvc := services.NewUsersService(usersRepo)
	roomsSvc := services.NewRoomsService(roomsRepo)
	schedulesSvc := services.NewSchedulesService(schedulesRepo, roomsSvc)
	slotsSvc := services.NewSlotsService(roomsSvc, schedulesSvc, slotsRepo)
	bookingsSvc := services.NewBookingsService(bookingsRepo, slotsSvc)

	tokenManager := auth.NewTokenManager("e2e-secret", time.Hour)
	router := controllers.NewRouter(
		usersSvc,
		roomsSvc,
		schedulesSvc,
		slotsSvc,
		bookingsSvc,
		slotsRepo,
		services.NewDefaultSlotsPlanner(),
		tokenManager,
	)

	return &e2eEnv{
		db:     db,
		server: httptest.NewServer(router),
	}
}

func (e *e2eEnv) close() {
	e.server.Close()
	_ = e.db.Close()
}

func mustCleanDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "TRUNCATE bookings, slots, schedules, rooms RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id NOT IN 
                        				($1::uuid, $2::uuid)`,
		testDummyAdminID, testDummyUserID)
	require.NoError(t, err)
}

func mustDummyLogin(t *testing.T, baseURL, role string) string {
	t.Helper()

	body := map[string]string{"role": role}
	res := doJSON(t, http.MethodPost, baseURL+"/dummyLogin", "", body)
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var out struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	require.NotEmpty(t, out.Token)
	return out.Token
}

func mustCreateRoom(t *testing.T, baseURL, token, name string) uuid.UUID {
	t.Helper()

	body := map[string]any{
		"name":        name,
		"description": "E2E room",
		"capacity":    6,
	}
	res := doJSON(t, http.MethodPost, baseURL+"/rooms/create", token, body)
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var out struct {
		Room struct {
			ID uuid.UUID `json:"id"`
		} `json:"room"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	require.NotEqual(t, uuid.Nil, out.Room.ID)
	return out.Room.ID
}

func mustCreateSchedule(t *testing.T, baseURL, token string, roomID uuid.UUID) {
	t.Helper()

	body := map[string]any{
		"roomId":     roomID,
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}
	res := doJSON(t, http.MethodPost, baseURL+"/rooms/"+roomID.String()+"/schedule/create", token, body)
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusCreated, res.StatusCode)
}

type e2eSlot struct {
	ID uuid.UUID `json:"id"`
}

type e2eBooking struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func mustListSlots(t *testing.T, baseURL, token string, roomID uuid.UUID, date time.Time) []e2eSlot {
	t.Helper()

	url := baseURL + "/rooms/" + roomID.String() + "/slots/list?date=" + date.UTC().Format("2006-01-02")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var out struct {
		Slots []e2eSlot `json:"slots"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	return out.Slots
}

func mustCreateBooking(t *testing.T, baseURL, token string, slotID uuid.UUID, createConferenceLink bool) uuid.UUID {
	t.Helper()

	body := map[string]any{
		"slotId":               slotID,
		"createConferenceLink": createConferenceLink,
	}
	res := doJSON(t, http.MethodPost, baseURL+"/bookings/create", token, body)
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var out struct {
		Booking struct {
			ID uuid.UUID `json:"id"`
		} `json:"booking"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	require.NotEqual(t, uuid.Nil, out.Booking.ID)
	return out.Booking.ID
}

func createBookingStatus(t *testing.T, baseURL, token string, slotID uuid.UUID, createConferenceLink bool) int {
	t.Helper()

	body := map[string]any{
		"slotId":               slotID,
		"createConferenceLink": createConferenceLink,
	}
	res := doJSON(t, http.MethodPost, baseURL+"/bookings/create", token, body)
	defer mustCloseBody(t, res)

	return res.StatusCode
}

func mustListMyBookings(t *testing.T, baseURL, token string) []e2eBooking {
	t.Helper()

	url := baseURL + "/bookings/my"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var out struct {
		Bookings []e2eBooking `json:"bookings"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	return out.Bookings
}

func mustCancelBooking(t *testing.T, baseURL, token string, bookingID uuid.UUID) string {
	t.Helper()

	res := doJSON(t, http.MethodPost, baseURL+"/bookings/"+bookingID.String()+"/cancel", token, map[string]any{})
	defer mustCloseBody(t, res)

	require.Equal(t, http.StatusOK, res.StatusCode)

	var out struct {
		Booking struct {
			Status string `json:"status"`
		} `json:"booking"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	return out.Booking.Status
}

func doJSON(t *testing.T, method, url, token string, payload any) *http.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return res
}

func mustCloseBody(t *testing.T, res *http.Response) {
	t.Helper()
	require.NotNil(t, res)
	require.NotNil(t, res.Body)
	require.NoError(t, res.Body.Close())
}
