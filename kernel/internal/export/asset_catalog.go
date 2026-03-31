package export

import sharedtypes "crona/shared/types"

type assetDescriptor struct {
	reportKind        sharedtypes.ExportReportKind
	assetKind         sharedtypes.ExportAssetKind
	label             string
	name              string
	engine            string
	userRelativePath  string
	legacyUserPath    string
	bundledPath       string
	legacyBundledPath string
	fallback          string
	resettable        bool
}

func assetDescriptors() []assetDescriptor {
	return []assetDescriptor{
		{reportKind: sharedtypes.ExportReportKindDaily, assetKind: sharedtypes.ExportAssetKindTemplateMarkdown, label: "Daily report template", name: "daily/report.hbs", engine: "hbs", userRelativePath: "daily/report.hbs", legacyUserPath: "daily-report.user.hbs", bundledPath: "daily/report.default.hbs", legacyBundledPath: "daily-report.default.hbs", fallback: fallbackDailyReportTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindDaily, assetKind: sharedtypes.ExportAssetKindTemplatePDFHTML, label: "Daily PDF HTML template", name: "daily/report.pdf.html.hbs", engine: "hbs", userRelativePath: "daily/report.pdf.html.hbs", legacyUserPath: "daily-report.pdf.user.hbs", bundledPath: "daily/report.pdf.html.default.hbs", legacyBundledPath: "daily-report.pdf.default.hbs", fallback: fallbackDailyReportPDFTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindDaily, assetKind: sharedtypes.ExportAssetKindTemplatePDFCSS, label: "Daily PDF stylesheet", name: "daily/report.pdf.css", engine: "css", userRelativePath: "daily/report.pdf.css", bundledPath: "daily/report.pdf.default.css", fallback: fallbackDailyReportPDFStyles, resettable: true},
		{reportKind: sharedtypes.ExportReportKindDaily, assetKind: sharedtypes.ExportAssetKindVariableDocs, label: "Daily variable docs", name: "daily/report.variables.md", engine: "md", userRelativePath: "daily/report.variables.md", legacyUserPath: "daily-report.variables.md", bundledPath: "daily/report.variables.md", legacyBundledPath: "daily-report.variables.md", fallback: fallbackDailyReportVariables},
		{reportKind: sharedtypes.ExportReportKindWeekly, assetKind: sharedtypes.ExportAssetKindTemplateMarkdown, label: "Weekly report template", name: "weekly/report.hbs", engine: "hbs", userRelativePath: "weekly/report.hbs", legacyUserPath: "weekly-report.user.hbs", bundledPath: "weekly/report.default.hbs", legacyBundledPath: "weekly-report.default.hbs", fallback: fallbackWeeklyReportTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindWeekly, assetKind: sharedtypes.ExportAssetKindTemplatePDFHTML, label: "Weekly PDF HTML template", name: "weekly/report.pdf.html.hbs", engine: "hbs", userRelativePath: "weekly/report.pdf.html.hbs", legacyUserPath: "weekly-report.pdf.user.hbs", bundledPath: "weekly/report.pdf.html.default.hbs", legacyBundledPath: "weekly-report.pdf.default.hbs", fallback: fallbackWeeklyReportPDFTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindWeekly, assetKind: sharedtypes.ExportAssetKindTemplatePDFCSS, label: "Weekly PDF stylesheet", name: "weekly/report.pdf.css", engine: "css", userRelativePath: "weekly/report.pdf.css", bundledPath: "weekly/report.pdf.default.css", fallback: fallbackWeeklyReportPDFStyles, resettable: true},
		{reportKind: sharedtypes.ExportReportKindWeekly, assetKind: sharedtypes.ExportAssetKindVariableDocs, label: "Weekly variable docs", name: "weekly/report.variables.md", engine: "md", userRelativePath: "weekly/report.variables.md", legacyUserPath: "weekly-report.variables.md", bundledPath: "weekly/report.variables.md", legacyBundledPath: "weekly-report.variables.md", fallback: fallbackWeeklyReportVariables},
		{reportKind: sharedtypes.ExportReportKindRepo, assetKind: sharedtypes.ExportAssetKindTemplateMarkdown, label: "Repo report template", name: "repo/report.hbs", engine: "hbs", userRelativePath: "repo/report.hbs", legacyUserPath: "repo-report.user.hbs", bundledPath: "repo/report.default.hbs", legacyBundledPath: "repo-report.default.hbs", fallback: fallbackRepoReportTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindRepo, assetKind: sharedtypes.ExportAssetKindTemplatePDF, label: "Repo PDF template", name: "repo/report.pdf.hbs", engine: "hbs", userRelativePath: "repo/report.pdf.hbs", legacyUserPath: "repo-report.pdf.user.hbs", bundledPath: "repo/report.pdf.default.hbs", legacyBundledPath: "repo-report.pdf.default.hbs", fallback: fallbackRepoReportPDFTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindRepo, assetKind: sharedtypes.ExportAssetKindVariableDocs, label: "Repo variable docs", name: "repo/report.variables.md", engine: "md", userRelativePath: "repo/report.variables.md", legacyUserPath: "repo-report.variables.md", bundledPath: "repo/report.variables.md", legacyBundledPath: "repo-report.variables.md", fallback: fallbackRepoReportVariables},
		{reportKind: sharedtypes.ExportReportKindStream, assetKind: sharedtypes.ExportAssetKindTemplateMarkdown, label: "Stream report template", name: "stream/report.hbs", engine: "hbs", userRelativePath: "stream/report.hbs", legacyUserPath: "stream-report.user.hbs", bundledPath: "stream/report.default.hbs", legacyBundledPath: "stream-report.default.hbs", fallback: fallbackStreamReportTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindStream, assetKind: sharedtypes.ExportAssetKindTemplatePDF, label: "Stream PDF template", name: "stream/report.pdf.hbs", engine: "hbs", userRelativePath: "stream/report.pdf.hbs", legacyUserPath: "stream-report.pdf.user.hbs", bundledPath: "stream/report.pdf.default.hbs", legacyBundledPath: "stream-report.pdf.default.hbs", fallback: fallbackStreamReportPDFTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindStream, assetKind: sharedtypes.ExportAssetKindVariableDocs, label: "Stream variable docs", name: "stream/report.variables.md", engine: "md", userRelativePath: "stream/report.variables.md", legacyUserPath: "stream-report.variables.md", bundledPath: "stream/report.variables.md", legacyBundledPath: "stream-report.variables.md", fallback: fallbackStreamReportVariables},
		{reportKind: sharedtypes.ExportReportKindIssueRollup, assetKind: sharedtypes.ExportAssetKindTemplateMarkdown, label: "Issue rollup template", name: "issue-rollup/report.hbs", engine: "hbs", userRelativePath: "issue-rollup/report.hbs", legacyUserPath: "issue-rollup-report.user.hbs", bundledPath: "issue-rollup/report.default.hbs", legacyBundledPath: "issue-rollup-report.default.hbs", fallback: fallbackIssueRollupReportTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindIssueRollup, assetKind: sharedtypes.ExportAssetKindTemplatePDF, label: "Issue rollup PDF template", name: "issue-rollup/report.pdf.hbs", engine: "hbs", userRelativePath: "issue-rollup/report.pdf.hbs", legacyUserPath: "issue-rollup-report.pdf.user.hbs", bundledPath: "issue-rollup/report.pdf.default.hbs", legacyBundledPath: "issue-rollup-report.pdf.default.hbs", fallback: fallbackIssueRollupReportPDFTemplate, resettable: true},
		{reportKind: sharedtypes.ExportReportKindIssueRollup, assetKind: sharedtypes.ExportAssetKindVariableDocs, label: "Issue rollup variable docs", name: "issue-rollup/report.variables.md", engine: "md", userRelativePath: "issue-rollup/report.variables.md", legacyUserPath: "issue-rollup-report.variables.md", bundledPath: "issue-rollup/report.variables.md", legacyBundledPath: "issue-rollup-report.variables.md", fallback: fallbackIssueRollupReportVariables},
		{reportKind: sharedtypes.ExportReportKindCSV, assetKind: sharedtypes.ExportAssetKindCSVSpec, label: "CSV export spec", name: "csv/export.spec.json", engine: "json", userRelativePath: "csv/export.spec.json", legacyUserPath: "csv-export.spec.json", bundledPath: "csv/export.spec.json", legacyBundledPath: "csv-export.spec.json", fallback: fallbackCSVExportSpec, resettable: true},
		{reportKind: sharedtypes.ExportReportKindCSV, assetKind: sharedtypes.ExportAssetKindCSVDocs, label: "CSV export docs", name: "csv/export.variables.md", engine: "md", userRelativePath: "csv/export.variables.md", legacyUserPath: "csv-export.variables.md", bundledPath: "csv/export.variables.md", legacyBundledPath: "csv-export.variables.md", fallback: fallbackCSVExportVariables},
	}
}

func findAssetDescriptor(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) (assetDescriptor, bool) {
	for _, descriptor := range assetDescriptors() {
		if descriptor.reportKind == reportKind && descriptor.assetKind == assetKind {
			return descriptor, true
		}
	}
	return assetDescriptor{}, false
}
