package controllers

type authTokenResponse struct {
	Token string `json:"token"`
}

type registerResponse struct {
	User userResponse `json:"user"`
}

type roomsListResponse struct {
	Rooms []roomResponse `json:"rooms"`
}

type roomCreateResponse struct {
	Room roomResponse `json:"room"`
}

type scheduleCreateResponse struct {
	Schedule scheduleResponse `json:"schedule"`
}

type slotsListResponse struct {
	Slots []slotResponse `json:"slots"`
}

type bookingCreateResponse struct {
	Booking bookingResponse `json:"booking"`
}

type bookingsListResponse struct {
	Bookings   []bookingResponse  `json:"bookings"`
	Pagination paginationResponse `json:"pagination"`
}

type myBookingsResponse struct {
	Bookings []bookingResponse `json:"bookings"`
}

type infoResponse struct {
	Status string `json:"status"`
}
