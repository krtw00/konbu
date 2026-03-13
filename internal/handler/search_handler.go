package handler

import (
	"net/http"
	"strconv"

	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

type SearchHandler struct {
	searchSvc *service.SearchService
}

func NewSearchHandler(searchSvc *service.SearchService) *SearchHandler {
	return &SearchHandler{searchSvc: searchSvc}
}

func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	query := r.URL.Query().Get("q")
	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	results, err := h.searchSvc.Search(r.Context(), user.ID, query, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, results)
}
