package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/sessions", h.listSessions)
	r.Post("/sessions", h.createSession)
	r.Get("/sessions/{id}", h.getSession)
	r.Put("/sessions/{id}", h.updateSession)
	r.Delete("/sessions/{id}", h.deleteSession)
	r.Post("/sessions/{id}/messages", h.sendMessage)
	r.Get("/config", h.getConfig)
	r.Put("/config", h.saveConfig)

	return r
}

func (h *ChatHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	sessions, err := h.chatSvc.ListSessions(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, sessions)
}

func (h *ChatHandler) createSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	sess, err := h.chatSvc.CreateSession(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, sess)
}

func (h *ChatHandler) getSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	detail, err := h.chatSvc.GetSession(r.Context(), id, user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, detail)
}

func (h *ChatHandler) updateSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req model.UpdateSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.chatSvc.UpdateSessionTitle(r.Context(), id, user.ID, req.Title); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *ChatHandler) deleteSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.chatSvc.DeleteSession(r.Context(), id, user.ID); err != nil {
		writeError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *ChatHandler) sendMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	var req model.SendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	stream, err := h.chatSvc.SendMessage(r.Context(), user.ID, sessionID, req.Content, user)
	if err != nil {
		writeError(w, err)
		return
	}

	// SSE response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for ev := range stream {
		data, _ := json.Marshal(ev.Data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, string(data))
		flusher.Flush()
	}
}

func (h *ChatHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	cfg, err := h.chatSvc.GetAIConfig(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, cfg)
}

func (h *ChatHandler) saveConfig(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	var req model.AIChatConfig
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := h.chatSvc.SaveAIConfig(r.Context(), user.ID, req); err != nil {
		writeError(w, err)
		return
	}
	cfg, err := h.chatSvc.GetAIConfig(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, cfg)
}
