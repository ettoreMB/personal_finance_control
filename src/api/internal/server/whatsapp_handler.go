package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
	"github.com/ettoreMB/personal_finance_control/api/internal/whatsapp"
)

type Dependencies struct {
	Sender     whatsapp.Sender
	Classifier whatsapp.Classifier
	Now        func() time.Time
}

type inboundMessage struct {
	MessageID string `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (inboundMessage) TableName() string { return "whatsapp_inbound_messages" }

type outboundMessage struct {
	MessageID string `gorm:"primaryKey"`
	Body      string
	CreatedAt time.Time
}

func (outboundMessage) TableName() string { return "whatsapp_outbound_messages" }

type outboundState struct {
	ID        int `gorm:"primaryKey"`
	LastBody  string
	MatchText int
}

func (outboundState) TableName() string { return "whatsapp_outbound_state" }

type draftRow struct {
	ID              int `gorm:"primaryKey"`
	SourceText      string
	AmountCents     *int
	EntryType       *string
	CategoryID      *uint
	EntryDate       *string
	MissingAmount   int
	MissingType     int
	MissingCategory int
	MissingDate     int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (draftRow) TableName() string { return "whatsapp_draft" }

type evolutionWebhook struct {
	Event         string `json:"event"`
	InstanceToken string `json:"instanceToken"`
	Data          struct {
		Info struct {
			Chat      string `json:"Chat"`
			ID        string `json:"ID"`
			IsGroup   bool   `json:"IsGroup"`
			MediaType string `json:"MediaType"`
		} `json:"Info"`
		Message struct {
			Conversation string          `json:"conversation"`
			Extended     *extendedText   `json:"extendedTextMessage"`
			Audio        json.RawMessage `json:"audioMessage"`
		} `json:"Message"`
	} `json:"data"`
}

type extendedText struct {
	Text string `json:"text"`
}

var errDuplicateMessage = errors.New("duplicate whatsapp message")

func whatsappWebhook(db *gorm.DB, cfg config.Config, deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload evolutionWebhook
		if err := c.BodyParser(&payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		if !tokenMatches(payload.InstanceToken, cfg.WhatsAppInstanceToken) {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		if payload.Event != "Message" {
			return c.SendStatus(fiber.StatusOK)
		}

		info := payload.Data.Info
		text := messageText(payload)
		if info.ID == "" || info.IsGroup || isStatusChat(info.Chat) || !ownerChat(info.Chat, cfg.WhatsAppOwnerPhone) {
			return c.SendStatus(fiber.StatusOK)
		}
		if ignored, err := isOwnOutbound(db, info.ID, text); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read outbound")
		} else if ignored {
			return c.SendStatus(fiber.StatusOK)
		}

		now := time.Now
		if deps.Now != nil {
			now = deps.Now
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&inboundMessage{MessageID: info.ID}).Error; err != nil {
				if isUniqueConstraint(err) {
					return errDuplicateMessage
				}
				return err
			}
			reply, err := handleAcceptedMessage(tx, deps, text, payload, now())
			if err != nil {
				return err
			}
			if reply == "" {
				return nil
			}
			messageID, err := deps.Sender.SendText(context.Background(), cfg.WhatsAppOwnerPhone, reply)
			if err != nil {
				return err
			}
			return rememberOutbound(tx, messageID, reply)
		})
		if errors.Is(err, errDuplicateMessage) {
			return c.SendStatus(fiber.StatusOK)
		}
		if err != nil {
			slog.Error("whatsapp webhook failed", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to handle message")
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func handleAcceptedMessage(tx *gorm.DB, deps Dependencies, text string, payload evolutionWebhook, now time.Time) (string, error) {
	if isAudio(payload) {
		return whatsapp.AudioNotAcceptedReply, nil
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", nil
	}
	switch whatsapp.Command(trimmed) {
	case "sim":
		return confirmDraft(tx)
	case "nao":
		return discardDraft(tx)
	default:
		return replaceDraft(tx, deps, trimmed, now)
	}
}

func confirmDraft(tx *gorm.DB) (string, error) {
	row, err := loadDraft(tx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return whatsapp.NothingToConfirmReply, nil
	}
	if err != nil {
		return "", err
	}
	view, err := draftView(tx, row, false)
	if err != nil {
		return "", err
	}
	if view.MissingAmount || view.MissingType || view.MissingCategory || view.MissingDate {
		return whatsapp.IncompleteConfirmReply(view), nil
	}
	entry, err := buildEntry(tx, entryRequest{
		Type:        *row.EntryType,
		AmountCents: *row.AmountCents,
		EntryDate:   *row.EntryDate,
		Description: row.SourceText,
		CategoryID:  *row.CategoryID,
	})
	if err != nil {
		return "", err
	}
	if err := tx.Create(&entry).Error; err != nil {
		return "", err
	}
	if err := tx.Delete(&draftRow{}, row.ID).Error; err != nil {
		return "", err
	}
	return whatsapp.ConfirmReply(view.TypeLabel, view.CategoryName, view.EntryDate, *row.AmountCents), nil
}

func discardDraft(tx *gorm.DB) (string, error) {
	row, err := loadDraft(tx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return whatsapp.NothingToDiscardReply, nil
	}
	if err != nil {
		return "", err
	}
	if err := tx.Delete(&draftRow{}, row.ID).Error; err != nil {
		return "", err
	}
	return whatsapp.DiscardedReply, nil
}

func replaceDraft(tx *gorm.DB, deps Dependencies, text string, now time.Time) (string, error) {
	_, err := loadDraft(tx)
	replaced := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	var names []string
	if err := tx.Model(&ledger.Category{}).Order("name").Pluck("name", &names).Error; err != nil {
		return "", err
	}
	if deps.Classifier == nil {
		return "", errors.New("whatsapp classifier is not configured")
	}
	classified, err := deps.Classifier.Classify(context.Background(), text, names)
	if err != nil {
		return "", err
	}

	today := ledger.CivilDateIn(now, ledger.SaoPaulo())
	parsed := whatsapp.ParseMessage(text, today)
	row := draftRow{
		ID:         1,
		SourceText: text,
	}
	if parsed.MissingAmount || parsed.AmountCents == nil {
		row.MissingAmount = 1
	} else {
		row.AmountCents = parsed.AmountCents
	}
	if parsed.MissingDate || parsed.EntryDate == nil {
		row.MissingDate = 1
	} else {
		iso := ledger.FormatEntryDate(*parsed.EntryDate)
		row.EntryDate = &iso
	}
	switch classified.Type {
	case ledger.EntryTypeIncome, ledger.EntryTypeExpense:
		row.EntryType = &classified.Type
	default:
		row.MissingType = 1
	}
	if classified.Category != "" {
		var category ledger.Category
		err := tx.Where("name = ?", classified.Category).First(&category).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row.MissingCategory = 1
		} else if err != nil {
			return "", err
		} else {
			row.CategoryID = &category.ID
		}
	} else {
		row.MissingCategory = 1
	}

	if replaced {
		err = tx.Model(&draftRow{ID: 1}).Select("*").Updates(&row).Error
	} else {
		err = tx.Create(&row).Error
	}
	if err != nil {
		return "", err
	}
	view, err := draftView(tx, row, replaced)
	if err != nil {
		return "", err
	}
	return whatsapp.DraftReply(view), nil
}

func loadDraft(tx *gorm.DB) (draftRow, error) {
	var row draftRow
	err := tx.First(&row, 1).Error
	return row, err
}

func draftView(tx *gorm.DB, row draftRow, replaced bool) (whatsapp.DraftView, error) {
	view := whatsapp.DraftView{
		SourceText:      row.SourceText,
		AmountCents:     row.AmountCents,
		MissingAmount:   row.MissingAmount == 1,
		MissingType:     row.MissingType == 1,
		MissingCategory: row.MissingCategory == 1,
		MissingDate:     row.MissingDate == 1,
		Replaced:        replaced,
	}
	if row.EntryType != nil {
		view.TypeLabel = whatsapp.TypeLabel(*row.EntryType)
	}
	if row.EntryDate != nil {
		view.EntryDate = *row.EntryDate
	}
	if row.CategoryID != nil {
		var category ledger.Category
		if err := tx.First(&category, *row.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				view.MissingCategory = true
				view.CategoryName = ""
			} else {
				return whatsapp.DraftView{}, err
			}
		} else {
			view.CategoryName = category.Name
		}
	}
	return view, nil
}

func rememberOutbound(tx *gorm.DB, messageID, body string) error {
	state := outboundState{ID: 1, LastBody: body, MatchText: 0}
	if messageID == "" {
		state.MatchText = 1
	} else if err := tx.Create(&outboundMessage{MessageID: messageID, Body: body}).Error; err != nil {
		return err
	}
	var existing outboundState
	err := tx.First(&existing, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&state).Error
	}
	if err != nil {
		return err
	}
	return tx.Model(&outboundState{ID: 1}).Select("last_body", "match_text").Updates(&state).Error
}

func isOwnOutbound(db *gorm.DB, messageID, text string) (bool, error) {
	var count int64
	if err := db.Model(&outboundMessage{}).Where("message_id = ?", messageID).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	var state outboundState
	err := db.First(&state, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state.MatchText == 1 && text != "" && text == state.LastBody, nil
}

func messageText(payload evolutionWebhook) string {
	if payload.Data.Message.Conversation != "" {
		return payload.Data.Message.Conversation
	}
	if payload.Data.Message.Extended != nil {
		return payload.Data.Message.Extended.Text
	}
	return ""
}

func isAudio(payload evolutionWebhook) bool {
	if strings.EqualFold(payload.Data.Info.MediaType, "audio") {
		return true
	}
	raw := strings.TrimSpace(string(payload.Data.Message.Audio))
	return raw != "" && raw != "null"
}

func ownerChat(chat, ownerPhone string) bool {
	return chatUser(chat) != "" && chatUser(chat) == digits(ownerPhone)
}

func chatUser(jid string) string {
	jid = strings.TrimSpace(jid)
	if i := strings.Index(jid, "@"); i >= 0 {
		jid = jid[:i]
	}
	if i := strings.Index(jid, ":"); i >= 0 {
		jid = jid[:i]
	}
	return digits(jid)
}

func digits(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isStatusChat(chat string) bool {
	return strings.HasSuffix(strings.ToLower(chat), "@broadcast")
}

func tokenMatches(got, want string) bool {
	if want == "" {
		return false
	}
	sumGot := sha256.Sum256([]byte(got))
	sumWant := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(sumGot[:], sumWant[:]) == 1
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
