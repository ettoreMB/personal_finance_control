package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
)

func doRecovery(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, secret, newPassword string) *http.Response {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"recovery_secret": secret,
		"new_password":    newPassword,
	})
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/auth/recovery", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return resp
}

func withRecoverySecret(secret string) func(*config.Config) {
	return func(cfg *config.Config) {
		cfg.RecoverySecret = secret
	}
}

func TestRecoverySuccessResetsPassword(t *testing.T) {
	app := newTestApp(t, withRecoverySecret("s3cr3t"))

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRecovery(t, app, "s3cr3t", "newpassword1!")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	resp = doLogin(t, app, "user@example.com", "newpassword1!")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected login with new password to succeed, got %d", resp.StatusCode)
	}
}

func TestRecoveryIncorrectSecret(t *testing.T) {
	app := newTestApp(t, withRecoverySecret("s3cr3t"))

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRecovery(t, app, "wrong-secret", "newpassword1!")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	resp = doLogin(t, app, "user@example.com", "abcdefg1!")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected original password to still work, got %d", resp.StatusCode)
	}
}

func TestRecoveryWeakPassword(t *testing.T) {
	app := newTestApp(t, withRecoverySecret("s3cr3t"))

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRecovery(t, app, "s3cr3t", "weak")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRecoveryNoUsers(t *testing.T) {
	app := newTestApp(t, withRecoverySecret("s3cr3t"))

	resp := doRecovery(t, app, "s3cr3t", "newpassword1!")
	if resp.StatusCode < 400 || resp.StatusCode >= 500 {
		t.Fatalf("expected an explicit 4xx error, got %d", resp.StatusCode)
	}
}

func TestRecoveryMoreThanOneUser(t *testing.T) {
	app := newTestApp(t, withRecoverySecret("s3cr3t"))

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRegister(t, app, map[string]string{
		"email": "second@example.com",
		"cpf":   "111.444.777-35",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected second registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRecovery(t, app, "s3cr3t", "newpassword1!")
	if resp.StatusCode < 400 || resp.StatusCode >= 500 {
		t.Fatalf("expected an explicit 4xx error, got %d", resp.StatusCode)
	}
}
