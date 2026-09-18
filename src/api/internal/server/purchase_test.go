package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

type purchasePayload struct {
	Description      string `json:"description"`
	PurchaseDate     string `json:"purchase_date"`
	AmountCents      int    `json:"amount_cents"`
	InstallmentCount int    `json:"installment_count"`
	CategoryID       uint   `json:"category_id"`
}

type purchaseInstallment struct {
	ID                 uint   `json:"id"`
	InstallmentNumber  int    `json:"installment_number"`
	AmountCents        int    `json:"amount_cents"`
	EntryDate          string `json:"entry_date"`
	CategoryID         uint   `json:"category_id"`
}

type purchaseResponse struct {
	ID               uint                  `json:"id"`
	Description      string                `json:"description"`
	PurchaseDate     string                `json:"purchase_date"`
	AmountCents      int                   `json:"amount_cents"`
	InstallmentCount int                   `json:"installment_count"`
	CategoryID       uint                  `json:"category_id"`
	CategoryName     string                `json:"category_name"`
	Category         *struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Installments []purchaseInstallment `json:"installments"`
}

type listedEntry struct {
	ID                uint   `json:"id"`
	Type              string `json:"type"`
	AmountCents       int    `json:"amount_cents"`
	EntryDate         string `json:"entry_date"`
	CategoryID        uint   `json:"category_id"`
	PurchaseID        *uint  `json:"purchase_id"`
	InstallmentNumber *int   `json:"installment_number"`
	Purchase          *struct {
		ID               uint   `json:"id"`
		Description      string `json:"description"`
		InstallmentCount int    `json:"installment_count"`
	} `json:"purchase"`
}

func doCreatePurchase(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, payload purchasePayload) *http.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/purchases", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return resp
}

func doListPurchases(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/purchases", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return resp
}

func doGetPurchase(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/purchases/%d", id), nil)
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

func decodePurchase(t *testing.T, resp *http.Response) purchaseResponse {
	t.Helper()
	var got purchaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unexpected error decoding purchase: %v", err)
	}
	return got
}

func TestPurchasesRequiresAuthentication(t *testing.T) {
	app := newTestApp(t)

	resp := doListPurchases(t, app, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	resp = doCreatePurchase(t, app, nil, purchasePayload{
		Description: "TV", PurchaseDate: "2026-01-31", AmountCents: 15000, InstallmentCount: 3, CategoryID: 1,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected create %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestCreatePurchaseGeneratesInstallments(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	resp := doCreatePurchase(t, app, cookie, purchasePayload{
		Description:      "TV Samsung",
		PurchaseDate:     "2026-01-31",
		AmountCents:      10000,
		InstallmentCount: 3,
		CategoryID:       categoryID,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	created := decodePurchase(t, resp)
	if created.ID == 0 {
		t.Fatal("expected purchase id")
	}
	if created.Description != "TV Samsung" || created.AmountCents != 10000 || created.InstallmentCount != 3 {
		t.Fatalf("unexpected purchase: %+v", created)
	}
	if created.CategoryName != "casa" {
		t.Fatalf("expected category_name casa, got %q", created.CategoryName)
	}
	if len(created.Installments) != 3 {
		t.Fatalf("expected 3 installments, got %d", len(created.Installments))
	}

	wantDates := []string{"2026-01-31", "2026-02-28", "2026-03-31"}
	wantAmounts := []int{3333, 3333, 3334}
	for i, inst := range created.Installments {
		if inst.InstallmentNumber != i+1 {
			t.Fatalf("expected installment %d, got %d", i+1, inst.InstallmentNumber)
		}
		if inst.EntryDate != wantDates[i] {
			t.Fatalf("installment %d: expected date %s, got %s", i+1, wantDates[i], inst.EntryDate)
		}
		if inst.AmountCents != wantAmounts[i] {
			t.Fatalf("installment %d: expected %d cents, got %d", i+1, wantAmounts[i], inst.AmountCents)
		}
		if inst.CategoryID != categoryID {
			t.Fatalf("expected category %d, got %d", categoryID, inst.CategoryID)
		}
	}

	listResp := doListEntries(t, app, cookie)
	var entries []listedEntry
	if err := json.NewDecoder(listResp.Body).Decode(&entries); err != nil {
		t.Fatalf("unexpected error decoding entries: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	parcela := entries[0]
	if parcela.Type != "expense" {
		t.Fatalf("expected expense, got %s", parcela.Type)
	}
	if parcela.PurchaseID == nil || *parcela.PurchaseID != created.ID {
		t.Fatalf("expected purchase_id %d, got %+v", created.ID, parcela.PurchaseID)
	}
	if parcela.Purchase == nil || parcela.Purchase.Description != "TV Samsung" || parcela.Purchase.InstallmentCount != 3 {
		t.Fatalf("expected nested purchase, got %+v", parcela.Purchase)
	}
}

func TestCreatePurchaseValidation(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	valid := purchasePayload{
		Description: "TV", PurchaseDate: "2026-06-01", AmountCents: 1000, InstallmentCount: 2, CategoryID: categoryID,
	}

	cases := []struct {
		name   string
		mutate func(*purchasePayload)
	}{
		{"empty description", func(p *purchasePayload) { p.Description = "  " }},
		{"n=1", func(p *purchasePayload) { p.InstallmentCount = 1 }},
		{"n=25", func(p *purchasePayload) { p.InstallmentCount = 25 }},
		{"total less than n", func(p *purchasePayload) { p.AmountCents = 1; p.InstallmentCount = 3 }},
		{"missing date", func(p *purchasePayload) { p.PurchaseDate = "" }},
		{"invalid date", func(p *purchasePayload) { p.PurchaseDate = "31-01-2026" }},
		{"missing category", func(p *purchasePayload) { p.CategoryID = 0 }},
		{"unknown category", func(p *purchasePayload) { p.CategoryID = 9999 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := valid
			tc.mutate(&payload)
			resp := doCreatePurchase(t, app, cookie, payload)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
			}
		})
	}
}

func TestCreatePurchaseAllowsDuplicateDescriptions(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	payload := purchasePayload{
		Description: "TV Samsung", PurchaseDate: "2026-06-01", AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}
	first := doCreatePurchase(t, app, cookie, payload)
	second := doCreatePurchase(t, app, cookie, payload)
	if first.StatusCode != http.StatusCreated || second.StatusCode != http.StatusCreated {
		t.Fatalf("expected both creates 201, got %d and %d", first.StatusCode, second.StatusCode)
	}
}

func TestListPurchasesNewestFirstAndGetByID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	resp := doListPurchases(t, app, cookie)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	var empty []purchaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&empty); err != nil {
		t.Fatalf("unexpected error decoding empty list: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty list, got %d", len(empty))
	}

	older := doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "Old", PurchaseDate: "2025-01-01", AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	})
	newer := doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "New", PurchaseDate: "2026-06-01", AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	})
	if older.StatusCode != http.StatusCreated || newer.StatusCode != http.StatusCreated {
		t.Fatal("expected both creates to succeed")
	}

	resp = doListPurchases(t, app, cookie)
	var list []purchaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("unexpected error decoding list: %v", err)
	}
	if len(list) != 2 || list[0].Description != "New" || list[1].Description != "Old" {
		t.Fatalf("expected newest first, got %+v", list)
	}

	got := decodePurchase(t, doGetPurchase(t, app, cookie, decodePurchase(t, newer).ID))
	if got.Description != "New" || len(got.Installments) != 2 {
		t.Fatalf("unexpected get: %+v", got)
	}

	missing := doGetPurchase(t, app, cookie, 9999)
	if missing.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", missing.StatusCode)
	}
}

func TestParcelaCannotBePatchedOrDeletedAsAvulso(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: "2026-06-01", AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}))
	parcelaID := created.Installments[0].ID

	resp := doPatchEntry(t, app, cookie, parcelaID, map[string]any{"amount_cents": 1})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected patch 409, got %d", resp.StatusCode)
	}

	resp = doDeleteEntry(t, app, cookie, parcelaID)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected delete 409, got %d", resp.StatusCode)
	}
}

func TestPostEntryRejectsPurchaseID(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	body, err := json.Marshal(map[string]any{
		"type":         "expense",
		"amount_cents": 100,
		"entry_date":   "2026-06-01",
		"category_id":  categoryID,
		"purchase_id":  1,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/entries", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestAvulsoListOmitsPurchase(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	createCasaExpense(t, app, cookie)

	resp := doListEntries(t, app, cookie)
	var entries []listedEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].PurchaseID != nil || entries[0].Purchase != nil {
		t.Fatalf("avulso should omit purchase, got %+v", entries[0])
	}
}

func doPatchPurchase(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint, body any) *http.Response {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("unexpected error marshalling body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/purchases/%d", id), bytes.NewBuffer(payload))
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

func doDeletePurchase(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, id uint) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/purchases/%d", id), nil)
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

func firstOfMonth(offset int) string {
	today := ledger.TodayCivil()
	d := ledger.AddMonthsClamped(time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC), offset)
	return ledger.FormatEntryDate(d)
}

func TestPatchPurchaseDescriptionDoesNotTouchAmounts(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: "2026-01-31", AmountCents: 10000, InstallmentCount: 3, CategoryID: categoryID,
	}))

	resp := doPatchPurchase(t, app, cookie, created.ID, map[string]any{"description": "TV 55"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	updated := decodePurchase(t, resp)
	if updated.Description != "TV 55" {
		t.Fatalf("expected new description, got %q", updated.Description)
	}
	if updated.Installments[0].AmountCents != 3333 || updated.Installments[0].EntryDate != "2026-01-31" {
		t.Fatalf("expected amounts/dates unchanged, got %+v", updated.Installments[0])
	}
}

func TestPatchPurchaseReclassifiesAllActiveParcelas(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casaID := casaCategoryID(t, app, cookie)

	catResp := doCreateCategory(t, app, cookie, "lazer-extra")
	if catResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected category create, got %d", catResp.StatusCode)
	}
	var lazer struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(catResp.Body).Decode(&lazer); err != nil {
		t.Fatalf("decode category: %v", err)
	}

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(-2), AmountCents: 3000, InstallmentCount: 3, CategoryID: casaID,
	}))

	resp := doPatchPurchase(t, app, cookie, created.ID, map[string]any{"category_id": lazer.ID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	updated := decodePurchase(t, resp)
	if updated.CategoryID != lazer.ID || updated.CategoryName != "lazer-extra" {
		t.Fatalf("expected category snapshot lazer-extra, got %+v", updated)
	}
	for _, inst := range updated.Installments {
		if inst.CategoryID != lazer.ID {
			t.Fatalf("expected parcela %d to follow category, got %d", inst.InstallmentNumber, inst.CategoryID)
		}
	}
}

func TestPatchPurchaseRedistributesRemainingOnly(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(-1), AmountCents: 3000, InstallmentCount: 3, CategoryID: categoryID,
	}))
	pastAmount := created.Installments[0].AmountCents
	pastDate := created.Installments[0].EntryDate

	resp := doPatchPurchase(t, app, cookie, created.ID, map[string]any{"amount_cents": 5000})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	updated := decodePurchase(t, resp)
	if updated.AmountCents != 5000 {
		t.Fatalf("expected total 5000, got %d", updated.AmountCents)
	}
	if updated.Installments[0].AmountCents != pastAmount || updated.Installments[0].EntryDate != pastDate {
		t.Fatalf("past parcela changed: %+v", updated.Installments[0])
	}
	leftover := 5000 - pastAmount
	if updated.Installments[1].AmountCents+updated.Installments[2].AmountCents != leftover {
		t.Fatalf("remaining did not absorb leftover %d: %+v", leftover, updated.Installments)
	}
	if updated.Installments[1].AmountCents+updated.Installments[2].AmountCents != leftover {
		t.Fatal("sum mismatch")
	}
	if updated.Installments[2].AmountCents < updated.Installments[1].AmountCents {
		t.Fatalf("remainder should land on last parcela: %+v", updated.Installments)
	}
}

func TestPatchPurchaseAmountRejectedWhenAllPastOrTooSmall(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	allPast := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "Old", PurchaseDate: firstOfMonth(-12), AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}))
	resp := doPatchPurchase(t, app, cookie, allPast.ID, map[string]any{"amount_cents": 400})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 when all past, got %d", resp.StatusCode)
	}

	partial := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(-1), AmountCents: 3000, InstallmentCount: 3, CategoryID: categoryID,
	}))
	resp = doPatchPurchase(t, app, cookie, partial.ID, map[string]any{"amount_cents": partial.Installments[0].AmountCents + 1})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 when leftover too small, got %d", resp.StatusCode)
	}
}

func TestPatchPurchaseRejectsImmutableFields(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(0), AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}))

	resp := doPatchPurchase(t, app, cookie, created.ID, map[string]any{"installment_count": 4})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for N, got %d", resp.StatusCode)
	}
	resp = doPatchPurchase(t, app, cookie, created.ID, map[string]any{"purchase_date": "2020-01-01"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for date, got %d", resp.StatusCode)
	}

	resp = doPatchPurchase(t, app, cookie, 9999, map[string]any{"description": "x"})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeletePurchaseUndoCurrentMonth(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(0), AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}))

	resp := doDeletePurchase(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	resp = doGetPurchase(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected get 404 after undo, got %d", resp.StatusCode)
	}

	list := doListPurchases(t, app, cookie)
	var purchases []purchaseResponse
	if err := json.NewDecoder(list.Body).Decode(&purchases); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(purchases) != 0 {
		t.Fatalf("expected undone purchase omitted, got %d", len(purchases))
	}

	entriesResp := doListEntries(t, app, cookie)
	var entries []listedEntry
	if err := json.NewDecoder(entriesResp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode entries: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected parcelas gone, got %d", len(entries))
	}

	resp = doDeletePurchase(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected second delete 404, got %d", resp.StatusCode)
	}
}

func TestDeletePurchaseRejectedWhenPastExists(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(-1), AmountCents: 3000, InstallmentCount: 3, CategoryID: categoryID,
	}))
	resp := doDeletePurchase(t, app, cookie, created.ID)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestListPurchasesIncludesCompletedOmitsUndone(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	categoryID := casaCategoryID(t, app, cookie)

	completed := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "Done", PurchaseDate: firstOfMonth(-12), AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}))
	current := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "Now", PurchaseDate: firstOfMonth(0), AmountCents: 200, InstallmentCount: 2, CategoryID: categoryID,
	}))
	if doDeletePurchase(t, app, cookie, current.ID).StatusCode != http.StatusNoContent {
		t.Fatal("expected undo of current purchase")
	}

	list := doListPurchases(t, app, cookie)
	var purchases []purchaseResponse
	if err := json.NewDecoder(list.Body).Decode(&purchases); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(purchases) != 1 || purchases[0].ID != completed.ID {
		t.Fatalf("expected only completed purchase, got %+v", purchases)
	}
}

func TestCategoryDeleteAfterUndo(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	catResp := doCreateCategory(t, app, cookie, "temporaria")
	var cat struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(catResp.Body).Decode(&cat); err != nil {
		t.Fatalf("decode: %v", err)
	}

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "TV", PurchaseDate: firstOfMonth(0), AmountCents: 200, InstallmentCount: 2, CategoryID: cat.ID,
	}))
	if doDeleteCategory(t, app, cookie, cat.ID).StatusCode != http.StatusConflict {
		t.Fatal("expected 409 while parcelas active")
	}
	if doDeletePurchase(t, app, cookie, created.ID).StatusCode != http.StatusNoContent {
		t.Fatal("expected undo")
	}
	if doDeleteCategory(t, app, cookie, cat.ID).StatusCode != http.StatusNoContent {
		t.Fatal("expected category delete after undo")
	}
}


