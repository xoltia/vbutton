package vbutton

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type UpdateHandler struct {
	vc *VoiceClipService
}

func NewUpdateHandler(vc *VoiceClipService) *UpdateHandler {
	return &UpdateHandler{vc}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, password, ok := r.BasicAuth()

	adminUsername := os.Getenv("VB_ADMIN_USERNAME")
	adminPassword := os.Getenv("VB_ADMIN_PASSWORD")

	fmt.Println(adminUsername, adminPassword)
	fmt.Println(username, password)

	if !ok || password != adminPassword || username != adminUsername {
		http.Header.Add(w.Header(), "WWW-Authenticate", `Basic realm="vbutton"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idString := r.URL.Query().Get("id")

	if idString == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clip, err := h.vc.GetVoiceClip(id)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "voice clip not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	updatePage(clip).Render(r.Context(), w)
}
