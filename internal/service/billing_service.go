package service

import (
	"context"
	"errors"
	"strings"

	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/model"
	stripe "github.com/stripe/stripe-go/v85"
	checkoutsession "github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/webhook"
)

type BillingService struct {
	cfg *config.Config
}

func NewBillingService(cfg *config.Config) *BillingService {
	if cfg.StripeSecretKey != "" {
		stripe.Key = cfg.StripeSecretKey
	}
	return &BillingService{cfg: cfg}
}

func (s *BillingService) CreateCheckoutSession(ctx context.Context, user *model.User, interval string) (string, error) {
	if s.cfg.StripeSecretKey == "" || s.cfg.BaseURL == "" {
		return "", apperror.ServiceUnavailable("billing is not configured")
	}

	var priceID string
	switch interval {
	case "month":
		priceID = s.cfg.StripePriceMonthly
	case "year":
		priceID = s.cfg.StripePriceYearly
	default:
		return "", apperror.BadRequest("invalid billing interval")
	}
	if priceID == "" {
		return "", apperror.ServiceUnavailable("selected billing plan is not configured")
	}

	baseURL := strings.TrimRight(s.cfg.BaseURL, "/")
	params := &stripe.CheckoutSessionParams{
		Mode:                stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:          stripe.String(baseURL + "/settings?billing=success"),
		CancelURL:           stripe.String(baseURL + "/settings?billing=cancel"),
		CustomerEmail:       stripe.String(user.Email),
		ClientReferenceID:   stripe.String(user.ID.String()),
		AllowPromotionCodes: stripe.Bool(true),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{},
	}
	params.AddMetadata("user_id", user.ID.String())
	params.AddMetadata("user_email", user.Email)
	params.AddMetadata("plan", "sponsor")
	params.SubscriptionData.AddMetadata("user_id", user.ID.String())
	params.SubscriptionData.AddMetadata("user_email", user.Email)
	params.SubscriptionData.AddMetadata("plan", "sponsor")
	params.Params.Context = ctx

	session, err := checkoutsession.New(params)
	if err != nil {
		return "", apperror.Internal(err)
	}
	if session.URL == "" {
		return "", apperror.Internal(errors.New("checkout session URL was empty"))
	}
	return session.URL, nil
}

func (s *BillingService) ConstructWebhookEvent(payload []byte, signature string) (*stripe.Event, error) {
	if s.cfg.StripeWebhookSecret == "" {
		return nil, apperror.ServiceUnavailable("stripe webhook is not configured")
	}
	event, err := webhook.ConstructEvent(payload, signature, s.cfg.StripeWebhookSecret)
	if err != nil {
		return nil, apperror.Unauthorized("invalid stripe webhook signature")
	}
	return &event, nil
}
