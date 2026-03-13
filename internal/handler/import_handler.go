package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

type ImportHandler struct {
	importSvc *service.ImportService
}

func NewImportHandler(eventSvc *service.EventService) *ImportHandler {
	return &ImportHandler{importSvc: service.NewImportService(eventSvc)}
}

func (h *ImportHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/ical", h.importICal)

	return r
}

func (h *ImportHandler) importICal(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, err)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, err)
		return
	}
	defer file.Close()

	events, err := h.importSvc.ImportICal(r.Context(), user.ID, file)
	if err != nil {
		writeError(w, err)
		return
	}

	writeCreated(w, events)
}
