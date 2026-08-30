package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func doLogin(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, email, password string) *http.Response {
	t.Helper()

	body, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
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

func TestLoginSuccessSetsCookieAndCreatesSession(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doLogin(t, app, "user@example.com", "abcdefg1!")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected a session cookie to be set")
	}
	if sessionCookie.Value == "" {
		t.Fatal("expected session cookie to have a value")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doLogin(t, app, "user@example.com", "wrongpassword1!")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	app := newTestApp(t)

	resp := doLogin(t, app, "nobody@example.com", "abcdefg1!")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestLogoutInvalidatesSessionAndClearsCookie(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doLogin(t, app, "user@example.com", "abcdefg1!")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected login to succeed, got %d", resp.StatusCode)
	}

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected a session cookie to be set")
	}

	req, err := http.NewRequest(http.MethodPost, "/logout", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(sessionCookie)

	logoutResp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, logoutResp.StatusCode)
	}

	var clearedCookie *http.Cookie
	for _, c := range logoutResp.Cookies() {
		if c.Name == "session" {
			clearedCookie = c
		}
	}
	if clearedCookie == nil {
		t.Fatal("expected logout response to clear the session cookie")
	}
	if clearedCookie.Value != "" {
		t.Fatalf("expected cleared cookie to have empty value, got %q", clearedCookie.Value)
	}
}
