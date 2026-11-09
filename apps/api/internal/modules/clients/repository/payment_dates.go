package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func AdvanceDueDate(current string, period string) (*string, error) {
	period = strings.TrimSpace(period)
	if period == "" {
		period = "monthly"
	}
	if period == "once" {
		return nil, nil
	}
	base := time.Now().UTC()
	if strings.TrimSpace(current) != "" {
		parsed, err := parseDateOnly(current)
		if err != nil {
			return nil, err
		}
		base = parsed
	}
	var next time.Time
	switch period {
	case "weekly":
		next = base.AddDate(0, 0, 7)
	case "monthly":
		next = base.AddDate(0, 1, 0)
	case "quarterly":
		next = base.AddDate(0, 3, 0)
	case "semiannual":
		next = base.AddDate(0, 6, 0)
	case "yearly":
		next = base.AddDate(1, 0, 0)
	default:
		return nil, fmt.Errorf("unknown billing period %q", period)
	}
	out := next.Format("2006-01-02")
	return &out, nil
}

func PaymentPeriodEnd(start string, period string) string {
	parsed, err := parseDateOnly(start)
	if err != nil {
		return strings.TrimSpace(start)
	}
	period = strings.TrimSpace(period)
	if period == "" {
		period = "monthly"
	}
	var end time.Time
	switch period {
	case "weekly":
		end = parsed.AddDate(0, 0, 7).AddDate(0, 0, -1)
	case "monthly":
		end = parsed.AddDate(0, 1, 0).AddDate(0, 0, -1)
	case "quarterly":
		end = parsed.AddDate(0, 3, 0).AddDate(0, 0, -1)
	case "semiannual":
		end = parsed.AddDate(0, 6, 0).AddDate(0, 0, -1)
	case "yearly":
		end = parsed.AddDate(1, 0, 0).AddDate(0, 0, -1)
	default:
		end = parsed
	}
	return end.Format("2006-01-02")
}

func parsePaymentDate(value string) (string, error) {
	parsed, err := parseDateOnly(value)
	if err != nil {
		return "", err
	}
	return parsed.Format("2006-01-02"), nil
}

func parsePaymentTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now().UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	parsed, err := parseDateOnly(value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func parseDateOnly(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	for _, layout := range []string{"2006-01-02", "02.01.2006"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, fmt.Errorf("expected date in DD.MM.YYYY or YYYY-MM-DD format")
}

func paymentPeriodLabel(start, end string) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" {
		return ""
	}
	parsedStart, err := time.Parse("2006-01-02", start)
	if err != nil {
		return start
	}
	if end != "" {
		parsedEnd, err := time.Parse("2006-01-02", end)
		if err == nil {
			monthEnd := time.Date(parsedStart.Year(), parsedStart.Month()+1, 0, 0, 0, 0, 0, time.UTC)
			if parsedStart.Day() == 1 && parsedEnd.Format("2006-01-02") == monthEnd.Format("2006-01-02") {
				return russianMonthName(parsedStart.Month()) + " " + strconv.Itoa(parsedStart.Year())
			}
		}
	}
	return parsedStart.Format("02.01.2006")
}

func russianMonthName(month time.Month) string {
	switch month {
	case time.January:
		return "январь"
	case time.February:
		return "февраль"
	case time.March:
		return "март"
	case time.April:
		return "апрель"
	case time.May:
		return "май"
	case time.June:
		return "июнь"
	case time.July:
		return "июль"
	case time.August:
		return "август"
	case time.September:
		return "сентябрь"
	case time.October:
		return "октябрь"
	case time.November:
		return "ноябрь"
	case time.December:
		return "декабрь"
	default:
		return ""
	}
}
