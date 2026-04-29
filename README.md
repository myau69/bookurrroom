# Сервис бронирования переговорок

## Назначение проекта
Сервис решает задачу единого бронирования переговорок:
- администратор создаёт переговорки;
- администратор задаёт расписание доступности переговорки (дни недели и время);
- система автоматически формирует слоты по расписанию;
- пользователь выбирает свободный слот и создаёт/отменяет бронь.

## Зависимости проекта
### Стек
- Go `1.26.1`
- PostgreSQL `16+`
- Docker
- Docker Compose
- GNU Make
- K6

### Библиотеки
- `github.com/go-chi/chi/v5` — HTTP-роутинг
- `github.com/golang-jwt/jwt/v5` — JWT-токены
- `github.com/golang-migrate/migrate/v4` — миграции БД
- `github.com/google/uuid` — UUID
- `github.com/lib/pq` — PostgreSQL драйвер
- `github.com/swaggo/http-swagger/v2` и `github.com/swaggo/swag` — Swagger UI и OpenAPI
- `golang.org/x/crypto` — хеширование паролей
- `github.com/stretchr/testify` — тестовые ассерты



## Структура проекта
- `cmd/app` — запуск HTTP API.
- `cmd/seed` — заполнение БД тестовыми данными.
- `internal/app` — сборка зависимостей приложения.
- `internal/config` — конфигурация из переменных окружения.
- `internal/controllers` — HTTP-хендлеры, роутинг, middleware авторизации.
- `internal/services` — бизнес-логика.
- `internal/models` — доменные сущности и валидация.
- `internal/repository` — интерфейсы доступа к данным.
- `internal/repository/postgres` — PostgreSQL-реализация репозиториев.
- `internal/bootstrap` — запуск миграций.
- `migrations` — SQL-схема и изменения БД.
- `tests` — E2E и нагрузочные сценарии.

## Бизнес-ограничения
- Один слот может иметь только одну активную бронь.
- Слоты одной переговорки не пересекаются по времени.
- Если расписание для переговорки не создано, слоты недоступны.
- Время хранится и передаётся в UTC.
- `admin` не может создавать брони.
- Отмена брони идемпотентна: повторный вызов возвращает актуальное состояние с `200`.
- Нельзя создать бронь на слот из прошлого (`400`).
- `/bookings/my` возвращает только будущие брони.
- Длительность слота фиксирована: 30 минут.

## Генерация слотов
Используется окно ближайших 7 дней:
- при создании расписания слоты генерируются на период `today..today+6`;
- слоты сохраняются в БД со стабильными UUID;
- список доступных слотов запрашивается по `roomId + date`.

## API и роли доступа
### Публичные ручки
- `POST /dummyLogin` — получить тестовый JWT по роли (`admin`/`user`).
- `GET /_info` — health endpoint (`200`).

### Ручки для `admin`
- `POST /rooms/create` — создать переговорку.
- `POST /rooms/{roomId}/schedule/create` — создать расписание переговорки (один раз).
- `GET /bookings/list` — список всех броней с пагинацией.

### Ручки для `admin` и `user`
- `GET /rooms/list` — список переговорок.
- `GET /rooms/{roomId}/slots/list?date=YYYY-MM-DD` — свободные слоты на дату.

### Ручки для `user`
- `POST /bookings/create` — создать бронь от имени пользователя из JWT.
- `POST /bookings/{bookingId}/cancel` — отменить бронь.
- `GET /bookings/my` — будущие брони текущего пользователя.

## Дополнительные задания
- Регистрация и логин по email/паролю: `POST /register`, `POST /login`.
- Опциональная ссылка на конференцию в брони через `POST /bookings/create` и `createConferenceLink` (`true` → `conferenceLink` в ответе, `false` → `null`).
- Swagger-документация: `/swagger`, `/swagger/doc.json`.
- Нагрузочный сценарий для списка слотов (`k6`).
- Результат нагрузочного теста (`/rooms/{roomId}/slots/list`, профиль `100 RPS`, `60s`): `http_req_failed=0.00%`, `p(95)=6.09ms` (порог `p(95)<200ms` выполнен с запасом).
- Конфигурация линтера в `.golangci.yaml` 


## Запуск проекта
### Вариант 1: через Makefile
Запуск:
```bash
make up
```

Наполнение тестовыми данными:
```bash
make seed
```

Остановка с очисткой данных:
```bash
make down
```

### Вариант 2: напрямую через Docker Compose
Запуск:
```bash
docker compose --env-file .env.example up --build --remove-orphans
```

Остановка с очисткой данных:
```bash
docker compose --env-file .env.example down -v --remove-orphans
```

### Вариант 3: ручной запуск (без Docker)
1. Поднять PostgreSQL локально.
2. Создать БД `booking` и пользователя `booking` (или использовать свои значения).
3. Установить переменные окружения и запустить приложение:

```bash
export HTTP_PORT='8080'
export DATABASE_URL='postgres://booking:booking@localhost:5432/booking?sslmode=disable'
export MIGRATIONS_PATH='migrations'
export JWT_SECRET='dev-secret'
export JWT_TTL_HOURS='24'
go run ./cmd/app
```

Приложение само применяет миграции на старте.

### Проверка после запуска
```bash
curl -i http://localhost:8080/_info
```

Swagger:
- `http://localhost:8080/swagger`
- `http://localhost:8080/swagger/index.html`
- `http://localhost:8080/swagger/doc.json`

## Тестирование
```bash
make test
make cover
make e2e-up
make e2e
make e2e-down
```

## Команды Makefile
- `make up`
- `make down`
- `make seed`
- `make test`
- `make cover`
- `make e2e-up`
- `make e2e`
- `make e2e-down`

## Нагрузочное тестирование (инструкция)
1. Поднять сервис: `make up`.
2. Получить `admin` токен через `POST /dummyLogin`.
3. Создать переговорку и расписание (чтобы появились слоты).
4. Получить `user` токен через `POST /dummyLogin`.
5. Запустить `k6`:

```bash
DATE=$(date -u -v+1d +%F) \
BASE_URL=http://localhost:8080 \
ROOM_ID=<room_uuid> \
TOKEN=<user_jwt> \
k6 run tests/load/slots_list.js
```

6. Опционально сохранить summary в файл:

```bash
DATE=$(date -u -v+1d +%F) \
BASE_URL=http://localhost:8080 \
ROOM_ID=<room_uuid> \
TOKEN=<user_jwt> \
k6 run --summary-export k6-summary.json tests/load/slots_list.js
```

## Линтер (запуск)
```bash
golangci-lint run ./...
```
