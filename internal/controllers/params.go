package controllers

import "github.com/google/uuid"

func parsePathUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}
