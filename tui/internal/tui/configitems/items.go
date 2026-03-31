package configitems

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
	helperpkg "crona/tui/internal/tui/helpers"
)

type Item struct {
	Label       string
	Value       string
	Path        string
	DetailTitle string
	DetailMeta  string
	DetailBody  string
	Editable    bool
	Mutable     bool
	ActionHint  string
	ReportKind  sharedtypes.ExportReportKind
	AssetKind   sharedtypes.ExportAssetKind
	Resettable  bool
	DialogKind  string
	PresetID    string
	PresetStyle bool
}

func Build(exportAssets *api.ExportAssetStatus) []Item {
	if exportAssets == nil {
		return nil
	}
	items := make([]Item, 0, len(exportAssets.TemplateAssets)+3)
	for _, asset := range exportAssets.TemplateAssets {
		if isHiddenNarrativePDFCSSAsset(asset) {
			continue
		}
		if len(asset.Presets) > 0 && asset.SelectedPreset != nil {
			items = append(items, buildPresetItem(asset, exportAssets.TemplateAssets))
			continue
		}
		state := helperpkg.ExportAssetStateLabel(asset)
		detailBody := "Path\n" + asset.UserPath + "\n\nBundled\n" + asset.BundledPath
		detailBody += "\n\nPress e to open in $EDITOR."
		if asset.Resettable {
			detailBody += "\nPress r to replace it with the bundled default."
		}
		items = append(items, Item{
			Label:       asset.Label,
			Value:       state,
			Path:        asset.UserPath,
			DetailTitle: asset.Label,
			DetailMeta:  "Engine " + asset.Engine + "   Source " + asset.ActiveSource + "   State " + state,
			DetailBody:  detailBody,
			Editable:    true,
			ReportKind:  asset.ReportKind,
			AssetKind:   asset.AssetKind,
			Resettable:  asset.Resettable && (asset.Customized || asset.UpdateAvailable),
		})
	}
	items = append(items, Item{
		Label:       "Reports directory",
		Value:       exportAssets.ReportsDir,
		DetailTitle: "Report Output Directory",
		DetailMeta:  helperpkg.ReportsDirMeta(exportAssets),
		DetailBody:  "Generated reports are written under\n" + exportAssets.ReportsDir + "\n\nDefault\n" + exportAssets.DefaultReportsDir + "\n\nPress c to change the directory.\nPress r to restore the default directory.",
		Mutable:     true,
		ActionHint:  "change dir",
		DialogKind:  "edit_export_reports_dir",
	})
	items = append(items, Item{
		Label:       "ICS export directory",
		Value:       exportAssets.ICSDir,
		DetailTitle: "ICS Export Directory",
		DetailMeta:  helperpkg.ICSDirMeta(exportAssets),
		DetailBody:  "Calendar exports are written under\n" + exportAssets.ICSDir + "\n\nDefault\n" + exportAssets.DefaultICSDir + "\n\nUse this directory for Shortcuts, Folder Actions, or other local automations.\nPress c to change the directory.\nPress r to restore the default directory.",
		Mutable:     true,
		ActionHint:  "change dir",
		DialogKind:  "edit_export_ics_dir",
	})
	items = append(items, Item{
		Label:       "PDF renderer",
		Value:       helperpkg.PDFRendererStateLabel(exportAssets),
		DetailTitle: "PDF Renderer",
		DetailMeta:  "External renderer discovery",
		DetailBody:  helperpkg.PDFRendererDetailBody(exportAssets),
	})
	return items
}

func buildPresetItem(asset api.ExportTemplateAsset, allAssets []api.ExportTemplateAsset) Item {
	selected := asset.SelectedPreset
	label := selected.Label
	value := label
	if asset.Customized {
		value += " [customized]"
	}
	detailBody := selected.Description
	if detailBody == "" {
		detailBody = "Built-in starter preset."
	}
	detailBody += "\n\nPreview\n" + firstNonEmpty(presetPreviewBody(asset, selected.ID), "No sample preview available.")
	if cssPath := pairedNarrativePDFCSSPath(asset, allAssets); cssPath != "" {
		detailBody += "\n\nPaired stylesheet\n" + cssPath + "\n\nPress e to open the active HTML template in $EDITOR."
		detailBody += "\nOpen the stylesheet separately if you want to tune fonts, spacing, or colors."
	} else {
		detailBody += "\n\nPress e to open the active template in $EDITOR."
	}
	detailBody += "\nPress space to cycle style.\nPress r to restore this style."
	return Item{
		Label:       helperpkg.ExportPresetLabel(asset),
		Value:       value,
		Path:        asset.UserPath,
		DetailTitle: helperpkg.ExportPresetLabel(asset),
		DetailMeta:  "Preset " + selected.Label + "   Template " + helperpkg.ExportAssetStateLabel(asset),
		DetailBody:  detailBody,
		Editable:    true,
		Mutable:     true,
		ActionHint:  "cycle style",
		ReportKind:  asset.ReportKind,
		AssetKind:   asset.AssetKind,
		Resettable:  true,
		PresetID:    selected.ID,
		PresetStyle: true,
	}
}

func isHiddenNarrativePDFCSSAsset(asset api.ExportTemplateAsset) bool {
	if asset.AssetKind != sharedtypes.ExportAssetKindTemplatePDFCSS {
		return false
	}
	return asset.ReportKind == sharedtypes.ExportReportKindDaily || asset.ReportKind == sharedtypes.ExportReportKindWeekly
}

func pairedNarrativePDFCSSPath(asset api.ExportTemplateAsset, allAssets []api.ExportTemplateAsset) string {
	if asset.AssetKind != sharedtypes.ExportAssetKindTemplatePDFHTML {
		return ""
	}
	if asset.ReportKind != sharedtypes.ExportReportKindDaily && asset.ReportKind != sharedtypes.ExportReportKindWeekly {
		return ""
	}
	for _, candidate := range allAssets {
		if candidate.ReportKind == asset.ReportKind && candidate.AssetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
			return candidate.UserPath
		}
	}
	return ""
}

func presetPreviewBody(asset api.ExportTemplateAsset, presetID string) string {
	for _, preset := range asset.Presets {
		if preset.ID == presetID {
			return preset.PreviewBody
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
