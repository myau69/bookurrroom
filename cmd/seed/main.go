package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"bookurrroom/internal/bootstrap"
	"bookurrroom/internal/models"
	"bookurrroom/internal/repository/postgres"
	"bookurrroom/internal/services"
)

func main() {
	databaseURL := getEnv("DATABASE_URL", "postgres://booking:booking@localhost:5432/booking?sslmode=disable")
	migrationsPath := getEnv("MIGRATIONS_PATH", "migrations")

	if err := bootstrap.RunMigrations(databaseURL, migrationsPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	usersRepo := postgres.NewUsersPostgresRepository(db)
	roomsRepo := postgres.NewRoomsPostgresRepository(db)
	schedulesRepo := postgres.NewSchedulesPostgresRepository(db)
	slotsRepo := postgres.NewSlotsPostgresRepository(db)

	usersSvc := services.NewUsersService(usersRepo)
	roomsSvc := services.NewRoomsService(roomsRepo)
	schedulesSvc := services.NewSchedulesService(schedulesRepo, roomsSvc)
	planner := services.NewDefaultSlotsPlanner()

	if _, found, err := usersSvc.GetByEmail(ctx, "employee@example.com"); err != nil {
		log.Fatalf("check seed user: %v", err)
	} else if !found {
		if _, err := usersSvc.Create(ctx, "employee@example.com", models.RoleUser, nil); err != nil {
			log.Fatalf("create seed user: %v", err)
		}
	}

	rooms, err := roomsSvc.GetList(ctx)
	if err != nil {
		log.Fatalf("list rooms: %v", err)
	}

	if len(rooms) > 0 {
		log.Printf("seed skipped: rooms already exist (%d)", len(rooms))
		return
	}

	room, err := roomsSvc.Create(ctx, "Alpha", strPtr("Seeded room"), intPtr(8))
	if err != nil {
		log.Fatalf("create room: %v", err)
	}

	schedule, err := schedulesSvc.Create(ctx, services.SchedulesCreateInput{
		RoomID:     room.ID,
		DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
		StartTime:  "09:00",
		EndTime:    "21:00",
	})
	if err != nil {
		log.Fatalf("create schedule: %v", err)
	}

	from := time.Now().UTC()
	to := from.AddDate(0, 0, 6)
	generatedSlots, err := planner.BuildWindow(schedule.RoomID, services.SlotsScheduleRule{
		DaysOfWeek: schedule.DaysOfWeek,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}, from, to)
	if err != nil {
		log.Fatalf("build slots: %v", err)
	}
	if err := slotsRepo.UpsertMany(ctx, generatedSlots); err != nil {
		log.Fatalf("save slots: %v", err)
	}

	log.Printf("seed done: room=%s slots=%d", room.ID, len(generatedSlots))
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}
