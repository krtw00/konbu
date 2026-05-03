package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/middleware"
	"github.com/krtw00/konbu/internal/service"
	stripe "github.com/stripe/stripe-go/v85"
)

type BillingHandler struct {
	billingSvc *service.BillingService
	authSvc    *service.AuthService
}

func NewBillingHandler(billingSvc *service.BillingService, authSvc *service.AuthService) *BillingHandler {
	return &BillingHandler{
		billingSvc: billingSvc,
		authSvc:    authSvc,
	}
}

func (h *BillingHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/checkout", h.createCheckoutSession)
	return r
}

func (h *BillingHandler) createCheckoutSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{"code": "unauthorized", "message": "not logged in"},
		})
		return
	}

	var req struct {
		Interval string `json:"interval"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	url, err := h.billingSvc.CreateCheckoutSession(r.Context(), user, req.Interval)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, map[string]string{"url": url})
}

func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	event, err := h.billingSvc.ConstructWebhookEvent(payload, r.Header.Get("Stripe-Signature"))
	if err != nil {
		log.Printf("stripe webhook verification failed: %v", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("stripe webhook: failed to decode checkout session: %v", err)
			break
		}
		if session.Mode == stripe.CheckoutSessionModeSubscription {
			h.applyPlanUpdate(r.Context(), session.ClientReferenceID, session.Metadata["user_email"], "sponsor", string(event.Type))
		}

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("stripe webhook: failed to decode subscription update: %v", err)
			break
		}
		switch string(sub.Status) {
		case "active", "trialing", "past_due":
			h.applyPlanUpdate(r.Context(), sub.Metadata["user_id"], sub.Metadata["user_email"], "sponsor", string(event.Type))
		case "canceled", "incomplete_expired", "unpaid", "paused":
			h.applyPlanUpdate(r.Context(), sub.Metadata["user_id"], sub.Metadata["user_email"], "free", string(event.Type))
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("stripe webhook: failed to decode subscription deletion: %v", err)
			break
		}
		h.applyPlanUpdate(r.Context(), sub.Metadata["user_id"], sub.Metadata["user_email"], "free", string(event.Type))
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) applyPlanUpdate(ctx context.Context, userIDRaw, email, plan, source string) {
	if userIDRaw != "" {
		if userID, err := uuid.Parse(userIDRaw); err == nil {
			if err := h.authSvc.UpdatePlan(ctx, userID, plan); err != nil {
				log.Printf("stripe webhook: failed to update plan by user id %s (%s): %v", userIDRaw, source, err)
			} else {
				log.Printf("stripe webhook: user %s -> %s (%s)", userIDRaw, plan, source)
			}
			return
		}
	}
	if email != "" {
		if err := h.authSvc.UpdatePlanByEmail(ctx, email, plan); err != nil {
			log.Printf("stripe webhook: failed to update plan by email %s (%s): %v", email, source, err)
		} else {
			log.Printf("stripe webhook: %s -> %s (%s)", email, plan, source)
		}
	}
}
