package whatsapp

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	AudioNotAcceptedReply = "Áudio ainda não entra."
	NothingToConfirmReply = "Não há o que confirmar."
	NothingToDiscardReply = "Não há o que descartar."
	DiscardedReply        = "Descartei o rascunho."
	MissingAmountText     = "falta o valor"
	MissingCategoryText   = "falta a categoria"
	MissingDateText       = "falta a data"
	MissingTypeText       = "falta se é ganho ou gasto"
)

type DraftView struct {
	SourceText      string
	AmountCents     *int
	TypeLabel       string
	CategoryName    string
	EntryDate       string
	MissingAmount   bool
	MissingType     bool
	MissingCategory bool
	MissingDate     bool
	Replaced        bool
}

func FormatBRL(cents int) string {
	reais := cents / 100
	frac := cents % 100
	raw := strconv.Itoa(reais)
	var b strings.Builder
	for i := 0; i < len(raw); i++ {
		if i > 0 && (len(raw)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(raw[i])
	}
	return fmt.Sprintf("R$ %s,%02d", b.String(), frac)
}

func FormatDateBR(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("02/01/2006")
}

func TypeLabel(entryType string) string {
	switch entryType {
	case "income":
		return "ganho"
	case "expense":
		return "gasto"
	default:
		return ""
	}
}

func DraftReply(view DraftView) string {
	var b strings.Builder
	if view.Replaced {
		b.WriteString("Substituí o rascunho anterior.\n\n")
	}
	if view.MissingAmount || view.MissingType || view.MissingCategory || view.MissingDate {
		b.WriteString("Rascunho incompleto\n")
	} else {
		b.WriteString("Rascunho de " + view.TypeLabel + "\n")
	}
	b.WriteString("Valor: " + amountLine(view) + "\n")
	if view.MissingType {
		b.WriteString("Tipo: " + MissingTypeText + "\n")
	} else if view.TypeLabel != "" && (view.MissingAmount || view.MissingCategory || view.MissingDate) {
		b.WriteString("Tipo: " + view.TypeLabel + "\n")
	}
	b.WriteString("Categoria: " + categoryLine(view) + "\n")
	b.WriteString("Data: " + dateLine(view) + "\n")
	b.WriteString("Descrição: " + view.SourceText + "\n\n")
	b.WriteString("Responda sim para lançar ou não para descartar.")
	return b.String()
}

func amountLine(view DraftView) string {
	if view.MissingAmount || view.AmountCents == nil {
		return MissingAmountText
	}
	return FormatBRL(*view.AmountCents)
}

func categoryLine(view DraftView) string {
	if view.MissingCategory || view.CategoryName == "" {
		return MissingCategoryText
	}
	return view.CategoryName
}

func dateLine(view DraftView) string {
	if view.MissingDate || view.EntryDate == "" {
		return MissingDateText
	}
	return FormatDateBR(view.EntryDate)
}

func IncompleteConfirmReply(view DraftView) string {
	parts := []string{"Ainda não lancei."}
	if view.MissingAmount {
		parts = append(parts, capitalize(MissingAmountText)+".")
	}
	if view.MissingType {
		parts = append(parts, capitalize(MissingTypeText)+".")
	}
	if view.MissingCategory {
		parts = append(parts, capitalize(MissingCategoryText)+".")
	}
	if view.MissingDate {
		parts = append(parts, capitalize(MissingDateText)+".")
	}
	return strings.Join(parts, " ")
}

func ConfirmReply(typeLabel, categoryName, entryDate string, amountCents int) string {
	return fmt.Sprintf(
		"Lancei %s de %s em %s, em %s.",
		typeLabel,
		FormatBRL(amountCents),
		categoryName,
		FormatDateBR(entryDate),
	)
}

func capitalize(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func Command(text string) string {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "sim":
		return "sim"
	case "não", "nao":
		return "nao"
	default:
		return ""
	}
}
