package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
	"github.com/ettoreMB/personal_finance_control/api/internal/db"
	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
	"github.com/ettoreMB/personal_finance_control/api/internal/migrate"
	"github.com/ettoreMB/personal_finance_control/api/internal/server"
	"github.com/ettoreMB/personal_finance_control/api/internal/whatsapp"
)

const (
	ownerPhone    = "5511999999999"
	instanceToken = "instance-secret"
)

func TestWhatsAppRejectsBadToken(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")

	status, _ := postWhatsApp(t, app, webhookJSON(t, "m1", ownerJID(), "gastei 45 reais", "", false, "wrong"))
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
	status, _ = postWhatsApp(t, app, webhookJSON(t, "m2", ownerJID(), "gastei 45 reais", "", false, ""))
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", status)
	}
	if send.count() != 0 || class.count() != 0 {
		t.Fatalf("bad token classified %d and sent %d", class.count(), send.count())
	}
}

func TestWhatsAppIgnoresOtherChats(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")

	cases := []struct {
		name  string
		chat  string
		group bool
	}{
		{name: "other person", chat: "5511888888888@s.whatsapp.net"},
		{name: "group", chat: "120363@g.us", group: true},
		{name: "status", chat: "status@broadcast"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := postWhatsApp(t, app, webhookJSON(t, "id-"+tc.name, tc.chat, "gastei 45 reais", "", tc.group, instanceToken))
			if status != http.StatusOK {
				t.Fatalf("expected 200, got %d", status)
			}
		})
	}
	if send.count() != 0 || class.count() != 0 {
		t.Fatalf("ignored chats classified %d and sent %d", class.count(), send.count())
	}
}

func TestWhatsAppIgnoresDuplicateAndEmpty(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")
	body := webhookJSON(t, "same-id", ownerJID(), "gastei 45 reais no mercado", "", false, instanceToken)

	if status, _ := postWhatsApp(t, app, body); status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if status, _ := postWhatsApp(t, app, body); status != http.StatusOK {
		t.Fatalf("expected 200 on retry, got %d", status)
	}
	if class.count() != 1 || send.count() != 1 {
		t.Fatalf("retry classified %d and sent %d", class.count(), send.count())
	}

	before := send.count()
	status, _ := postWhatsApp(t, app, webhookJSON(t, "blank", ownerJID(), "   ", "", false, instanceToken))
	if status != http.StatusOK {
		t.Fatalf("expected 200 for empty text, got %d", status)
	}
	if send.count() != before {
		t.Fatalf("empty text sent a reply: %q", send.last())
	}
}

func TestWhatsAppDraftFromText(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")
	class.classify = func(text string, _ []string) (whatsapp.Classification, error) {
		kind := "expense"
		if strings.Contains(strings.ToLower(text), "recebi") {
			kind = "income"
		}
		return whatsapp.Classification{Type: kind, Category: "comida"}, nil
	}

	cases := []struct {
		text string
		want []string
	}{
		{text: "gastei 45 reais no mercado", want: []string{"Rascunho de gasto", "R$ 45,00", "comida", "01/10/2026", "Descrição: gastei 45 reais no mercado"}},
		{text: "almoço 45,90", want: []string{"R$ 45,90"}},
		{text: "paguei R$ 1.234,50 no mercado", want: []string{"R$ 1.234,50"}},
		{text: "gastei 45 reais ontem", want: []string{"30/09/2026"}},
		{text: "gastei 45 reais anteontem", want: []string{"29/09/2026"}},
		{text: "gastei 20 reais em 15/08", want: []string{"15/08/2026"}},
		{text: "gastei 20 reais em 05/09/2026", want: []string{"05/09/2026"}},
		{text: "recebi 200 do João", want: []string{"Rascunho de ganho", "R$ 200,00"}},
	}
	for i, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			body := webhookJSON(t, "phrase-"+string(rune('a'+i)), "5511999999999:38@s.whatsapp.net", tc.text, "", false, instanceToken)
			status, _ := postWhatsApp(t, app, body)
			if status != http.StatusOK {
				t.Fatalf("expected 200, got %d", status)
			}
			reply := send.last()
			for _, want := range tc.want {
				if !strings.Contains(reply, want) {
					t.Fatalf("reply %q does not contain %q", reply, want)
				}
			}
		})
	}

	extended := webhookJSON(t, "extended", ownerJID(), "", "gastei 15 reais no mercado", false, instanceToken)
	var payload map[string]any
	if err := json.Unmarshal(extended, &payload); err != nil {
		t.Fatal(err)
	}
	status, _ := postWhatsApp(t, app, extended)
	if status != http.StatusOK || !strings.Contains(send.last(), "R$ 15,00") {
		t.Fatalf("extended text reply: %q", send.last())
	}
}

func TestWhatsAppDraftHoles(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")

	class.classify = func(string, []string) (whatsapp.Classification, error) {
		return whatsapp.Classification{Type: "expense", Category: "comida"}, nil
	}
	postWhatsApp(t, app, webhookJSON(t, "no-amount", ownerJID(), "fui no mercado", "", false, instanceToken))
	if !strings.Contains(send.last(), "falta o valor") {
		t.Fatalf("reply: %q", send.last())
	}

	class.classify = func(string, []string) (whatsapp.Classification, error) {
		return whatsapp.Classification{Type: "expense", Category: ""}, nil
	}
	postWhatsApp(t, app, webhookJSON(t, "no-category", ownerJID(), "gastei 10 reais", "", false, instanceToken))
	if !strings.Contains(send.last(), "falta a categoria") || !strings.Contains(send.last(), "Substituí o rascunho anterior.") {
		t.Fatalf("reply: %q", send.last())
	}

	class.classify = func(string, []string) (whatsapp.Classification, error) {
		return whatsapp.Classification{Type: "", Category: "comida"}, nil
	}
	postWhatsApp(t, app, webhookJSON(t, "no-type", ownerJID(), "10 reais de alguma coisa", "", false, instanceToken))
	if !strings.Contains(send.last(), "falta se é ganho ou gasto") {
		t.Fatalf("reply: %q", send.last())
	}

	class.classify = func(string, []string) (whatsapp.Classification, error) {
		return whatsapp.Classification{Type: "expense", Category: "comida"}, nil
	}
	postWhatsApp(t, app, webhookJSON(t, "bad-date", ownerJID(), "gastei 10 reais em 31/02/2026", "", false, instanceToken))
	if !strings.Contains(send.last(), "Data: falta a data") {
		t.Fatalf("reply guessed a bad date: %q", send.last())
	}
}

func TestWhatsAppAudioAndEmptyDoNotReplaceDraft(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")
	postWhatsApp(t, app, webhookJSON(t, "base", ownerJID(), "gastei 45 reais no mercado", "", false, instanceToken))
	class.calls = 0

	status, _ := postWhatsApp(t, app, webhookJSON(t, "audio-1", ownerJID(), "", "audio", false, instanceToken))
	if status != http.StatusOK || send.last() != whatsapp.AudioNotAcceptedReply {
		t.Fatalf("audio reply: %q", send.last())
	}
	status, _ = postWhatsApp(t, app, webhookJSON(t, "empty-1", ownerJID(), "  ", "", false, instanceToken))
	if status != http.StatusOK || send.last() != whatsapp.AudioNotAcceptedReply {
		t.Fatalf("empty text changed the reply: %q", send.last())
	}
	if class.count() != 0 {
		t.Fatalf("audio or empty classified %d times", class.count())
	}

	postWhatsApp(t, app, webhookJSON(t, "confirm-kept", ownerJID(), "sim", "", false, instanceToken))
	if !strings.Contains(send.last(), "Lancei gasto de R$ 45,00") {
		t.Fatalf("draft was replaced before confirm: %q", send.last())
	}
}

func TestWhatsAppIgnoresOwnOutbound(t *testing.T) {
	app, _, send, class := newWhatsAppApp(t, "")
	send.ids = []string{"bot-1"}
	postWhatsApp(t, app, webhookJSON(t, "audio-own", ownerJID(), "", "audio", false, instanceToken))
	sent := send.count()
	classified := class.count()

	status, _ := postWhatsApp(t, app, webhookJSON(t, "bot-1", ownerJID(), "gastei 80 reais", "", false, instanceToken))
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if send.count() != sent || class.count() != classified {
		t.Fatalf("own outbound id was processed, sent %d classified %d", send.count(), class.count())
	}

	send.ids = []string{""}
	postWhatsApp(t, app, webhookJSON(t, "audio-no-id", ownerJID(), "", "audio", false, instanceToken))
	echo := send.last()
	sent = send.count()
	classified = class.count()
	status, _ = postWhatsApp(t, app, webhookJSON(t, "echo-text", ownerJID(), echo, "", false, instanceToken))
	if status != http.StatusOK || send.count() != sent || class.count() != classified {
		t.Fatalf("echo text was processed, reply %q", send.last())
	}
}

func TestWhatsAppDraftSurvivesNewProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.sqlite")
	app, _, _, _ := newWhatsAppApp(t, path)
	postWhatsApp(t, app, webhookJSON(t, "persist", ownerJID(), "gastei 45 reais no mercado", "", false, instanceToken))

	app2, _, send, _ := newWhatsAppApp(t, path)
	postWhatsApp(t, app2, webhookJSON(t, "persist-2", ownerJID(), "gastei 12 reais no mercado", "", false, instanceToken))
	if !strings.Contains(send.last(), "Substituí o rascunho anterior.") || !strings.Contains(send.last(), "R$ 12,00") {
		t.Fatalf("draft did not survive restart: %q", send.last())
	}
}

func TestWhatsAppConfirmAndDiscard(t *testing.T) {
	app, path, send, class := newWhatsAppApp(t, "")
	postWhatsApp(t, app, webhookJSON(t, "draft", ownerJID(), "  gastei 45 reais no mercado  ", "", false, instanceToken))

	status, _ := postWhatsApp(t, app, webhookJSON(t, "yes", ownerJID(), "  Sim  ", "", false, instanceToken))
	if status != http.StatusOK || !strings.Contains(send.last(), "Lancei gasto de R$ 45,00 em comida, em 01/10/2026.") {
		t.Fatalf("confirm reply: %q", send.last())
	}
	entry := onlyEntry(t, path)
	if entry.Type != "expense" || entry.AmountCents != 4500 || entry.Description != "gastei 45 reais no mercado" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
	if ledger.FormatEntryDate(entry.EntryDate) != "2026-10-01" || entry.PurchaseID != nil {
		t.Fatalf("unexpected entry date or purchase: %+v", entry)
	}
	if entry.Category.Name != "comida" {
		t.Fatalf("expected comida, got %+v", entry.Category)
	}

	cookie := loginCookie(t, app)
	resp := doPatchEntry(t, app, cookie, entry.ID, map[string]any{"description": "mercado"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected patch 200, got %d", resp.StatusCode)
	}
	resp = doDeleteEntry(t, app, cookie, entry.ID)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected delete 204, got %d", resp.StatusCode)
	}
	if n := entryCount(t, path); n != 0 {
		t.Fatalf("expected no active entries, got %d", n)
	}

	classified := class.count()
	postWhatsApp(t, app, webhookJSON(t, "yes-again", ownerJID(), "sim", "", false, instanceToken))
	if send.last() != whatsapp.NothingToConfirmReply || class.count() != classified {
		t.Fatalf("second sim reply %q classified %d", send.last(), class.count())
	}
	postWhatsApp(t, app, webhookJSON(t, "no-empty", ownerJID(), "não", "", false, instanceToken))
	if send.last() != whatsapp.NothingToDiscardReply {
		t.Fatalf("discard without draft: %q", send.last())
	}

	postWhatsApp(t, app, webhookJSON(t, "draft-2", ownerJID(), "recebi 200 do João", "", false, instanceToken))
	postWhatsApp(t, app, webhookJSON(t, "no", ownerJID(), "Nao", "", false, instanceToken))
	if send.last() != whatsapp.DiscardedReply {
		t.Fatalf("discard reply: %q", send.last())
	}
	if n := entryCount(t, path); n != 0 {
		t.Fatalf("discard created %d entries", n)
	}

	class.classify = func(string, []string) (whatsapp.Classification, error) {
		return whatsapp.Classification{Type: "expense", Category: "comida"}, nil
	}
	postWhatsApp(t, app, webhookJSON(t, "hole", ownerJID(), "fui no mercado", "", false, instanceToken))
	postWhatsApp(t, app, webhookJSON(t, "yes-hole", ownerJID(), "sim", "", false, instanceToken))
	if !strings.Contains(send.last(), "Ainda não lancei.") || !strings.Contains(send.last(), "Falta o valor.") {
		t.Fatalf("incomplete confirm: %q", send.last())
	}
	if n := entryCount(t, path); n != 0 {
		t.Fatalf("incomplete sim created %d entries", n)
	}
	postWhatsApp(t, app, webhookJSON(t, "full-then-hole", ownerJID(), "gastei 45 reais no mercado", "", false, instanceToken))
	postWhatsApp(t, app, webhookJSON(t, "cleared", ownerJID(), "fui no mercado de novo", "", false, instanceToken))
	postWhatsApp(t, app, webhookJSON(t, "yes-cleared", ownerJID(), "sim", "", false, instanceToken))
	if !strings.Contains(send.last(), "Falta o valor.") {
		t.Fatalf("old amount survived a new message: %q", send.last())
	}
	if n := entryCount(t, path); n != 0 {
		t.Fatalf("cleared draft created %d entries", n)
	}

	postWhatsApp(t, app, webhookJSON(t, "fix", ownerJID(), "gastei 30 reais no mercado", "", false, instanceToken))
	postWhatsApp(t, app, webhookJSON(t, "yes-fix", ownerJID(), "sim", "", false, instanceToken))
	fixed := onlyEntry(t, path)
	if fixed.AmountCents != 3000 || fixed.Description != "gastei 30 reais no mercado" {
		t.Fatalf("replacement was not confirmed: %+v", fixed)
	}
}

type fakeSender struct {
	mu    sync.Mutex
	texts []string
	ids   []string
	next  int
}

func (f *fakeSender) send(text string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.texts = append(f.texts, text)
	id := "out"
	if f.next < len(f.ids) {
		id = f.ids[f.next]
	} else if len(f.ids) == 1 {
		id = f.ids[0]
	} else {
		id = "out-" + string(rune('a'+f.next))
	}
	f.next++
	return id, nil
}

func (f *fakeSender) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.texts)
}

func (f *fakeSender) last() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.texts) == 0 {
		return ""
	}
	return f.texts[len(f.texts)-1]
}

type recordingSender struct {
	inner *fakeSender
}

func (r recordingSender) SendText(_ context.Context, _ string, text string) (string, error) {
	return r.inner.send(text)
}

type fakeClassifier struct {
	mu       sync.Mutex
	calls    int
	classify func(text string, categories []string) (whatsapp.Classification, error)
}

func (f *fakeClassifier) Classify(_ context.Context, text string, categories []string) (whatsapp.Classification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.classify != nil {
		return f.classify(text, categories)
	}
	kind := "expense"
	if strings.Contains(strings.ToLower(text), "recebi") {
		kind = "income"
	}
	return whatsapp.Classification{Type: kind, Category: "comida"}, nil
}

func (f *fakeClassifier) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func newWhatsAppApp(t *testing.T, path string) (*fiber.App, string, *fakeSender, *fakeClassifier) {
	t.Helper()
	if path == "" {
		path = filepath.Join(t.TempDir(), "test.sqlite")
	}
	if err := migrate.Up(migrationsDir(t), path); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	conn, err := db.Connect(path)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	send := &fakeSender{}
	class := &fakeClassifier{}
	now := time.Date(2026, 10, 1, 15, 0, 0, 0, ledger.SaoPaulo())
	cfg := config.Config{
		SQLitePath:            path,
		SessionTTL:            30 * 24 * time.Hour,
		WhatsAppInstanceToken: instanceToken,
		WhatsAppOwnerPhone:    ownerPhone,
	}
	app := server.NewWith(conn, cfg, server.Dependencies{
		Sender:     recordingSender{inner: send},
		Classifier: class,
		Now:        func() time.Time { return now },
	})
	return app, path, send, class
}

func ownerJID() string {
	return ownerPhone + "@s.whatsapp.net"
}

func webhookJSON(t *testing.T, id, chat, conversation, extended string, group bool, token string) []byte {
	t.Helper()
	payload := map[string]any{
		"event":         "Message",
		"instanceToken": token,
		"data": map[string]any{
			"Info": map[string]any{
				"Chat":      chat,
				"ID":        id,
				"IsFromMe":  true,
				"IsGroup":   group,
				"MediaType": "",
			},
			"Message": map[string]any{
				"conversation": conversation,
			},
		},
	}
	if extended == "audio" {
		payload["data"].(map[string]any)["Info"].(map[string]any)["MediaType"] = "audio"
		payload["data"].(map[string]any)["Message"].(map[string]any)["audioMessage"] = map[string]any{}
		payload["data"].(map[string]any)["Message"].(map[string]any)["conversation"] = ""
	} else if extended != "" {
		payload["data"].(map[string]any)["Message"].(map[string]any)["extendedTextMessage"] = map[string]string{"text": extended}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func postWhatsApp(t *testing.T, app *fiber.App, body []byte) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func onlyEntry(t *testing.T, path string) ledger.Entry {
	t.Helper()
	conn, err := db.Connect(path)
	if err != nil {
		t.Fatal(err)
	}
	var entries []ledger.Entry
	if err := conn.Preload("Category").Find(&entries).Error; err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	return entries[0]
}

func entryCount(t *testing.T, path string) int {
	t.Helper()
	conn, err := db.Connect(path)
	if err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := conn.Model(&ledger.Entry{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return int(n)
}
