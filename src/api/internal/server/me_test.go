package server_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
	"github.com/ettoreMB/personal_finance_control/api/internal/config"
	"github.com/ettoreMB/personal_finance_control/api/internal/db"
	"github.com/ettoreMB/personal_finance_control/api/internal/migrate"
	"github.com/ettoreMB/personal_finance_control/api/internal/server"
)

func loginCookie(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}) *http.Cookie {
	t.Helper()

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	resp = doLogin(t, app, "user@example.com", "abcdefg1!")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected login to succeed, got %d", resp.StatusCode)
	}

	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			return c
		}
	}

	t.Fatal("expected a session cookie to be set")
	return nil
}

func TestMeWithoutCookie(t *testing.T) {
	app := newTestApp(t)

	req, err := http.NewRequest(http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestMeWithInvalidCookie(t *testing.T) {
	app := newTestApp(t)

	req, err := http.NewRequest(http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: "not-a-real-token"})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestMeWithExpiredCookie(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	if err := migrate.Up(migrationsDir(t), dbPath); err != nil {
		t.Fatalf("unexpected error applying migrations: %v", err)
	}

	conn, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to db: %v", err)
	}

	app := server.New(conn, config.Config{SQLitePath: dbPath})

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected registration to succeed, got %d", resp.StatusCode)
	}

	var user auth.User
	if err := conn.Where("email = ?", "user@example.com").First(&user).Error; err != nil {
		t.Fatalf("unexpected error loading user: %v", err)
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	expiredSession := auth.Session{
		UserID:    user.ID,
		TokenHash: auth.HashSessionToken(token),
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if err := conn.Create(&expiredSession).Error; err != nil {
		t.Fatalf("unexpected error creating expired session: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestMeWithValidCookieReturnsUserAndExtendsSession(t *testing.T) {
	app := newTestApp(t)

	cookie := loginCookie(t, app)

	req, err := http.NewRequest(http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestPublicRoutesRemainAccessibleWithoutSession(t *testing.T) {
	app := newTestApp(t)

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/healthz"},
	}

	for _, p := range paths {
		req, err := http.NewRequest(p.method, p.path, nil)
		if err != nil {
			t.Fatalf("unexpected error building request: %v", err)
		}

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StatusCode == http.StatusUnauthorized {
			t.Fatalf("expected %s %s to be accessible without a session", p.method, p.path)
		}
	}
}
