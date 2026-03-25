package main

import (
	"bookurrroom/internal/app"
	"bookurrroom/internal/config"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "bookurrroom/internal/controllers/swaggerdocs"
)

// @title Booking API
// @version 1.0
// @description API для бронирования переговорок
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: Bearer <token>

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.Build(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	go func() {
		if err := application.Server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) == false {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
