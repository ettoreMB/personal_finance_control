package ledger_test

import (
	"testing"
	"time"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

func TestCivilDateInSaoPauloKeepsSeptemberAt21UTCOnThe30th(t *testing.T) {
	instant := time.Date(2026, time.September, 30, 21, 0, 0, 0, time.UTC)
	got := ledger.CivilDateIn(instant, ledger.SaoPaulo())
	want := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339), got.Format(time.RFC3339))
	}
}

func TestCivilDateInSaoPauloTurnsOctoberUTCIntoOctober(t *testing.T) {
	instant := time.Date(2026, time.October, 1, 3, 0, 0, 0, time.UTC)
	got := ledger.CivilDateIn(instant, ledger.SaoPaulo())
	want := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339), got.Format(time.RFC3339))
	}
}

func TestAddMonthsClampedJanuary31ToFebruary(t *testing.T) {
	anchor := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	got := ledger.AddMonthsClamped(anchor, 1)
	want := time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format("2006-01-02"), got.Format("2006-01-02"))
	}

	march := ledger.AddMonthsClamped(anchor, 2)
	wantMarch := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)
	if !march.Equal(wantMarch) {
		t.Fatalf("expected %s, got %s", wantMarch.Format("2006-01-02"), march.Format("2006-01-02"))
	}
}

func TestSplitCentsPutsRemainderOnLast(t *testing.T) {
	got := ledger.SplitCents(10000, 3)
	want := []int{3333, 3333, 3334}
	if len(got) != len(want) {
		t.Fatalf("expected %d parts, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("part %d: expected %d, got %d", i, want[i], got[i])
		}
	}
}

func TestIsPastMonth(t *testing.T) {
	today := time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)
	if !ledger.IsPastMonth(time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC), today) {
		t.Fatal("expected August to be past in September")
	}
	if ledger.IsPastMonth(time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), today) {
		t.Fatal("expected current month not to be past")
	}
	if ledger.IsPastMonth(time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), today) {
		t.Fatal("expected next month not to be past")
	}
}
