package auth_test

import (
	"testing"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
)

func TestGenerateSessionTokenIsRandomAndOpaque(t *testing.T) {
	t1, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t2, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if t1 == "" || t2 == "" {
		t.Fatal("expected non-empty tokens")
	}

	if t1 == t2 {
		t.Fatal("expected different tokens across calls")
	}
}

func TestHashSessionTokenIsDeterministic(t *testing.T) {
	token := "some-raw-token"

	h1 := auth.HashSessionToken(token)
	h2 := auth.HashSessionToken(token)

	if h1 != h2 {
		t.Fatal("expected hashing the same token to be deterministic")
	}

	if h1 == token {
		t.Fatal("expected hash to differ from raw token")
	}
}
