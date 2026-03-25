package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"bookurrroom/internal/auth"
	"bookurrroom/internal/bootstrap"
	"bookurrroom/internal/config"
	"bookurrroom/internal/controllers"
	"bookurrroom/internal/repository/postgres"
	"bookurrroom/internal/services"
)

type Application struct {
	Server    *http.Server
	closeFns  []func() error
	closeName []string
}

func Build(cfg config.Config) (*Application, error) {
	if err := bootstrap.RunMigrations(cfg.Database.URL, cfg.Migrations.Path); err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		return nil, err
	}

	pingCtx, cancelPing := context.WithTimeout(context.Background(), cfg.ReadinessPing.DBPingTimeout)
	err = db.PingContext(pingCtx)
	cancelPing()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	usersRepo := postgres.NewUsersPostgresRepository(db)
	roomsRepo := postgres.NewRoomsPostgresRepository(db)
	schedulesRepo := postgres.NewSchedulesPostgresRepository(db)
	slotsRepo := postgres.NewSlotsPostgresRepository(db)
	bookingsRepo := postgres.NewBookingsPostgresRepository(db)

	usersSvc := services.NewUsersService(usersRepo)
	roomsSvc := services.NewRoomsService(roomsRepo)
	schedulesSvc := services.NewSchedulesService(schedulesRepo, roomsSvc)

	closeFns := []func() error{db.Close}
	closeNames := []string{"postgres"}

	slotsSvc := services.NewSlotsService(roomsSvc, schedulesSvc, slotsRepo)
	bookingsSvc := services.NewBookingsService(bookingsRepo, slotsSvc)

	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
	handler := controllers.NewRouter(
		usersSvc,
		roomsSvc,
		schedulesSvc,
		slotsSvc,
		bookingsSvc,
		slotsRepo,
		services.NewDefaultSlotsPlanner(),
		tokens,
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Application{
		Server:    server,
		closeFns:  closeFns,
		closeName: closeNames,
	}, nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	var allErr error

	if err := a.Server.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) == false {
		allErr = errors.Join(allErr, err)
	}

	for i := len(a.closeFns) - 1; i >= 0; i-- {
		if err := a.closeFns[i](); err != nil {
			allErr = errors.Join(allErr, err)
		}
	}

	return allErr
}
