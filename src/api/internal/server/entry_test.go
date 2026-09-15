package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

type entryPayload struct {
	Type        string `json:"type"`
	AmountCents int    `json:"amount_cents"`
	EntryDate   string `json:"entry_date"`
	CategoryID  uint   `json:"category_id"`
}

type entryResponse struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	AmountCents int    `json:"amount_cents"`
	EntryDate   string `json:"entry_date"`
	CategoryID  uint   `json:"category_id"`
	Category    *struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}

func casaCategoryID(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie) uint {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/categories", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected categories list, got %d", resp.StatusCode)
	}

	var categories []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		t.Fatalf("unexpected error decoding categories: %v", err)
	}
	for _, c := range categories {
		if c.Name == "casa" {
			return c.ID
		}
	}
	t.Fatal("expected seeded category casa")
	return 0
}

func doCreateEntry(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, payload entryPayload) *http.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/entries", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return resp
}

func doListEntries(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/entries", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return resp
}

func TestEntriesRequiresAuthentication(t *testing.T) {
	app := newTestApp(t)

	req, err := http.NewRequest(http.MethodGet, "/entries", nil)
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

func TestCreateExpenseAndIncome(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	resp := doCreateEntry(t, app, cookie, entryPayload{
		Type:        "expense",
		AmountCents: 4500,
		EntryDate:   "2026-01-15",
		CategoryID:  categoryID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var created entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created entry to have a non-zero id")
	}
	if created.Type != "expense" || created.AmountCents != 4500 || created.EntryDate != "2026-01-15" {
		t.Fatalf("unexpected created entry: %+v", created)
	}
	if created.CategoryID != categoryID {
		t.Fatalf("expected category_id %d, got %d", categoryID, created.CategoryID)
	}

	resp = doCreateEntry(t, app, cookie, entryPayload{
		Type:        "income",
		AmountCents: 10000,
		EntryDate:   "2026-12-01",
		CategoryID:  categoryID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected income create %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestCreateEntryPastAndFutureDates(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	for _, date := range []string{"2020-01-01", "2099-12-31"} {
		resp := doCreateEntry(t, app, cookie, entryPayload{
			Type:        "expense",
			AmountCents: 1,
			EntryDate:   date,
			CategoryID:  categoryID,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected date %s to be accepted, got %d", date, resp.StatusCode)
		}
	}
}

func TestCreateEntryValidation(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	valid := entryPayload{
		Type:        "expense",
		AmountCents: 100,
		EntryDate:   "2026-06-01",
		CategoryID:  categoryID,
	}

	cases := []struct {
		name    string
		mutate  func(*entryPayload)
		status  int
	}{
		{"zero amount", func(p *entryPayload) { p.AmountCents = 0 }, http.StatusBadRequest},
		{"negative amount", func(p *entryPayload) { p.AmountCents = -1 }, http.StatusBadRequest},
		{"invalid type", func(p *entryPayload) { p.Type = "transfer" }, http.StatusBadRequest},
		{"missing date", func(p *entryPayload) { p.EntryDate = "" }, http.StatusBadRequest},
		{"invalid date", func(p *entryPayload) { p.EntryDate = "15-01-2026" }, http.StatusBadRequest},
		{"missing category", func(p *entryPayload) { p.CategoryID = 0 }, http.StatusBadRequest},
		{"unknown category", func(p *entryPayload) { p.CategoryID = 9999 }, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := valid
			tc.mutate(&payload)
			resp := doCreateEntry(t, app, cookie, payload)
			if resp.StatusCode != tc.status {
				t.Fatalf("expected status %d, got %d", tc.status, resp.StatusCode)
			}
		})
	}
}

func TestListEntriesNewestFirst(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	resp := doListEntries(t, app, cookie)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	var empty []entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&empty); err != nil {
		t.Fatalf("unexpected error decoding empty list: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty list, got %d items", len(empty))
	}

	older := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 100, EntryDate: "2026-01-01", CategoryID: categoryID,
	})
	newer := doCreateEntry(t, app, cookie, entryPayload{
		Type: "income", AmountCents: 200, EntryDate: "2026-02-01", CategoryID: categoryID,
	})
	if older.StatusCode != http.StatusCreated || newer.StatusCode != http.StatusCreated {
		t.Fatal("expected both creates to succeed")
	}

	resp = doListEntries(t, app, cookie)
	var list []entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("unexpected error decoding list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	if list[0].EntryDate != "2026-02-01" || list[1].EntryDate != "2026-01-01" {
		t.Fatalf("expected newest first, got %s then %s", list[0].EntryDate, list[1].EntryDate)
	}
	if list[0].Category == nil || list[0].Category.Name != "casa" {
		t.Fatalf("expected nested category casa, got %+v", list[0].Category)
	}
}

func doPatchEntry(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint, body any) *http.Response {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/entries/%d", id), bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return resp
}

func doGetEntry(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/entries/%d", id), nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return resp
}

func doDeleteEntry(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/entries/%d", id), nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return resp
}

func createCasaExpense(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie) entryResponse {
	t.Helper()

	resp := doCreateEntry(t, app, cookie, entryPayload{
		Type:        "expense",
		AmountCents: 4500,
		EntryDate:   "2026-01-15",
		CategoryID:  casaCategoryID(t, app, cookie),
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create to succeed, got %d", resp.StatusCode)
	}

	var created entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}
	return created
}

func TestGetEntryByID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	created := createCasaExpense(t, app, cookie)

	resp := doGetEntry(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetEntryUnknownID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doGetEntry(t, app, cookie, 9999)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestPatchEntryFields(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	created := createCasaExpense(t, app, cookie)

	resp := doCreateCategory(t, app, cookie, "saúde")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected category create, got %d", resp.StatusCode)
	}
	var health struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("unexpected error decoding category: %v", err)
	}

	resp = doPatchEntry(t, app, cookie, created.ID, map[string]any{
		"amount_cents": 9900,
		"entry_date":   "2025-12-01",
		"type":         "income",
		"category_id":  health.ID,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var updated entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}
	if updated.AmountCents != 9900 || updated.EntryDate != "2025-12-01" || updated.Type != "income" {
		t.Fatalf("unexpected updated entry: %+v", updated)
	}
	if updated.CategoryID != health.ID {
		t.Fatalf("expected reclassified category %d, got %d", health.ID, updated.CategoryID)
	}
}

func TestPatchEntryUnknownID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doPatchEntry(t, app, cookie, 9999, map[string]any{"amount_cents": 1})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestPatchEntryInvalidAmount(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	created := createCasaExpense(t, app, cookie)

	resp := doPatchEntry(t, app, cookie, created.ID, map[string]any{"amount_cents": 0})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestSoftDeleteEntryHidesFromListAndGet(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	created := createCasaExpense(t, app, cookie)

	resp := doDeleteEntry(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	resp = doListEntries(t, app, cookie)
	var list []entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("unexpected error decoding list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected deleted entry to be omitted, got %d items", len(list))
	}

	resp = doGetEntry(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected get after delete %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	resp = doPatchEntry(t, app, cookie, created.ID, map[string]any{"amount_cents": 1})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected patch after delete %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	resp = doDeleteEntry(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected second delete %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
