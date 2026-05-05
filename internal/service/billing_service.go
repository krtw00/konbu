package service

import (
	"context"
	"errors"
	"strings"

	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/model"
	stripe "github.com/stripe/stripe-go/v85"
	portalsession "github.com/stripe/stripe-go/v85/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/customer"
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

func (s *BillingService) CreatePortalSession(ctx context.Context, user *model.User) (string, error) {
	if s.cfg.StripeSecretKey == "" || s.cfg.BaseURL == "" {
		return "", apperror.ServiceUnavailable("billing is not configured")
	}

	customerID, err := s.findCustomerIDByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}
	if customerID == "" {
		return "", apperror.NotFound("no Stripe customer found for this account")
	}

	baseURL := strings.TrimRight(s.cfg.BaseURL, "/")
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(baseURL + "/settings"),
	}
	params.Params.Context = ctx

	session, err := portalsession.New(params)
	if err != nil {
		return "", apperror.Internal(err)
	}
	if session.URL == "" {
		return "", apperror.Internal(errors.New("billing portal session URL was empty"))
	}
	return session.URL, nil
}

func (s *BillingService) findCustomerIDByEmail(ctx context.Context, email string) (string, error) {
	if email == "" {
		return "", apperror.BadRequest("user email is required")
	}
	params := &stripe.CustomerListParams{
		Email: stripe.String(email),
	}
	params.Limit = stripe.Int64(1)
	params.Context = ctx

	iter := customer.List(params)
	for iter.Next() {
		c := iter.Customer()
		if c != nil && c.ID != "" {
			return c.ID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return "", apperror.Internal(err)
	}
	return "", nil
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
