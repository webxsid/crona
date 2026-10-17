package viewhelpers

import "github.com/charmbracelet/lipgloss"

func SplitVertical(total, topMin, bottomMin, topPreferred int) (int, int) {
	if total < topMin+bottomMin {
		if total <= bottomMin {
			return 0, total
		}
		return total - bottomMin, bottomMin
	}
	top := topPreferred
	top = max(top, topMin)
	top = min(top, total-bottomMin)
	return top, total - top
}

func SplitHorizontal(total, leftMin, rightMin, leftPreferred int) (int, int) {
	if total < leftMin+rightMin {
		if total <= rightMin {
			return 0, total
		}
		return total - rightMin, rightMin
	}
	left := leftPreferred
	left = max(left, leftMin)
	left = min(left, total-rightMin)
	return left, total - left
}

func RenderedLineCount(lines []string) int {
	total := 0
	for _, line := range lines {
		h := lipgloss.Height(line)
		h = max(h, 1)
		total += h
	}
	return total
}

type WidthClass int

const (
	WidthNarrow WidthClass = iota
	WidthMedium
	WidthWide
)

func ClassFor(width int) WidthClass {
	switch {
	case width >= 160:
		return WidthWide
	case width >= 100:
		return WidthMedium
	default:
		return WidthNarrow
	}
}

// ColumnGutter is the width of the " │ " separator between side-by-side sections.
const ColumnGutter = 3

// Columns splits total into proportional widths, leaving room for gutters between them.
func Columns(total int, weights ...int) []int {
	out := make([]int, len(weights))
	if len(weights) == 0 {
		return out
	}
	available := max(0, total-ColumnGutter*(len(weights)-1))
	weightSum := 0
	for _, weight := range weights {
		weightSum += weight
	}
	if weightSum <= 0 {
		return out
	}
	used := 0
	for i, weight := range weights {
		out[i] = available * weight / weightSum
		used += out[i]
	}
	out[0] += available - used
	return out
}
