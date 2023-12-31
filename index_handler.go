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

	view := IndexModel{
		Clips: clips,
		Tags: []string{
			"ポン", "挨拶", "センシティブ", "おやすみ",
			"ヤンデレ", "ツンデレ", "圧", "かわいい", "英語",
		},
		VTubers: []string{
			"博衣こより", "湊あくあ", "しぐれうい",
			"白上フブキ", "夏色まつり", "百鬼あやめ",
			"紫咲シオン", "癒月ちょこ", "大空スバル",
			"月ノ美兎", "樋口楓", "猫又おかゆ",
		},
		Agencies: []string{
			"ホロライブ", "にじさんじ", "ホロスターズ",
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexPage(view).Render(r.Context(), w)
}
