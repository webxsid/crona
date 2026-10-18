package state

import "time"

// DefaultReviewWindowDays is the Review window when none has been chosen.
const DefaultReviewWindowDays = 7

// ReviewRange is the Review window: days ending on anchor while date is inside it,
// otherwise days ending on date, so moving the date out of the window carries it along.
func ReviewRange(date, anchor string, days int) (string, string) {
	if days < 1 {
		days = DefaultReviewWindowDays
	}
	end := anchor
	if end == "" || date > end || date < shiftDate(end, -(days-1)) {
		end = date
	}
	return shiftDate(end, -(days - 1)), end
}

func shiftDate(date string, days int) string {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format(time.DateOnly)
}
