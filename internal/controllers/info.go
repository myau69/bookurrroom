package controllers

import "net/http"

// @Summary Service health
// @Tags info
// @Produce json
// @Success 200 {object} infoResponse
// @Router /_info [get]
func infoHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
