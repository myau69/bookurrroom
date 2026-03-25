package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserModel(t *testing.T) {
	now := time.Now().UTC()

	user, err := NewUser(uuid.New(), "USER@Example.com", RoleUser, now)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", user.Email)

	_, err = NewUser(uuid.Nil, "user@example.com", RoleUser, now)
	require.ErrorIs(t, err, ErrUserInvalidID)

	_, err = NewUser(uuid.New(), "bad-email", RoleUser, now)
	require.ErrorIs(t, err, ErrUserInvalidEmail)

	_, err = ParseRole("admin")
	require.NoError(t, err)
	_, err = ParseRole("unknown")
	require.ErrorIs(t, err, ErrUserInvalidRole)
}

func TestRoomModel(t *testing.T) {
	now := time.Now().UTC()
	capacity := 4
	description := "Room"

	room, err := NewRoom(uuid.New(), " Alpha ", &description, &capacity, now)
	require.NoError(t, err)
	require.Equal(t, "Alpha", room.Name)

	_, err = NewRoom(uuid.Nil, "Alpha", nil, nil, now)
	require.ErrorIs(t, err, ErrRoomInvalidID)

	_, err = NewRoom(uuid.New(), "   ", nil, nil, now)
	require.ErrorIs(t, err, ErrRoomNameRequired)

	invalidCapacity := 0
	_, err = NewRoom(uuid.New(), "Alpha", nil, &invalidCapacity, now)
	require.ErrorIs(t, err, ErrRoomInvalidCapacity)
}

func TestScheduleModel(t *testing.T) {
	now := time.Now().UTC()
	roomID := uuid.New()

	s, err := NewSchedule(uuid.New(), roomID, []int{3, 1, 2}, "09:00", "10:00", now)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, s.DaysOfWeek)

	_, err = NewSchedule(uuid.Nil, roomID, []int{1}, "09:00", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleInvalidID)

	_, err = NewSchedule(uuid.New(), uuid.Nil, []int{1}, "09:00", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleInvalidRoomID)

	_, err = NewSchedule(uuid.New(), roomID, nil, "09:00", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleEmptyDaysOfWeek)

	_, err = NewSchedule(uuid.New(), roomID, []int{8}, "09:00", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleInvalidDayOfWeek)

	_, err = NewSchedule(uuid.New(), roomID, []int{1, 1}, "09:00", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleDuplicateDayOfWeek)

	_, err = NewSchedule(uuid.New(), roomID, []int{1}, "bad", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleInvalidTimeFormat)

	_, err = NewSchedule(uuid.New(), roomID, []int{1}, "11:00", "10:00", now)
	require.ErrorIs(t, err, ErrScheduleInvalideTimeRange)

	start, end, err := s.TimeBounds()
	require.NoError(t, err)
	require.Equal(t, 9*time.Hour, start)
	require.Equal(t, 10*time.Hour, end)
}

func TestSlotModel(t *testing.T) {
	roomID := uuid.New()
	start := time.Now().UTC().Truncate(time.Minute)
	end := start.Add(SlotDuration)

	slot, err := NewSlot(uuid.New(), roomID, start, end)
	require.NoError(t, err)
	require.False(t, slot.IsPast(start.Add(-time.Minute)))

	_, err = NewSlot(uuid.Nil, roomID, start, end)
	require.ErrorIs(t, err, ErrSlotInvalidID)

	_, err = NewSlot(uuid.New(), roomID, start, start.Add(10*time.Minute))
	require.ErrorIs(t, err, ErrSlotInvalidDuration)

	localStart := time.Date(2026, 1, 1, 9, 0, 0, 0, time.FixedZone("UTC+3", 3*3600))
	localEnd := localStart.Add(SlotDuration)
	_, err = NewSlot(uuid.New(), roomID, localStart, localEnd)
	require.ErrorIs(t, err, ErrSlotTimeUTC)
}

func TestBookingModel(t *testing.T) {
	now := time.Now().UTC()
	slotID := uuid.New()
	userID := uuid.New()

	booking, err := NewBooking(uuid.New(), slotID, userID, now)
	require.NoError(t, err)
	require.Equal(t, BookingStatusActive, booking.Status)
	require.True(t, booking.IsOwnedBy(userID))

	_, err = NewBooking(uuid.Nil, slotID, userID, now)
	require.ErrorIs(t, err, ErrBookingInvalidID)

	_, err = ParseBookingStatus("active")
	require.NoError(t, err)
	_, err = ParseBookingStatus("unknown")
	require.ErrorIs(t, err, ErrBookingInvalidStatus)

	booking.Cancel(now)
	require.Equal(t, BookingStatusCancelled, booking.Status)
	require.NotNil(t, booking.CancelledAtUTC)

	cancelledAt := *booking.CancelledAtUTC
	booking.Cancel(now.Add(time.Hour))
	require.Equal(t, cancelledAt, *booking.CancelledAtUTC)
}
