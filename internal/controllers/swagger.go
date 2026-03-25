package controllers

import (
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func swaggerOpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	for _, candidate := range []string{"api.yaml", "../../api.yaml"} {
		if _, err := os.Stat(candidate); err == nil {
			http.ServeFile(w, r, candidate)
			return
		}
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "openapi spec is not available")
}

func swaggerUIHandler() http.Handler {
	return httpSwagger.Handler(httpSwagger.URL("/swagger/openapi.yaml"))
}
