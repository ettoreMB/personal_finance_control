package server_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

type summaryPeriod struct {
	Kind  string `json:"kind"`
	Year  int    `json:"year"`
	Month int    `json:"month"`
}

type summaryCategory struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	IncomeCents  int    `json:"income_cents"`
	ExpenseCents int    `json:"expense_cents"`
	BalanceCents int    `json:"balance_cents"`
}

type summaryResponse struct {
	Period       summaryPeriod     `json:"period"`
	IncomeCents  int               `json:"income_cents"`
	ExpenseCents int               `json:"expense_cents"`
	BalanceCents int               `json:"balance_cents"`
	Categories   []summaryCategory `json:"categories"`
}

func doGetSummary(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "/summary", nil)
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

func decodeSummary(t *testing.T, resp *http.Response) summaryResponse {
	t.Helper()
	var got summaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("unexpected error decoding summary: %v", err)
	}
	return got
}

func TestSummaryRequiresAuthentication(t *testing.T) {
	app := newTestApp(t)

	resp := doGetSummary(t, app, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestSummaryEmptyMonthReturnsZeros(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)

	resp := doGetSummary(t, app, cookie)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	got := decodeSummary(t, resp)
	today := ledger.TodayCivil()
	y, m, _ := today.Date()
	if got.Period.Kind != "month" || got.Period.Year != y || got.Period.Month != int(m) {
		t.Fatalf("unexpected period: %+v (today %s)", got.Period, today.Format("2006-01-02"))
	}
	if got.IncomeCents != 0 || got.ExpenseCents != 0 || got.BalanceCents != 0 {
		t.Fatalf("expected zeros, got income=%d expense=%d balance=%d", got.IncomeCents, got.ExpenseCents, got.BalanceCents)
	}
	if got.Categories == nil {
		t.Fatal("expected categories array, got null")
	}
	if len(got.Categories) != 0 {
		t.Fatalf("expected empty categories, got %+v", got.Categories)
	}
}

func lastOfCurrentMonth() string {
	today := ledger.TodayCivil()
	d := time.Date(today.Year(), today.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	return ledger.FormatEntryDate(d)
}

func categoryIDByName(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, cookie *http.Cookie, name string) uint {
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
	var categories []struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		t.Fatalf("unexpected error decoding categories: %v", err)
	}
	for _, c := range categories {
		if c.Name == name {
			return c.ID
		}
	}
	t.Fatalf("expected seeded category %q", name)
	return 0
}

func TestSummaryCountsAvulsosInCurrentMonth(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casa := casaCategoryID(t, app, cookie)
	comida := categoryIDByName(t, app, cookie, "comida")

	inMonth := firstOfMonth(0)
	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "income", AmountCents: 10000, EntryDate: inMonth, CategoryID: casa,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create income: %d", resp.StatusCode)
	}
	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 3000, EntryDate: lastOfCurrentMonth(), CategoryID: comida,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create expense in month: %d", resp.StatusCode)
	}
	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 4000, EntryDate: firstOfMonth(-1), CategoryID: casa,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create previous month: %d", resp.StatusCode)
	}
	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 5000, EntryDate: firstOfMonth(1), CategoryID: comida,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create next month: %d", resp.StatusCode)
	}

	got := decodeSummary(t, doGetSummary(t, app, cookie))
	if got.IncomeCents != 10000 || got.ExpenseCents != 3000 || got.BalanceCents != 7000 {
		t.Fatalf("unexpected totals: %+v", got)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %+v", got.Categories)
	}
	if got.Categories[0].CategoryName != "comida" || got.Categories[0].ExpenseCents != 3000 {
		t.Fatalf("expected comida first by expense, got %+v", got.Categories)
	}
	if got.Categories[1].CategoryName != "casa" || got.Categories[1].IncomeCents != 10000 || got.Categories[1].ExpenseCents != 0 {
		t.Fatalf("expected casa income-only, got %+v", got.Categories)
	}
}

func TestSummaryOmitsDeletedAvulsoAndMergesCategory(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casa := casaCategoryID(t, app, cookie)
	inMonth := firstOfMonth(0)

	keep := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 2000, EntryDate: inMonth, CategoryID: casa,
	})
	if keep.StatusCode != http.StatusCreated {
		t.Fatalf("create keep: %d", keep.StatusCode)
	}
	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 1000, EntryDate: inMonth, CategoryID: casa,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create merge: %d", resp.StatusCode)
	}

	drop := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 9000, EntryDate: inMonth, CategoryID: casa,
	})
	if drop.StatusCode != http.StatusCreated {
		t.Fatalf("create drop: %d", drop.StatusCode)
	}
	var dropped entryResponse
	if err := json.NewDecoder(drop.Body).Decode(&dropped); err != nil {
		t.Fatalf("decode drop: %v", err)
	}
	if resp := doDeleteEntry(t, app, cookie, dropped.ID); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", resp.StatusCode)
	}

	got := decodeSummary(t, doGetSummary(t, app, cookie))
	if got.ExpenseCents != 3000 || got.BalanceCents != -3000 {
		t.Fatalf("expected 3000 expense after delete, got %+v", got)
	}
	if len(got.Categories) != 1 || got.Categories[0].CategoryName != "casa" || got.Categories[0].ExpenseCents != 3000 {
		t.Fatalf("expected merged casa, got %+v", got.Categories)
	}
}

func TestSummarySortsCategoriesByExpenseThenName(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casa := casaCategoryID(t, app, cookie)
	carro := categoryIDByName(t, app, cookie, "carro")
	inMonth := firstOfMonth(0)

	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 1000, EntryDate: inMonth, CategoryID: casa,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("casa: %d", resp.StatusCode)
	}
	if resp := doCreateEntry(t, app, cookie, entryPayload{
		Type: "expense", AmountCents: 1000, EntryDate: inMonth, CategoryID: carro,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("carro: %d", resp.StatusCode)
	}

	got := decodeSummary(t, doGetSummary(t, app, cookie))
	if len(got.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %+v", got.Categories)
	}
	if got.Categories[0].CategoryName != "carro" || got.Categories[1].CategoryName != "casa" {
		t.Fatalf("expected carro then casa by name, got %+v", got.Categories)
	}
}

func TestSummaryCountsParcelaInCurrentMonthOnly(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casa := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description:      "TV",
		PurchaseDate:     firstOfMonth(0),
		AmountCents:      2000,
		InstallmentCount: 2,
		CategoryID:       casa,
	}))
	if created.ID == 0 {
		t.Fatal("expected purchase")
	}

	got := decodeSummary(t, doGetSummary(t, app, cookie))
	if got.ExpenseCents != 1000 || got.BalanceCents != -1000 {
		t.Fatalf("expected only current parcela 1000, got %+v", got)
	}
	if len(got.Categories) != 1 || got.Categories[0].ExpenseCents != 1000 {
		t.Fatalf("expected casa 1000, got %+v", got.Categories)
	}
}

func TestSummaryMergesParcelasFromDifferentPurchases(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casa := casaCategoryID(t, app, cookie)

	if resp := doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "A", PurchaseDate: firstOfMonth(0), AmountCents: 2000, InstallmentCount: 2, CategoryID: casa,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("purchase A: %d", resp.StatusCode)
	}
	if resp := doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "B", PurchaseDate: firstOfMonth(0), AmountCents: 4000, InstallmentCount: 2, CategoryID: casa,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("purchase B: %d", resp.StatusCode)
	}

	got := decodeSummary(t, doGetSummary(t, app, cookie))
	if got.ExpenseCents != 3000 {
		t.Fatalf("expected 1000+2000 from two first parcelas, got %+v", got)
	}
}

func TestSummaryOmitsUndonePurchase(t *testing.T) {
	app := newTestApp(t)
	cookie := loginCookie(t, app)
	casa := casaCategoryID(t, app, cookie)

	created := decodePurchase(t, doCreatePurchase(t, app, cookie, purchasePayload{
		Description: "erro", PurchaseDate: firstOfMonth(0), AmountCents: 2000, InstallmentCount: 2, CategoryID: casa,
	}))
	if resp := doDeletePurchase(t, app, cookie, created.ID); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("undo: %d", resp.StatusCode)
	}

	got := decodeSummary(t, doGetSummary(t, app, cookie))
	if got.ExpenseCents != 0 || len(got.Categories) != 0 {
		t.Fatalf("expected undone compra omitted, got %+v", got)
	}
}
