package handler

import (
	"testing"
	"time"
)

func TestGenerateStateRoundTrip(t *testing.T) {
	secret := "test-secret"

	state := generateState(secret)
	if state == "" {
		t.Fatal("expected non-empty state")
	}
	if !verifyState(state, secret) {
		t.Fatal("expected generated state to verify")
	}
}

func TestVerifyStateRejectsExpiredToken(t *testing.T) {
	secret := "test-secret"
	state := makeStateToken(time.Now().Add(-time.Minute), []byte("nonce"), secret)

	if verifyState(state, secret) {
		t.Fatal("expected expired state to be rejected")
	}
}

func TestVerifyStateRejectsWrongSecret(t *testing.T) {
	state := generateState("secret-a")

	if verifyState(state, "secret-b") {
		t.Fatal("expected state signed with another secret to be rejected")
	}
}
