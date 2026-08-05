package utils

import (
	"fmt"
	"strings"
	"time"
)

// GenerateDateRange returns a slice of date strings (YYYY-MM-DD) from start to end inclusive.
func GenerateDateRange(start, end string) ([]string, error) {
	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse("2006-01-02", end)
	if err != nil {
		return nil, err
	}
	var dates []string
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}
	return dates, nil
}

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ParseDateTime parses a check_in/check_out string into time.Time.
// Accepts "HH:mm" (uses the given date), "yyyy-MM-ddTHH:mm" or "yyyy-MM-dd HH:mm:ss" (full datetime).
func ParseDateTime(val, date string) (time.Time, error) {
	if val == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}
	if len(val) == 5 && val[2] == ':' {
		return time.Parse("2006-01-02 15:04:05", date+" "+val+":00")
	}
	if len(val) >= 16 && val[10] == 'T' {
		return time.Parse("2006-01-02T15:04", val)
	}
	if len(val) == 19 && val[10] == ' ' && val[4] == '-' {
		return time.Parse("2006-01-02 15:04:05", val)
	}
	return time.Parse("2006-01-02 15:04:05", val)
}

// NormalizeDate converts driver date values like "2026-06-01T00:00:00Z" to "YYYY-MM-DD".
func NormalizeDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	if len(dateStr) >= 10 && dateStr[4] == '-' && dateStr[7] == '-' {
		return dateStr[:10]
	}
	for _, layout := range []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999",
	} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return dateStr
}

// IsWeekend checks whether a given date falls on a weekend configured by comma-separated day names/abbreviations (e.g. "Fri,Sat").
func IsWeekend(dateStr string, weekendDays string) bool {
	if weekendDays == "" {
		return false
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	dayName := t.Weekday().String()
	dayAbbr := dayName[:3]
	for _, d := range strings.Split(weekendDays, ",") {
		tr := strings.TrimSpace(d)
		if strings.EqualFold(tr, dayAbbr) || strings.EqualFold(tr, dayName) {
			return true
		}
	}
	return false
}
