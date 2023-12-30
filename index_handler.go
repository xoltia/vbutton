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
	clips, err := h.vc.GetRecentVoiceClips(100)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := IndexModel{
		Clips: clips,
	}

	indexPage(view).Render(r.Context(), w)
}
