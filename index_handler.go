package vbutton

import (
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

	clips, err := h.vc.GetRecentVoiceClips(100)

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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexPage(view).Render(r.Context(), w)
}
