package runtime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func normalizeUONLeadDateFields(fields map[string]any) {
	if raw := strings.TrimSpace(fmt.Sprint(fields["date_from"])); raw != "" && raw != "<nil>" {
		if from, to, ok := parseRussianTourDateRange(raw); ok {
			fields["date_from"] = from
			if rawTo := strings.TrimSpace(fmt.Sprint(fields["date_to"])); rawTo == "" || rawTo == "<nil>" {
				fields["date_to"] = to
			}
		}
	}
	for _, key := range []string{"date_from", "date_to"} {
		raw := strings.TrimSpace(fmt.Sprint(fields[key]))
		if raw == "" || raw == "<nil>" {
			continue
		}
		if parsed, ok := parseUONLeadDate(raw); ok {
			fields[key] = parsed
		} else {
			// Keep the original free-form date in the lead note, but never send it
			// as a U-ON date field: the provider rejects such leads with HTTP 404.
			delete(fields, key)
		}
	}
}

var russianTourDateRange = regexp.MustCompile(`^(\d{1,2})(?:\s+([а-яё]+)(?:\s+(\d{4}))?)?\s*[—–-]\s*(\d{1,2})\s+([а-яё]+)\s+(\d{4})\s*г\.(?:\s*[—–-].*)?$`)

var russianMonthNumbers = map[string]time.Month{
	"января": time.January, "февраля": time.February, "марта": time.March,
	"апреля": time.April, "мая": time.May, "июня": time.June,
	"июля": time.July, "августа": time.August, "сентября": time.September,
	"октября": time.October, "ноября": time.November, "декабря": time.December,
}

// The tour booking form sends the displayed range and price in selected_date.
// Extract both actual dates before calling U-ON, including ranges across years.
func parseRussianTourDateRange(raw string) (string, string, bool) {
	raw = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "\u00a0", " "))
	parts := russianTourDateRange.FindStringSubmatch(raw)
	if parts == nil {
		return "", "", false
	}
	startDay, _ := strconv.Atoi(parts[1])
	endDay, _ := strconv.Atoi(parts[4])
	endYear, _ := strconv.Atoi(parts[6])
	startYear := endYear
	if parts[3] != "" {
		startYear, _ = strconv.Atoi(parts[3])
	}
	startMonthName := parts[2]
	if startMonthName == "" {
		startMonthName = parts[5]
	}
	startMonth, startOK := russianMonthNumbers[startMonthName]
	endMonth, endOK := russianMonthNumbers[parts[5]]
	if !startOK || !endOK {
		return "", "", false
	}
	start := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)
	end := time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, time.UTC)
	if start.Day() != startDay || start.Month() != startMonth || end.Day() != endDay || end.Month() != endMonth || start.After(end) {
		return "", "", false
	}
	return start.Format("2006-01-02"), end.Format("2006-01-02"), true
}

func parseUONLeadDate(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{"2006-01-02", "02.01.2006"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.Format("2006-01-02"), true
		}
	}
	return "", false
}
