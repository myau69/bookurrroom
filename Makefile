APP_ENV ?= .env.example
DATABASE_URL ?= postgres://booking:booking@localhost:5432/booking?sslmode=disable
MIGRATIONS_PATH ?= migrations

.PHONY: up down seed test cover e2e e2e-up e2e-down

up:
	docker compose --env-file $(APP_ENV) up --build --remove-orphans

down:
	docker compose --env-file $(APP_ENV) down -v --remove-orphans

seed:
	DATABASE_URL="$(DATABASE_URL)" MIGRATIONS_PATH="$(MIGRATIONS_PATH)" go run ./cmd/seed

test:
	go test ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

e2e-up:
	docker compose -f docker-compose.e2e.yaml up -d --remove-orphans

e2e-down:
	docker compose -f docker-compose.e2e.yaml down -v --remove-orphans

e2e:
	TEST_DATABASE_URL=postgres://booking:booking@127.0.0.1:55432/booking_e2e?sslmode=disable \
	MIGRATIONS_PATH=../migrations \
	go test ./tests -run TestE2E -v