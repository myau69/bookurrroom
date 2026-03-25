package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"bookurrroom/internal/auth"
	"bookurrroom/internal/models"
	"bookurrroom/internal/services"
)

type memStore struct {
	mu sync.Mutex

	users       map[uuid.UUID]models.User
	userByEmail map[string]uuid.UUID

	rooms map[uuid.UUID]models.Room

	schedulesByRoom map[uuid.UUID]models.Schedule

	slots map[uuid.UUID]models.Slot

	bookings map[uuid.UUID]models.Booking
}

func newMemStore() *memStore {
	return &memStore{
		users:           map[uuid.UUID]models.User{},
		userByEmail:     map[string]uuid.UUID{},
		rooms:           map[uuid.UUID]models.Room{},
		schedulesByRoom: map[uuid.UUID]models.Schedule{},
		slots:           map[uuid.UUID]models.Slot{},
		bookings:        map[uuid.UUID]models.Booking{},
	}
}

func (m *memStore) Create(ctx context.Context, user models.User) (models.User, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[user.ID] = user
	m.userByEmail[strings.ToLower(strings.TrimSpace(user.Email))] = user.ID
	return user, nil
}

func (m *memStore) GetByID(ctx context.Context, id uuid.UUID) (models.User, bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[id]
	return user, ok, nil
}

func (m *memStore) GetByEmail(ctx context.Context, email string) (models.User, bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	id, ok := m.userByEmail[strings.ToLower(strings.TrimSpace(email))]
	if ok == false {
		return models.User{}, false, nil
	}
	user, exists := m.users[id]
	return user, exists, nil
}

func (m *memStore) GetList(ctx context.Context) ([]models.User, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	return mapValuesSorted(m.users, func(left, right models.User) bool {
		return left.CreatedAtUTC.After(right.CreatedAtUTC)
	}), nil
}

func (m *memStore) CreateRoom(ctx context.Context, room models.Room) (models.Room, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rooms[room.ID] = room
	return room, nil
}

func (m *memStore) CreateSchedule(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	m.schedulesByRoom[schedule.RoomID] = schedule
	return schedule, nil
}

func (m *memStore) GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	schedule, ok := m.schedulesByRoom[roomID]
	return schedule, ok, nil
}

func (m *memStore) CreateBooking(ctx context.Context, booking models.Booking) (models.Booking, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bookings[booking.ID] = booking
	return booking, nil
}

func (m *memStore) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	booking, ok := m.bookings[bookingID]
	return booking, ok, nil
}

func (m *memStore) CancelBooking(ctx context.Context, bookingID uuid.UUID) (models.Booking, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	booking := m.bookings[bookingID]
	booking.Cancel(time.Now().UTC())
	m.bookings[bookingID] = booking
	return booking, nil
}

func (m *memStore) ListAllBookings(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	items := make([]models.Booking, 0, len(m.bookings))
	for _, b := range m.bookings {
		items = append(items, b)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedatUTC.After(items[j].CreatedatUTC)
	})

	total := len(items)
	start := (page - 1) * pageSize
	if start > len(items) {
		return []models.Booking{}, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (m *memStore) ListMyFutureBookings(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	items := make([]models.Booking, 0, len(m.bookings))
	for _, b := range m.bookings {
		if b.UserID != userID || b.Status != models.BookingStatusActive {
			continue
		}
		slot, ok := m.slots[b.SlotID]
		if ok == false || slot.StartUTC.Before(now.UTC()) {
			continue
		}
		items = append(items, b)
	}
	sort.Slice(items, func(i, j int) bool {
		slotI := m.slots[items[i].SlotID]
		slotJ := m.slots[items[j].SlotID]
		return slotI.StartUTC.Before(slotJ.StartUTC)
	})
	return items, nil
}

func (m *memStore) ExistsActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, b := range m.bookings {
		if b.SlotID == slotID && b.Status == models.BookingStatusActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *memStore) CreateRoomRepo(ctx context.Context, room models.Room) (models.Room, error) {
	return m.CreateRoom(ctx, room)
}

func (m *memStore) GetListRooms(ctx context.Context) ([]models.Room, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	return mapValuesSorted(m.rooms, func(left, right models.Room) bool {
		return left.CreatedAtUtc.After(right.CreatedAtUtc)
	}), nil
}

func (m *memStore) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.rooms[roomID]
	return ok, nil
}

func (m *memStore) UpsertMany(ctx context.Context, slots []models.Slot) error {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, slot := range slots {
		m.slots[slot.ID] = slot
	}
	return nil
}

func (m *memStore) ListFreeByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	dayStart := time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	activeBySlot := map[uuid.UUID]struct{}{}
	for _, booking := range m.bookings {
		if booking.Status == models.BookingStatusActive {
			activeBySlot[booking.SlotID] = struct{}{}
		}
	}

	out := make([]models.Slot, 0, 16)
	for _, slot := range m.slots {
		if slot.RoomID != roomID {
			continue
		}
		if slot.StartUTC.Before(dayStart) || slot.StartUTC.After(dayEnd) || slot.StartUTC.Equal(dayEnd) {
			continue
		}
		if _, booked := activeBySlot[slot.ID]; booked {
			continue
		}
		out = append(out, slot)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartUTC.Before(out[j].StartUTC)
	})
	return out, nil
}

func (m *memStore) GetSlotByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	slot, ok := m.slots[slotID]
	return slot, ok, nil
}

type usersRepoAdapter struct{ s *memStore }

func (a usersRepoAdapter) Create(ctx context.Context, user models.User) (models.User, error) {
	return a.s.Create(ctx, user)
}

func (a usersRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (models.User, bool, error) {
	return a.s.GetByID(ctx, id)
}

func (a usersRepoAdapter) GetByEmail(ctx context.Context, email string) (models.User, bool, error) {
	return a.s.GetByEmail(ctx, email)
}

func (a usersRepoAdapter) GetList(ctx context.Context) ([]models.User, error) {
	return a.s.GetList(ctx)
}

type roomsRepoAdapter struct{ s *memStore }

func (a roomsRepoAdapter) Create(ctx context.Context, room models.Room) (models.Room, error) {
	return a.s.CreateRoomRepo(ctx, room)
}

func (a roomsRepoAdapter) GetList(ctx context.Context) ([]models.Room, error) {
	return a.s.GetListRooms(ctx)
}

func (a roomsRepoAdapter) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	return a.s.Exists(ctx, roomID)
}

type schedulesRepoAdapter struct{ s *memStore }

func (a schedulesRepoAdapter) Create(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	return a.s.CreateSchedule(ctx, schedule)
}

func (a schedulesRepoAdapter) GetByRoomID(ctx context.Context, roomID uuid.UUID) (models.Schedule, bool, error) {
	return a.s.GetByRoomID(ctx, roomID)
}

type slotsRepoAdapter struct{ s *memStore }

func (a slotsRepoAdapter) UpsertMany(ctx context.Context, slots []models.Slot) error {
	return a.s.UpsertMany(ctx, slots)
}

func (a slotsRepoAdapter) ListFreeByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	return a.s.ListFreeByRoomAndDate(ctx, roomID, date)
}

func (a slotsRepoAdapter) GetByID(ctx context.Context, slotID uuid.UUID) (models.Slot, bool, error) {
	return a.s.GetSlotByID(ctx, slotID)
}

type bookingsRepoAdapter struct{ s *memStore }

func (a bookingsRepoAdapter) Create(ctx context.Context, booking models.Booking) (models.Booking, error) {
	return a.s.CreateBooking(ctx, booking)
}

func (a bookingsRepoAdapter) GetByID(ctx context.Context, bookingID uuid.UUID) (models.Booking, bool, error) {
	return a.s.GetBookingByID(ctx, bookingID)
}

func (a bookingsRepoAdapter) Cancel(ctx context.Context, bookingID uuid.UUID) (models.Booking, error) {
	return a.s.CancelBooking(ctx, bookingID)
}

func (a bookingsRepoAdapter) ListAll(ctx context.Context, page, pageSize int) ([]models.Booking, int, error) {
	return a.s.ListAllBookings(ctx, page, pageSize)
}

func (a bookingsRepoAdapter) ListMyFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Booking, error) {
	return a.s.ListMyFutureBookings(ctx, userID, now)
}

func (a bookingsRepoAdapter) ExistsActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error) {
	return a.s.ExistsActiveBySlotID(ctx, slotID)
}

type controllerTestEnv struct {
	server *httptest.Server
}

func newControllerTestEnv(t *testing.T) controllerTestEnv {
	t.Helper()

	store := newMemStore()
	usersSvc := services.NewUsersService(usersRepoAdapter{s: store})
	roomsSvc := services.NewRoomsService(roomsRepoAdapter{s: store})
	schedulesSvc := services.NewSchedulesService(schedulesRepoAdapter{s: store}, roomsSvc)
	slotsRepo := slotsRepoAdapter{s: store}
	slotsSvc := services.NewSlotsService(roomsSvc, schedulesSvc, slotsRepo)
	bookingsSvc := services.NewBookingsService(bookingsRepoAdapter{s: store}, slotsSvc)

	handler := NewRouter(
		usersSvc,
		roomsSvc,
		schedulesSvc,
		slotsSvc,
		bookingsSvc,
		slotsRepo,
		services.NewDefaultSlotsPlanner(),
		auth.NewTokenManager("test-secret", time.Hour),
	)

	return controllerTestEnv{server: httptest.NewServer(handler)}
}

func doJSONRequest(t *testing.T, method string, url string, token string, payload any) *http.Response {
	t.Helper()

	var body []byte
	var err error
	if payload == nil {
		body = []byte{}
	} else {
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}

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

func mapValuesSorted[T any](input map[uuid.UUID]T, less func(left, right T) bool) []T {
	out := make([]T, 0, len(input))
	for _, item := range input {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return less(out[i], out[j])
	})
	return out
}

func scheduleCreatePayload(roomID uuid.UUID) map[string]any {
	return map[string]any{
		"roomId":     roomID,
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}
}

func closeBody(t *testing.T, res *http.Response) {
	t.Helper()
	require.NotNil(t, res)
	require.NotNil(t, res.Body)
	require.NoError(t, res.Body.Close())
}

func mustDummyToken(t *testing.T, baseURL string, role string) string {
	t.Helper()

	res := doJSONRequest(t, http.MethodPost, baseURL+"/dummyLogin", "", map[string]string{"role": role})
	defer closeBody(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var out struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	require.NotEmpty(t, out.Token)
	return out.Token
}

func TestControllersInfoAndSwagger(t *testing.T) {
	env := newControllerTestEnv(t)
	defer env.server.Close()

	res, err := http.Get(env.server.URL + "/_info")
	require.NoError(t, err)
	defer closeBody(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.Get(env.server.URL + "/swagger")
	require.NoError(t, err)
	defer closeBody(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestControllersRegisterLoginAndAuth(t *testing.T) {
	env := newControllerTestEnv(t)
	defer env.server.Close()

	badRegister := doJSONRequest(t, http.MethodPost, env.server.URL+"/register", "", map[string]string{
		"email":    "user@example.com",
		"password": "secret",
		"role":     "nope",
	})
	defer closeBody(t, badRegister)
	require.Equal(t, http.StatusBadRequest, badRegister.StatusCode)

	register := doJSONRequest(t, http.MethodPost, env.server.URL+"/register", "", map[string]string{
		"email":    "user@example.com",
		"password": "secret",
		"role":     "user",
	})
	defer closeBody(t, register)
	require.Equal(t, http.StatusCreated, register.StatusCode)

	loginBad := doJSONRequest(t, http.MethodPost, env.server.URL+"/login", "", map[string]string{
		"email":    "user@example.com",
		"password": "wrong",
	})
	defer closeBody(t, loginBad)
	require.Equal(t, http.StatusUnauthorized, loginBad.StatusCode)

	login := doJSONRequest(t, http.MethodPost, env.server.URL+"/login", "", map[string]string{
		"email":    "user@example.com",
		"password": "secret",
	})
	defer closeBody(t, login)
	require.Equal(t, http.StatusOK, login.StatusCode)

	var tokenResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(login.Body).Decode(&tokenResp))
	require.NotEmpty(t, tokenResp.Token)

	noAuthRooms, err := http.Get(env.server.URL + "/rooms/list")
	require.NoError(t, err)
	defer closeBody(t, noAuthRooms)
	require.Equal(t, http.StatusUnauthorized, noAuthRooms.StatusCode)
}

func TestControllersRoleProtectedFlow(t *testing.T) {
	env := newControllerTestEnv(t)
	defer env.server.Close()

	adminToken := mustDummyToken(t, env.server.URL, "admin")
	userToken := mustDummyToken(t, env.server.URL, "user")

	userCreateRoom := doJSONRequest(t, http.MethodPost, env.server.URL+"/rooms/create", userToken, map[string]any{
		"name": "User should fail",
	})
	defer closeBody(t, userCreateRoom)
	require.Equal(t, http.StatusForbidden, userCreateRoom.StatusCode)

	adminCreateRoom := doJSONRequest(t, http.MethodPost, env.server.URL+"/rooms/create", adminToken, map[string]any{
		"name":        "Alpha",
		"description": "Room A",
		"capacity":    6,
	})
	defer closeBody(t, adminCreateRoom)
	require.Equal(t, http.StatusCreated, adminCreateRoom.StatusCode)

	var roomOut struct {
		Room struct {
			ID uuid.UUID `json:"id"`
		} `json:"room"`
	}
	require.NoError(t, json.NewDecoder(adminCreateRoom.Body).Decode(&roomOut))

	roomsList := doJSONRequest(t, http.MethodGet, env.server.URL+"/rooms/list", userToken, nil)
	defer closeBody(t, roomsList)
	require.Equal(t, http.StatusOK, roomsList.StatusCode)

	userSchedule := doJSONRequest(t, http.MethodPost, env.server.URL+"/rooms/"+roomOut.Room.ID.String()+"/schedule/create", userToken, scheduleCreatePayload(roomOut.Room.ID))
	defer closeBody(t, userSchedule)
	require.Equal(t, http.StatusForbidden, userSchedule.StatusCode)

	adminSchedule := doJSONRequest(t, http.MethodPost, env.server.URL+"/rooms/"+roomOut.Room.ID.String()+"/schedule/create", adminToken, scheduleCreatePayload(roomOut.Room.ID))
	defer closeBody(t, adminSchedule)
	require.Equal(t, http.StatusCreated, adminSchedule.StatusCode)

	date := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	slotsList := doJSONRequest(t, http.MethodGet, env.server.URL+"/rooms/"+roomOut.Room.ID.String()+"/slots/list?date="+date, userToken, nil)
	defer closeBody(t, slotsList)
	require.Equal(t, http.StatusOK, slotsList.StatusCode)

	var slotsOut struct {
		Slots []struct {
			ID uuid.UUID `json:"id"`
		} `json:"slots"`
	}
	require.NoError(t, json.NewDecoder(slotsList.Body).Decode(&slotsOut))
	require.NotEmpty(t, slotsOut.Slots)

	adminCreateBooking := doJSONRequest(t, http.MethodPost, env.server.URL+"/bookings/create", adminToken, map[string]any{
		"slotId": slotsOut.Slots[0].ID,
	})
	defer closeBody(t, adminCreateBooking)
	require.Equal(t, http.StatusForbidden, adminCreateBooking.StatusCode)

	userCreateBooking := doJSONRequest(t, http.MethodPost, env.server.URL+"/bookings/create", userToken, map[string]any{
		"slotId":               slotsOut.Slots[0].ID,
		"createConferenceLink": true,
	})
	defer closeBody(t, userCreateBooking)
	require.Equal(t, http.StatusCreated, userCreateBooking.StatusCode)

	var bookingOut struct {
		Booking struct {
			ID uuid.UUID `json:"id"`
		} `json:"booking"`
	}
	require.NoError(t, json.NewDecoder(userCreateBooking.Body).Decode(&bookingOut))

	adminMyBookings := doJSONRequest(t, http.MethodGet, env.server.URL+"/bookings/my", adminToken, nil)
	defer closeBody(t, adminMyBookings)
	require.Equal(t, http.StatusForbidden, adminMyBookings.StatusCode)

	userMyBookings := doJSONRequest(t, http.MethodGet, env.server.URL+"/bookings/my", userToken, nil)
	defer closeBody(t, userMyBookings)
	require.Equal(t, http.StatusOK, userMyBookings.StatusCode)

	userBookingsList := doJSONRequest(t, http.MethodGet, env.server.URL+"/bookings/list?page=1&pageSize=20", userToken, nil)
	defer closeBody(t, userBookingsList)
	require.Equal(t, http.StatusForbidden, userBookingsList.StatusCode)

	adminBookingsList := doJSONRequest(t, http.MethodGet, env.server.URL+"/bookings/list?page=1&pageSize=20", adminToken, nil)
	defer closeBody(t, adminBookingsList)
	require.Equal(t, http.StatusOK, adminBookingsList.StatusCode)

	cancel1 := doJSONRequest(t, http.MethodPost, env.server.URL+"/bookings/"+bookingOut.Booking.ID.String()+"/cancel", userToken, map[string]any{})
	defer closeBody(t, cancel1)
	require.Equal(t, http.StatusOK, cancel1.StatusCode)

	cancel2 := doJSONRequest(t, http.MethodPost, env.server.URL+"/bookings/"+bookingOut.Booking.ID.String()+"/cancel", userToken, map[string]any{})
	defer closeBody(t, cancel2)
	require.Equal(t, http.StatusOK, cancel2.StatusCode)
}
