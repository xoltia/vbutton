package vbutton

import (
	"database/sql"
	"fmt"
	"net/http"
)

type IndexHandler struct {
	vc *VoiceClipService
}

func NewIndexHandler(vc *VoiceClipService) *IndexHandler {
	return &IndexHandler{vc}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	matchQuery := query.Get("q")
	vtuberQuery := query.Get("v")
	agencyQuery := query.Get("a")
	tagQuery := query.Get("t")

	var clips []*VoiceClip
	var err error

	if matchQuery != "" || vtuberQuery != "" || agencyQuery != "" || tagQuery != "" {
		clips, err = h.vc.SearchClips(
			sql.NullString{String: fmt.Sprintf("%%%s%%", matchQuery), Valid: matchQuery != ""},
			sql.NullString{String: vtuberQuery, Valid: vtuberQuery != ""},
			sql.NullString{String: agencyQuery, Valid: agencyQuery != ""},
			sql.NullString{String: tagQuery, Valid: tagQuery != ""},
			20,
		)
	} else {
		clips, err = h.vc.GetRecentVoiceClips(20)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tags, err := h.vc.GetTopTags(30)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vtubers, err := h.vc.GetTopVTubers(30)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agencies, err := h.vc.GetTopAgencies(30)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := IndexModel{
		Clips:    clips,
		Tags:     tags,
		VTubers:  vtubers,
		Agencies: agencies,
	}

	if matchQuery != "" {
		view.CurrentSearch = matchQuery
	}

	if vtuberQuery != "" {
		view.CurrentVTuber = vtuberQuery
	}

	if agencyQuery != "" {
		view.CurrentAgency = agencyQuery
	}

	if tagQuery != "" {
		view.CurrentTag = tagQuery
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexPage(view).Render(r.Context(), w)
}
