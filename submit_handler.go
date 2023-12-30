package vbutton

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
)

type SubmitHandler struct {
	vc *VoiceClipService
}

func NewSubmitHandler(vc *VoiceClipService) *SubmitHandler {
	return &SubmitHandler{vc}
}

func (h *SubmitHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	submitPage().Render(r.Context(), w)
}

func (h *SubmitHandler) servePOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	if header.Size > 32<<20 {
		http.Error(w, "file too large", http.StatusBadRequest)
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

	clip := &VoiceClip{
		Title:        title,
		VTuberName:   vtuber,
		AgencyName:   sql.NullString{String: agency, Valid: agency != ""},
		Tags:         tags,
		ReferenceURL: sql.NullString{String: refURL, Valid: refURL != ""},
	}

	err = h.vc.CreateVoiceClip(clip, file)

	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/submit", http.StatusSeeOther)
}

func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveGET(w, r)
	case http.MethodPost:
		h.servePOST(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
