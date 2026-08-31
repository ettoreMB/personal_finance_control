package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func registerBody(t *testing.T, overrides map[string]string) *bytes.Buffer {
	t.Helper()

	payload := map[string]string{
		"email":            "user@example.com",
		"password":         "abcdefg1!",
		"confirm_password": "abcdefg1!",
		"cpf":              "529.982.247-25",
	}
	for k, v := range overrides {
		payload[k] = v
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	return bytes.NewBuffer(body)
}

func doRegister(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, overrides map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, "/register", registerBody(t, overrides))
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

func TestRegisterSuccess(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, nil)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, map[string]string{
		"password":         "abcdefgh",
		"confirm_password": "abcdefgh",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRegisterPasswordMismatch(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, map[string]string{
		"confirm_password": "different1!",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRegisterInvalidCPF(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, map[string]string{
		"cpf": "11111111111",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRegister(t, app, map[string]string{
		"cpf": "111.444.777-35",
	})

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func TestRegisterDuplicateCPF(t *testing.T) {
	app := newTestApp(t)

	resp := doRegister(t, app, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first registration to succeed, got %d", resp.StatusCode)
	}

	resp = doRegister(t, app, map[string]string{
		"email": "other@example.com",
	})

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}
