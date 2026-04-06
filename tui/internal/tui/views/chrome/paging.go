package viewchrome

func ListWindow(cursor, total, visible int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if visible < 1 {
		visible = 1
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	if total <= visible {
		return 0, total
	}
	start := cursor - visible/2
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > total {
		end = total
		start = end - visible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}
