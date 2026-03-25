package controllers

import (
	"bookurrroom/internal/auth"
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository"
	"bookurrroom/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	usersSvc *services.UsersService,
	roomsSvc *services.RoomsService,
	schedulesSvc *services.SchedulesService,
	slotsSvc *services.SlotsService,
	bookingsSvc *services.BookingsService,

	slotsRepo repository.SlotsRepository,
	planner services.SlotsPlanner,
	tokens *auth.TokenManager,
) http.Handler {
	r := chi.NewRouter()

	authHandler := NewAuthHandler(usersSvc, tokens)
	roomsHandler := NewRoomsHandler(roomsSvc)
	schedulesHandler := NewSchedulesHandler(schedulesSvc, slotsRepo, planner)
	slotsHandler := NewSlotsHandler(slotsSvc)
	bookingsHandler := NewBookingsHandler(bookingsSvc)

	r.HandleFunc("/_info", infoHandler)
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusTemporaryRedirect)
	})
	r.Get("/swagger/openapi.yaml", swaggerOpenAPIHandler)
	r.Handle("/swagger/*", swaggerUIHandler())

	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	r.Post("/dummyLogin", authHandler.DummyLogin)

	r.Group(func(r chi.Router) {
		r.Use(requireAuth(tokens))

		r.With(requireRoles(models.RoleAdmin, models.RoleUser)).Get("/rooms/list", roomsHandler.List)
		r.With(requireRoles(models.RoleAdmin)).Post("/rooms/create", roomsHandler.Create)
		r.With(requireRoles(models.RoleAdmin)).Post("/rooms/{roomId}/schedule/create", schedulesHandler.Create)
		r.With(requireRoles(models.RoleAdmin, models.RoleUser)).Get("/rooms/{roomId}/slots/list", slotsHandler.ListByRoomAndDate)

		r.With(requireRoles(models.RoleUser)).Post("/bookings/create", bookingsHandler.Create)
		r.With(requireRoles(models.RoleAdmin)).Get("/bookings/list", bookingsHandler.ListAll)
		r.With(requireRoles(models.RoleUser)).Get("/bookings/my", bookingsHandler.ListMy)
		r.With(requireRoles(models.RoleUser)).Post("/bookings/{bookingId}/cancel", bookingsHandler.Cancel)
	})

	return r
}
