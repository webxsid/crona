package helpers

import (
	"crona/tui/internal/api"

	"github.com/charmbracelet/bubbles/viewport"
)

func ScratchpadPaneSize(mainContentWidth, contentHeight int) (int, int) {
	availH := contentHeight
	if availH < 4 {
		availH = 4
	}
	return mainContentWidth, availH
}

func ScratchpadRenderWidth(mainContentWidth, contentHeight int) int {
	paneW, _ := ScratchpadPaneSize(mainContentWidth, contentHeight)
	return maxInt(20, paneW-6)
}

func SyncScratchpadViewport(v viewport.Model, mainContentWidth, contentHeight int, rendered string) viewport.Model {
	paneW, paneH := ScratchpadPaneSize(mainContentWidth, contentHeight)
	contentW := maxInt(20, paneW-6)
	contentH := paneH - 7
	if contentH < 1 {
		contentH = 1
	}
	v.Width = contentW
	v.Height = contentH
	if rendered != "" {
		v.SetContent(rendered)
	}
	return v
}

func ScratchpadMetaAt(scratchpads []api.ScratchPad, idx int) *api.ScratchPad {
	if idx < 0 || idx >= len(scratchpads) {
		return nil
	}
	pad := scratchpads[idx]
	return &api.ScratchPad{
		ID:           pad.ID,
		Name:         pad.Name,
		Path:         pad.Path,
		Pinned:       pad.Pinned,
		LastOpenedAt: pad.LastOpenedAt,
	}
}

func ScratchpadTabIndexByID(scratchpads []api.ScratchPad, id string) int {
	for i, pad := range scratchpads {
		if pad.ID == id {
			return i
		}
	}
	return -1
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
