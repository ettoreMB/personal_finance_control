package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"testing"
)

func TestCategoriesRequiresAuthentication(t *testing.T) {
	app := newTestApp(t)

	req, err := http.NewRequest(http.MethodGet, "/categories", nil)
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

func TestCreateCategoryRequiresAuthentication(t *testing.T) {
	app := newTestApp(t)

	resp := doCreateCategory(t, app, &http.Cookie{Name: "session", Value: ""}, "nova")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestCategoriesListsSeededNames(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

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
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var got []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}

	names := make([]string, 0, len(got))
	for _, c := range got {
		if c.ID == 0 {
			t.Fatal("expected each category to have a non-zero id")
		}
		names = append(names, c.Name)
	}
	slices.Sort(names)

	want := []string{"carro", "casa", "comida", "lazer"}
	if !slices.Equal(names, want) {
		t.Fatalf("expected seeded categories %v, got %v", want, names)
	}
}

func doCreateCategory(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, name string) *http.Response {
	t.Helper()

	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
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

func TestCreateCategory(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "saúde")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var created struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created category to have a non-zero id")
	}
	if created.Name != "saúde" {
		t.Fatalf("expected name %q, got %q", "saúde", created.Name)
	}

	req, err := http.NewRequest(http.MethodGet, "/categories", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	listResp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var list []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatalf("unexpected error decoding list: %v", err)
	}

	names := make([]string, 0, len(list))
	for _, c := range list {
		names = append(names, c.Name)
	}
	if !slices.Contains(names, "saúde") {
		t.Fatalf("expected list to include %q, got %v", "saúde", names)
	}
}

func TestCreateCategoryEmptyName(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "   ")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCreateCategoryDuplicateName(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "casa")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func doPatchCategory(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint, name string) *http.Response {
	t.Helper()

	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/categories/%d", id), bytes.NewBuffer(body))
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

func TestRenameCategory(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "saude")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create to succeed, got %d", resp.StatusCode)
	}

	var created struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}

	resp = doPatchCategory(t, app, cookie, created.ID, "saúde")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var updated struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("expected id %d, got %d", created.ID, updated.ID)
	}
	if updated.Name != "saúde" {
		t.Fatalf("expected name %q, got %q", "saúde", updated.Name)
	}
}

func TestRenameCategoryUnknownID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doPatchCategory(t, app, cookie, 9999, "nova")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestRenameCategoryEmptyName(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "temp")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create to succeed, got %d", resp.StatusCode)
	}

	var created struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}

	resp = doPatchCategory(t, app, cookie, created.ID, "  ")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRenameCategoryDuplicateName(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "temp")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create to succeed, got %d", resp.StatusCode)
	}

	var created struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}

	resp = doPatchCategory(t, app, cookie, created.ID, "casa")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func doDeleteCategory(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/categories/%d", id), nil)
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

func TestDeleteUnusedCategory(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doCreateCategory(t, app, cookie, "temp")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create to succeed, got %d", resp.StatusCode)
	}

	var created struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding body: %v", err)
	}

	resp = doDeleteCategory(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, "/categories", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.AddCookie(cookie)

	listResp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var list []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatalf("unexpected error decoding list: %v", err)
	}

	names := make([]string, 0, len(list))
	for _, c := range list {
		names = append(names, c.Name)
	}
	if slices.Contains(names, "temp") {
		t.Fatalf("expected deleted category to be gone, got %v", names)
	}
}

func TestDeleteCategoryUnknownID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doDeleteCategory(t, app, cookie, 9999)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestDeleteCategoryInUseConflict(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	resp := doCreateEntry(t, app, cookie, entryPayload{
		Type:        "expense",
		AmountCents: 100,
		EntryDate:   "2026-01-01",
		CategoryID:  categoryID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected entry create, got %d", resp.StatusCode)
	}

	resp = doDeleteCategory(t, app, cookie, categoryID)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func TestDeleteCategoryAfterReclassify(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casaID := casaCategoryID(t, app, cookie)

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

	resp = doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 100, EntryDate: "2026-01-01", CategoryID: casaID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected entry create, got %d", resp.StatusCode)
	}
	var created entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding entry: %v", err)
	}

	resp = doPatchEntry(t, app, cookie, created.ID, map[string]any{"category_id": health.ID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected reclassify, got %d", resp.StatusCode)
	}

	resp = doDeleteCategory(t, app, cookie, casaID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestDeleteCategoryAfterSoftDeletingEntries(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 100, EntryDate: "2026-01-01", CategoryID: categoryID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected entry create, got %d", resp.StatusCode)
	}
	var created entryResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("unexpected error decoding entry: %v", err)
	}

	resp = doDeleteEntry(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected entry delete, got %d", resp.StatusCode)
	}

	resp = doDeleteCategory(t, app, cookie, categoryID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}
