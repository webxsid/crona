package wellbeing

import (
	"fmt"
	"math"

	"crona/tui/internal/api"
	types "crona/tui/internal/tui/views/types"
)

func trendStrips(theme types.Theme, state types.ContentState) []string {
	if len(state.MetricsRange) == 0 {
		return nil
	}
	window := state.MetricsRange
	if len(window) > 7 {
		window = window[len(window)-7:]
	}
	mood := make([]float64, 0, len(window))
	energy := make([]float64, 0, len(window))
	work := make([]float64, 0, len(window))
	recovery := make([]float64, 0, len(window))
	for _, day := range window {
		if day.CheckIn != nil {
			mood = append(mood, float64(day.CheckIn.Mood))
			energy = append(energy, float64(day.CheckIn.Energy))
		} else {
			mood = append(mood, 0)
			energy = append(energy, 0)
		}
		work = append(work, float64(day.WorkedSeconds)/3600.0)
		sleep := 0.0
		if day.CheckIn != nil {
			if day.CheckIn.SleepHours != nil {
				sleep = clamp01(*day.CheckIn.SleepHours / 8.0)
			} else if day.CheckIn.SleepScore != nil {
				sleep = clamp01(float64(*day.CheckIn.SleepScore) / 100.0)
			}
		}
		breakRatio := 0.0
		if day.WorkedSeconds > 0 {
			breakRatio = float64(day.RestSeconds) / float64(day.WorkedSeconds)
		}
		recovery = append(recovery, clamp01((sleep+clamp01(breakRatio/0.2))/2.0))
	}
	return []string{
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Mood"), sparkline(mood, 1.0, 5.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Energy"), sparkline(energy, 1.0, 5.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Work"), sparkline(work, 0.0, 8.0)),
		fmt.Sprintf("%s  %s", theme.StyleHeader.Render("Recovery"), sparkline(recovery, 0.0, 1.0)),
	}
}

func sparkline(values []float64, minValue, maxValue float64) string {
	if len(values) == 0 {
		return "-"
	}
	glyphs := []rune("▁▂▃▄▅▆▇█")
	span := maxValue - minValue
	if span <= 0 {
		span = 1
	}
	out := make([]rune, 0, len(values))
	for _, value := range values {
		norm := clamp01((value - minValue) / span)
		idx := int(math.Round(norm * float64(len(glyphs)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(glyphs) {
			idx = len(glyphs) - 1
		}
		out = append(out, glyphs[idx])
	}
	return string(out)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func countsForCheckInStreak(checkIn *api.DailyCheckIn) bool {
	if checkIn == nil {
		return false
	}
	return len(checkIn.CreatedAt) >= 10 && checkIn.CreatedAt[:10] == checkIn.Date
}
