package whatsapp

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	numericDatePattern = regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})(?:/(\d{2,4}))?\b`)
	anteontemPattern   = regexp.MustCompile(`(?i)\banteontem\b`)
	ontemPattern       = regexp.MustCompile(`(?i)\bontem\b`)
	amountPatterns     = []*regexp.Regexp{
		regexp.MustCompile(`(?i)R\$\s*(\d{1,3}(?:\.\d{3})*(?:,\d{1,2})?|\d+(?:,\d{1,2})?)`),
		regexp.MustCompile(`(?i)(\d{1,3}(?:\.\d{3})*(?:,\d{1,2})?|\d+(?:,\d{1,2})?)\s*reais?`),
		regexp.MustCompile(`(?:^|[^\d])(\d{1,3}(?:\.\d{3})*,\d{1,2}|\d+,\d{1,2}|\d+)\b`),
	}
)

type ParsedMessage struct {
	AmountCents   *int
	MissingAmount bool
	EntryDate     *time.Time
	MissingDate   bool
}

func ParseMessage(text string, today time.Time) ParsedMessage {
	parsed := ParsedMessage{}
	if date, missing, ok := parseDate(text, today); ok {
		if missing {
			parsed.MissingDate = true
		} else {
			parsed.EntryDate = &date
		}
	} else {
		day := civil(today)
		parsed.EntryDate = &day
	}

	if cents, ok := parseAmount(maskDates(text)); ok && cents > 0 {
		parsed.AmountCents = &cents
	} else {
		parsed.MissingAmount = true
	}
	return parsed
}

func parseDate(text string, today time.Time) (time.Time, bool, bool) {
	if match := numericDatePattern.FindStringSubmatch(text); match != nil {
		day, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		year := today.Year()
		if match[3] != "" {
			if len(match[3]) != 4 {
				return time.Time{}, true, true
			}
			year, _ = strconv.Atoi(match[3])
		}
		if !validCivil(year, month, day) {
			return time.Time{}, true, true
		}
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), false, true
	}
	if anteontemPattern.MatchString(text) {
		return civil(today).AddDate(0, 0, -2), false, true
	}
	if ontemPattern.MatchString(text) {
		return civil(today).AddDate(0, 0, -1), false, true
	}
	return time.Time{}, false, false
}

func validCivil(year, month, day int) bool {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t.Year() == year && int(t.Month()) == month && t.Day() == day
}

func civil(today time.Time) time.Time {
	return time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
}

func maskDates(text string) string {
	return numericDatePattern.ReplaceAllStringFunc(text, func(s string) string {
		return strings.Repeat(" ", len(s))
	})
}

func parseAmount(text string) (int, bool) {
	bestAt := -1
	bestPriority := 0
	bestRaw := ""
	for priority, pattern := range amountPatterns {
		matches := pattern.FindAllStringSubmatchIndex(text, -1)
		for _, match := range matches {
			if len(match) < 4 || match[2] < 0 {
				continue
			}
			at := match[2]
			if bestAt >= 0 && (at > bestAt || (at == bestAt && priority >= bestPriority)) {
				continue
			}
			bestAt = at
			bestPriority = priority
			bestRaw = text[match[2]:match[3]]
		}
	}
	if bestAt < 0 {
		return 0, false
	}
	return parseBRL(bestRaw)
}

func parseBRL(raw string) (int, bool) {
	raw = strings.ReplaceAll(raw, ".", "")
	reaisPart, fracPart, hasFrac := strings.Cut(raw, ",")
	reais, err := strconv.Atoi(reaisPart)
	if err != nil || reais < 0 {
		return 0, false
	}
	cents := 0
	if hasFrac {
		if len(fracPart) == 1 {
			fracPart += "0"
		}
		if len(fracPart) > 2 {
			fracPart = fracPart[:2]
		}
		cents, err = strconv.Atoi(fracPart)
		if err != nil {
			return 0, false
		}
	}
	return reais*100 + cents, true
}
