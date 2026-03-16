package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
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
	q := r.URL.Query()

	params := model.SearchParams{
		Query:  q.Get("q"),
		Tag:    q.Get("tag"),
		Limit:  20,
		Offset: 0,
	}

	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		params.Limit = l
	}
	if o, err := strconv.Atoi(q.Get("offset")); err == nil && o >= 0 {
		params.Offset = o
	}
	if t := q.Get("type"); t != "" {
		params.Types = strings.Split(t, ",")
	}
	if f := q.Get("from"); f != "" {
		if t, err := time.Parse(time.RFC3339, f); err == nil {
			params.From = &t
		}
	}
	if t := q.Get("to"); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			params.To = &parsed
		}
	}

	resp, err := h.searchSvc.SearchAdvanced(r.Context(), user.ID, params)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
