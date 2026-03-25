FROM golang:1.26.1 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/seed ./cmd/seed

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /out/seed /app/seed
COPY --from=builder /src/migrations /app/migrations

EXPOSE 8080

CMD ["/app/app"]