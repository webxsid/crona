package repositories

import "testing"

func TestClampManageHistoryDays(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{name: "minimum", value: -1, want: 0},
		{name: "today only", value: 0, want: 0},
		{name: "maximum", value: 5, want: 5},
		{name: "above maximum", value: 8, want: 5},
		{name: "float setting payload", value: float64(3), want: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := clampManageHistoryDays(test.value); got != test.want {
				t.Fatalf("clampManageHistoryDays(%v) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}
