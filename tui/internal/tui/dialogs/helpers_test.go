package dialogs

import "testing"

func TestParseEstimateInputAcceptsFlexibleDurations(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		want  int
		isNil bool
	}{
		{name: "empty", raw: "", isNil: true},
		{name: "plain minutes", raw: "90", want: 90},
		{name: "minutes suffix", raw: "90m", want: 90},
		{name: "mixed hours and minutes", raw: "1h30m", want: 90},
		{name: "decimal hours", raw: "1.5h", want: 90},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseEstimateInput(tc.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.isNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", *got)
				}
				return
			}
			if got == nil || *got != tc.want {
				t.Fatalf("expected %d, got %v", tc.want, got)
			}
		})
	}
}

func TestParseOptionalDurationHoursAcceptsDomainFriendlyInputs(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "bare hours", raw: "8", want: 8},
		{name: "decimal hours", raw: "7.5", want: 7.5},
		{name: "hours suffix", raw: "7.5h", want: 7.5},
		{name: "mixed duration", raw: "7h30m", want: 7.5},
		{name: "minutes duration", raw: "450m", want: 7.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseOptionalDurationHours(tc.raw, "Sleep")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || *got != tc.want {
				t.Fatalf("expected %.2f, got %v", tc.want, got)
			}
		})
	}
}

func TestFormatDurationInputsUseCompactDurations(t *testing.T) {
	minutes := 90
	if got := FormatDurationMinutesInput(&minutes); got != "1h30m" {
		t.Fatalf("expected 1h30m, got %q", got)
	}

	hours := 7.5
	if got := FormatDurationHoursInput(&hours); got != "7h30m" {
		t.Fatalf("expected 7h30m, got %q", got)
	}
}
