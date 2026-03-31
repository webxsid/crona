package helpers

import (
	"strings"

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

func ExportPresetLabel(asset sharedtypes.ExportTemplateAsset) string {
	format := "Markdown"
	if asset.AssetKind == sharedtypes.ExportAssetKindTemplatePDF || asset.AssetKind == sharedtypes.ExportAssetKindTemplatePDFHTML || asset.AssetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
		format = "PDF"
	}
	kind := strings.ReplaceAll(string(asset.ReportKind), "_", " ")
	if kind != "" {
		kind = strings.ToUpper(kind[:1]) + kind[1:]
	}
	return kind + " " + format + " Style"
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
		return "WeasyPrint was not detected.\n\nInstall weasyprint and press R in Config to rescan PDF support."
	}
	return "Renderer\n" + status.PDFRendererName + "\n\nPath\n" + status.PDFRendererPath + "\n\nPress R in Config to rescan available PDF tools."
}
