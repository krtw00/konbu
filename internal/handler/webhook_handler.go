package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/krtw00/konbu/internal/service"
)

type WebhookHandler struct {
	authSvc       *service.AuthService
	webhookSecret string
}

func NewWebhookHandler(authSvc *service.AuthService, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{authSvc: authSvc, webhookSecret: webhookSecret}
}

func (h *WebhookHandler) HandleGitHubSponsors(w http.ResponseWriter, r *http.Request) {
	if h.webhookSecret == "" {
		http.Error(w, "webhook not configured", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if !h.verifySignature(body, sig) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event struct {
		Action      string `json:"action"`
		Sponsorship struct {
			Sponsor struct {
				Login string `json:"login"`
			} `json:"sponsor"`
			SponsorEntity struct {
				Email string `json:"email"`
			} `json:"sponsor_entity"`
		} `json:"sponsorship"`
		Sender struct {
			Login string `json:"login"`
		} `json:"sender"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	email := event.Sponsorship.SponsorEntity.Email
	if email == "" {
		log.Printf("sponsors webhook: no email for sponsor %s, action=%s", event.Sponsorship.Sponsor.Login, event.Action)
		w.WriteHeader(http.StatusOK)
		return
	}

	var plan string
	switch event.Action {
	case "created":
		plan = "sponsor"
	case "cancelled", "tier_changed":
		if event.Action == "cancelled" {
			plan = "free"
		} else {
			plan = "sponsor"
		}
	default:
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.authSvc.UpdatePlanByEmail(r.Context(), email, plan); err != nil {
		log.Printf("sponsors webhook: failed to update plan for %s: %v", email, err)
	} else {
		log.Printf("sponsors webhook: %s -> %s (action=%s)", email, plan, event.Action)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) verifySignature(body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	sig, err := hex.DecodeString(signature[7:])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	return hmac.Equal(sig, mac.Sum(nil))
}
