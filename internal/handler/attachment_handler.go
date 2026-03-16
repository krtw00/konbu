package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
)

type AttachmentHandler struct {
	r2Svc *service.R2Service
}

func NewAttachmentHandler(r2Svc *service.R2Service) *AttachmentHandler {
	return &AttachmentHandler{r2Svc: r2Svc}
}

func (h *AttachmentHandler) UploadRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.upload)
	return r
}

func (h *AttachmentHandler) upload(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user.Plan != "sponsor" && !user.IsAdmin {
		writeError(w, apperror.Forbidden("attachments require a sponsor plan"))
		return
	}

	if h.r2Svc == nil {
		writeError(w, apperror.Internal(nil))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20+1024) // 5MB + header overhead

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	key, err := h.r2Svc.Upload(r.Context(), user.ID, header.Filename, file, header.Size)
	if err != nil {
		if strings.Contains(err.Error(), "file too large") || strings.Contains(err.Error(), "unsupported file type") {
			writeError(w, apperror.BadRequest(err.Error()))
			return
		}
		writeError(w, apperror.Internal(err))
		return
	}

	url := "/api/v1/attachments/" + key
	writeCreated(w, map[string]string{"url": url})
}

func (h *AttachmentHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if h.r2Svc == nil {
		http.NotFound(w, r)
		return
	}

	key := chi.URLParam(r, "*")
	if key == "" {
		http.NotFound(w, r)
		return
	}

	body, contentType, err := h.r2Svc.Get(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	io.Copy(w, body)
}
