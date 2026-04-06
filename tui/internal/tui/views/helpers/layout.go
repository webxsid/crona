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
	if top < topMin {
		top = topMin
	}
	if top > total-bottomMin {
		top = total - bottomMin
	}
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
	if left < leftMin {
		left = leftMin
	}
	if left > total-rightMin {
		left = total - rightMin
	}
	return left, total - left
}

func RenderedLineCount(lines []string) int {
	total := 0
	for _, line := range lines {
		h := lipgloss.Height(line)
		if h < 1 {
			h = 1
		}
		total += h
	}
	return total
}
