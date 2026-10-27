package capture

import (
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday, "wednesday": time.Wednesday,
	"thursday": time.Thursday, "friday": time.Friday, "saturday": time.Saturday,
}

// Abbreviations are accepted only in `@` annotations, where they cannot be title words.
var weekdayShort = map[string]time.Weekday{
	"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday, "tues": time.Tuesday,
	"wed": time.Wednesday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday,
	"fri": time.Friday, "sat": time.Saturday,
}

var months = map[string]time.Month{
	"jan": time.January, "feb": time.February, "mar": time.March, "apr": time.April,
	"may": time.May, "jun": time.June, "jul": time.July, "aug": time.August,
	"sep": time.September, "sept": time.September, "oct": time.October, "nov": time.November, "dec": time.December,
	"january": time.January, "february": time.February, "march": time.March, "april": time.April,
	"june": time.June, "july": time.July, "august": time.August, "september": time.September,
	"october": time.October, "november": time.November, "december": time.December,
}

// parseDateWord reads an `@` annotation: today, tomorrow, a weekday, or YYYY-MM-DD.
func parseDateWord(word, today string) (string, bool) {
	base, err := time.Parse(dateLayout, today)
	if err != nil {
		return "", false
	}
	word = strings.ToLower(word)
	switch word {
	case "today", "tonight":
		return today, true
	case "tomorrow":
		return base.AddDate(0, 0, 1).Format(dateLayout), true
	}
	if day, ok := weekdays[word]; ok {
		return upcoming(base, day).Format(dateLayout), true
	}
	if day, ok := weekdayShort[word]; ok {
		return upcoming(base, day).Format(dateLayout), true
	}
	if date, err := time.Parse(dateLayout, word); err == nil {
		return date.Format(dateLayout), true
	}
	return "", false
}

// upcoming is the next day falling on weekday, today included.
func upcoming(base time.Time, day time.Weekday) time.Time {
	return base.AddDate(0, 0, (int(day)-int(base.Weekday())+7)%7)
}

// trailingDate matches a date phrase ending words and returns the date and how many words
// it used, or 0 when the words do not end in one.
func trailingDate(words []string, today string) (string, int) {
	base, err := time.Parse(dateLayout, today)
	if err != nil {
		return "", 0
	}
	tail := func(n int) []string {
		if len(words) < n {
			return nil
		}
		out := make([]string, n)
		for i, w := range words[len(words)-n:] {
			out[i] = strings.ToLower(w)
		}
		return out
	}
	if t := tail(3); t != nil {
		if t[0] == "in" {
			if n, err := strconv.Atoi(t[1]); err == nil && n > 0 {
				switch strings.TrimSuffix(t[2], "s") {
				case "day":
					return base.AddDate(0, 0, n).Format(dateLayout), 3
				case "week":
					return base.AddDate(0, 0, 7*n).Format(dateLayout), 3
				}
			}
		}
		if t[0] == "on" {
			if date, ok := monthDay(t[1], t[2], base); ok {
				return date, 3
			}
		}
	}
	if t := tail(2); t != nil {
		if day, ok := weekdays[t[1]]; ok {
			switch t[0] {
			case "on", "this":
				return upcoming(base, day).Format(dateLayout), 2
			case "next":
				return upcoming(base, day).AddDate(0, 0, 7).Format(dateLayout), 2
			}
		}
		if t[0] == "next" && t[1] == "week" {
			return upcoming(base.AddDate(0, 0, 1), time.Monday).Format(dateLayout), 2
		}
		if t[0] == "on" {
			if date, err := time.Parse(dateLayout, t[1]); err == nil {
				return date.Format(dateLayout), 2
			}
		}
	}
	if t := tail(1); t != nil {
		switch t[0] {
		case "today", "tonight":
			return today, 1
		case "tomorrow":
			return base.AddDate(0, 0, 1).Format(dateLayout), 1
		}
		if day, ok := weekdays[t[0]]; ok {
			return upcoming(base, day).Format(dateLayout), 1
		}
	}
	return "", 0
}

// monthDay reads "oct 30" or "30 oct" as the next such date, today included.
func monthDay(a, b string, base time.Time) (string, bool) {
	month, ok := months[a]
	day, err := strconv.Atoi(b)
	if !ok {
		month, ok = months[b]
		day, err = strconv.Atoi(a)
	}
	if !ok || err != nil || day < 1 || day > 31 {
		return "", false
	}
	date := time.Date(base.Year(), month, day, 0, 0, 0, 0, time.UTC)
	if date.Day() != day {
		return "", false
	}
	if date.Before(base) {
		date = date.AddDate(1, 0, 0)
	}
	return date.Format(dateLayout), true
}

// trailingEstimate matches "for 45 min", "for 1h", "takes 20m" ending words.
func trailingEstimate(words []string) (int, int) {
	n := len(words)
	if n >= 2 && isEstimateLead(words[n-2]) {
		if minutes, ok := parseDuration(words[n-1]); ok {
			return minutes, 2
		}
	}
	if n >= 3 && isEstimateLead(words[n-3]) {
		if minutes, ok := durationWords(words[n-2], words[n-1]); ok {
			return minutes, 3
		}
	}
	return 0, 0
}

func isEstimateLead(word string) bool {
	word = strings.ToLower(word)
	return word == "for" || word == "takes"
}

// durationWords reads a number and a unit word: "45 min", "1.5 hours".
func durationWords(number, unit string) (int, bool) {
	value, err := strconv.ParseFloat(number, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	switch strings.ToLower(unit) {
	case "m", "min", "mins", "minute", "minutes":
		return roundMinutes(value), true
	case "h", "hr", "hrs", "hour", "hours":
		return roundMinutes(value * 60), true
	}
	return 0, false
}

// parseDuration reads 45m, 1h, 1h30m, 1.5h.
func parseDuration(raw string) (int, bool) {
	raw = strings.ToLower(raw)
	total := 0.0
	seen := false
	for raw != "" {
		end := strings.IndexAny(raw, "hm")
		if end <= 0 {
			return 0, false
		}
		value, err := strconv.ParseFloat(raw[:end], 64)
		if err != nil || value < 0 {
			return 0, false
		}
		if raw[end] == 'h' {
			value *= 60
		}
		total += value
		seen = true
		raw = raw[end+1:]
	}
	if !seen || total <= 0 {
		return 0, false
	}
	return roundMinutes(total), true
}

func roundMinutes(value float64) int {
	return int(value + 0.5)
}
