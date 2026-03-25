package controllers

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func swaggerUIHandler() http.Handler {
	return httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json"))
}
