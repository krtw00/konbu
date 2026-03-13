package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

type ExportHandler struct {
	exportSvc *service.ExportService
}

func NewExportHandler(exportSvc *service.ExportService) *ExportHandler {
	return &ExportHandler{exportSvc: exportSvc}
}

func (h *ExportHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/json", h.exportJSON)
	r.Get("/markdown", h.exportMarkdown)

	return r
}

func (h *ExportHandler) exportJSON(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	data, err := h.exportSvc.ExportJSON(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	filename := fmt.Sprintf("konbu-export-%s.json", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	json.NewEncoder(w).Encode(data)
}

func (h *ExportHandler) exportMarkdown(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	buf, err := h.exportSvc.ExportMarkdown(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	filename := fmt.Sprintf("konbu-export-%s.zip", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write(buf.Bytes())
}
