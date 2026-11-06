package config

import (
	"fmt"

	"crona.work/tui/internal/api"
	configitems "crona.work/tui/internal/tui/configitems"
)

func configItems(status *api.ExportAssetStatus) []string {
	if status == nil {
		return nil
	}
	built := configitems.Build(status)
	items := make([]string, 0, len(built))
	for _, item := range built {
		items = append(items, fmt.Sprintf("%-24s %s", item.Label, item.Value))
	}
	return items
}
