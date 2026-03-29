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
}

func Build(exportAssets *api.ExportAssetStatus) []Item {
	if exportAssets == nil {
		return nil
	}
	items := make([]Item, 0, len(exportAssets.TemplateAssets)+3)
	for _, asset := range exportAssets.TemplateAssets {
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
