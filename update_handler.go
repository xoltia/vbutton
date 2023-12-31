package vbutton

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type UpdateHandler struct {
	vc *VoiceClipService
}

func NewUpdateHandler(vc *VoiceClipService) *UpdateHandler {
	return &UpdateHandler{vc}
}

func (h *UpdateHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	idString := r.FormValue("id")

	if idString == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")

	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	vtuber := r.FormValue("vtuber")

	if vtuber == "" {
		http.Error(w, "vtuber is required", http.StatusBadRequest)
		return
	}

	agency := r.FormValue("agency")
	tags := strings.Split(r.FormValue("tags"), ",")
	refURL := r.FormValue("url")
	approved := r.FormValue("approved") == "true"
	rejected := r.FormValue("rejected") == "true"

	clip := &VoiceClip{
		ID:           id,
		Title:        title,
		VTuberName:   vtuber,
		AgencyName:   sql.NullString{String: agency, Valid: agency != ""},
		Tags:         tags,
		ReferenceURL: sql.NullString{String: refURL, Valid: refURL != ""},
	}

	if approved {
		clip.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	if rejected {
		h.vc.DeleteVoiceClip(id)
		http.Redirect(w, r, "/update", http.StatusSeeOther)
		return
	}

	err = h.vc.UpdateVoiceClip(clip)

	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/update?id="+idString, http.StatusSeeOther)

}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()

	adminUsername := os.Getenv("VB_ADMIN_USERNAME")
	adminPassword := os.Getenv("VB_ADMIN_PASSWORD")

	if !ok || password != adminPassword || username != adminUsername {
		http.Header.Add(w.Header(), "WWW-Authenticate", `Basic realm="vbutton"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		h.handlePost(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idString := r.URL.Query().Get("id")

	if idString == "" {
		clips, err := h.vc.GetUnapprovedVoiceClips(time.Hour * 24 * 7)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		overviewPage(clips).Render(r.Context(), w)
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
