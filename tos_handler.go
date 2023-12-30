package vbutton

import "net/http"

type TOSHandler struct{}

func NewTOSHandler() *TOSHandler {
	return &TOSHandler{}
}

func (h *TOSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tosPage().Render(r.Context(), w)
}
