package helpers

import (
	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func ReportsDirMeta(status *api.ExportAssetStatus) string {
	if status == nil {
		return ""
	}
	if status.ReportsDirCustomized {
		return "Mode file export   Source custom"
	}
	return "Mode file export   Source default"
}

func ICSDirMeta(status *api.ExportAssetStatus) string {
	if status == nil {
		return ""
	}
	if status.ICSDirCustomized {
		return "Mode calendar export   Source custom"
	}
	return "Mode calendar export   Source default"
}

func ExportAssetStateLabel(asset sharedtypes.ExportTemplateAsset) string {
	if asset.Resettable {
		switch {
		case asset.Customized:
			return "customized"
		case asset.UpdateAvailable:
			return "new default available"
		default:
			return "default"
		}
	}
	return Truncate(asset.UserPath, 28)
}

func PDFRendererStateLabel(status *api.ExportAssetStatus) string {
	if status == nil {
		return ""
	}
	if status.PDFRendererAvailable {
		return status.PDFRendererName
	}
	return "unavailable"
}

func PDFRendererDetailBody(status *api.ExportAssetStatus) string {
	if status == nil {
		return ""
	}
	if !status.PDFRendererAvailable {
		return "No supported PDF renderer detected.\n\nInstall pandoc with a supported PDF engine and press R in Config to rescan."
	}
	return "Renderer\n" + status.PDFRendererName + "\n\nPath\n" + status.PDFRendererPath + "\n\nPress R in Config to rescan available PDF tools."
}
