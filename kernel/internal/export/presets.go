package export

import (
	"fmt"
	"strings"

	sharedtypes "crona/shared/types"
)

type templatePreset struct {
	ID           string
	Label        string
	Description  string
	PreviewTitle string
	PreviewBody  string
	Body         string
}

func presetSelectionForAsset(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind, presetID string) *sharedtypes.ExportTemplatePresetSelection {
	if assetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
		return nil
	}
	for _, preset := range presetsForAsset(reportKind, assetKind) {
		if preset.ID == presetID {
			return &sharedtypes.ExportTemplatePresetSelection{
				ID:          preset.ID,
				Label:       preset.Label,
				Description: preset.Description,
			}
		}
	}
	return nil
}

func presetMetadataForAsset(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) []sharedtypes.ExportTemplatePreset {
	if assetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
		return nil
	}
	presets := presetsForAsset(reportKind, assetKind)
	out := make([]sharedtypes.ExportTemplatePreset, 0, len(presets))
	for _, preset := range presets {
		out = append(out, sharedtypes.ExportTemplatePreset{
			ID:           preset.ID,
			Label:        preset.Label,
			Description:  preset.Description,
			PreviewTitle: preset.PreviewTitle,
			PreviewBody:  preset.PreviewBody,
		})
	}
	return out
}

func presetTemplateBody(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind, presetID string) (string, bool) {
	for _, preset := range presetsForAsset(reportKind, assetKind) {
		if preset.ID == presetID {
			return preset.Body, true
		}
	}
	return "", false
}

func defaultPresetIDForAsset(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) string {
	if len(presetsForAsset(reportKind, assetKind)) == 0 {
		return ""
	}
	return "balanced"
}

func normalizePresetID(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind, presetID string) string {
	presetID = strings.TrimSpace(presetID)
	if presetID == "" {
		return defaultPresetIDForAsset(reportKind, assetKind)
	}
	if _, ok := presetTemplateBody(reportKind, assetKind, presetID); ok {
		return presetID
	}
	return defaultPresetIDForAsset(reportKind, assetKind)
}

func presetsForAsset(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) []templatePreset {
	if assetKind != sharedtypes.ExportAssetKindTemplateMarkdown &&
		assetKind != sharedtypes.ExportAssetKindTemplatePDF &&
		assetKind != sharedtypes.ExportAssetKindTemplatePDFHTML &&
		assetKind != sharedtypes.ExportAssetKindTemplatePDFCSS {
		return nil
	}
	switch reportKind {
	case sharedtypes.ExportReportKindDaily:
		if assetKind == sharedtypes.ExportAssetKindTemplatePDF || assetKind == sharedtypes.ExportAssetKindTemplatePDFHTML {
			return dailyPDFPresets()
		}
		if assetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
			return dailyPDFCSSPresets()
		}
		return dailyMarkdownPresets()
	case sharedtypes.ExportReportKindWeekly:
		if assetKind == sharedtypes.ExportAssetKindTemplatePDF || assetKind == sharedtypes.ExportAssetKindTemplatePDFHTML {
			return weeklyPDFPresets()
		}
		if assetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
			return weeklyPDFCSSPresets()
		}
		return weeklyMarkdownPresets()
	default:
		return nil
	}
}

func dailyMarkdownPresets() []templatePreset {
	return []templatePreset{
		{ID: "brief", Label: "Brief", Description: "Short executive skim with only the strongest signals.", PreviewTitle: "Brief Daily", PreviewBody: "🧭 One screen summary\n\n• Today in 4-6 bullets\n• Short wins and watchouts\n• Tiny issue rollup", Body: dailyBriefMarkdownTemplate},
		{ID: "balanced", Label: "Balanced", Description: "Readable default with highlights, metrics, and compact sections.", PreviewTitle: "Balanced Daily", PreviewBody: "✨ Summary first\n\n• Snapshot\n• Highlights / risks\n• Compact habits and issues", Body: fallbackDailyReportTemplate},
		{ID: "visual", Label: "Visual", Description: "More badges, dividers, and emoji-led scannability.", PreviewTitle: "Visual Daily", PreviewBody: "📊 Card-like sections\n\n• Big top summary\n• Emoji headings\n• Compact visual lists", Body: dailyVisualMarkdownTemplate},
		{ID: "deep", Label: "Deep", Description: "Fuller daily context without the old wall-of-text layout.", PreviewTitle: "Deep Daily", PreviewBody: "📝 Detailed but structured\n\n• Summary\n• Wellbeing\n• Plan accountability\n• Work breakdown", Body: dailyDeepMarkdownTemplate},
	}
}

func dailyPDFPresets() []templatePreset {
	return []templatePreset{
		{ID: "brief", Label: "Brief", Description: "Short PDF summary with strong hierarchy.", PreviewTitle: "Brief Daily PDF", PreviewBody: "🗂 Tight top card + compact work lists", Body: dailyBriefPDFTemplate},
		{ID: "balanced", Label: "Balanced", Description: "Default PDF summary with readable sections.", PreviewTitle: "Balanced Daily PDF", PreviewBody: "📄 Clean sectioned summary with highlights", Body: fallbackDailyReportPDFTemplate},
		{ID: "visual", Label: "Visual", Description: "More section cards and visual markers for print/export.", PreviewTitle: "Visual Daily PDF", PreviewBody: "🎯 Metric blocks, callouts, and icon-led sections", Body: dailyVisualPDFTemplate},
		{ID: "deep", Label: "Deep", Description: "Longer PDF with still-compact structured detail.", PreviewTitle: "Deep Daily PDF", PreviewBody: "📚 Detailed summary plus grouped work sections", Body: dailyDeepPDFTemplate},
	}
}

func weeklyMarkdownPresets() []templatePreset {
	return []templatePreset{
		{ID: "brief", Label: "Brief", Description: "Fast weekly skim with rollup and top days only.", PreviewTitle: "Brief Weekly", PreviewBody: "📅 Weekly rollup in a glance\n\n• Core metrics\n• Best / hardest days", Body: weeklyBriefMarkdownTemplate},
		{ID: "balanced", Label: "Balanced", Description: "Default readable weekly narrative.", PreviewTitle: "Balanced Weekly", PreviewBody: "✨ Rollup + streaks + day snapshots", Body: fallbackWeeklyReportTemplate},
		{ID: "visual", Label: "Visual", Description: "Weekly report with stronger signposting and lighter prose.", PreviewTitle: "Visual Weekly", PreviewBody: "📈 Strong metric framing and emoji-led day cards", Body: weeklyVisualMarkdownTemplate},
		{ID: "deep", Label: "Deep", Description: "More complete weekly recap with compact daily detail.", PreviewTitle: "Deep Weekly", PreviewBody: "📝 Expanded weekly review without the old density", Body: weeklyDeepMarkdownTemplate},
	}
}

func weeklyPDFPresets() []templatePreset {
	return []templatePreset{
		{ID: "brief", Label: "Brief", Description: "Compressed weekly PDF with strong rollup hierarchy.", PreviewTitle: "Brief Weekly PDF", PreviewBody: "📌 Core rollup and short day list", Body: weeklyBriefPDFTemplate},
		{ID: "balanced", Label: "Balanced", Description: "Default weekly PDF summary.", PreviewTitle: "Balanced Weekly PDF", PreviewBody: "📄 Clear rollup and daily snapshots", Body: fallbackWeeklyReportPDFTemplate},
		{ID: "visual", Label: "Visual", Description: "More visual structure and emphasis for PDF export.", PreviewTitle: "Visual Weekly PDF", PreviewBody: "🎨 Card-like rollup with stronger visual cues", Body: weeklyVisualPDFTemplate},
		{ID: "deep", Label: "Deep", Description: "Most detailed weekly PDF while staying structured.", PreviewTitle: "Deep Weekly PDF", PreviewBody: "📚 More complete week review with compact daily sections", Body: weeklyDeepPDFTemplate},
	}
}

func dailyPDFCSSPresets() []templatePreset {
	return []templatePreset{
		{ID: "brief", Body: dailyBriefPDFStyles},
		{ID: "balanced", Body: fallbackDailyReportPDFStyles},
		{ID: "visual", Body: dailyVisualPDFStyles},
		{ID: "deep", Body: dailyDeepPDFStyles},
	}
}

func weeklyPDFCSSPresets() []templatePreset {
	return []templatePreset{
		{ID: "brief", Body: weeklyBriefPDFStyles},
		{ID: "balanced", Body: fallbackWeeklyReportPDFStyles},
		{ID: "visual", Body: weeklyVisualPDFStyles},
		{ID: "deep", Body: weeklyDeepPDFStyles},
	}
}

func presetAssetLabel(reportKind sharedtypes.ExportReportKind, assetKind sharedtypes.ExportAssetKind) string {
	format := "Markdown"
	if assetKind == sharedtypes.ExportAssetKindTemplatePDF || assetKind == sharedtypes.ExportAssetKindTemplatePDFHTML || assetKind == sharedtypes.ExportAssetKindTemplatePDFCSS {
		format = "PDF"
	}
	return fmt.Sprintf("%s %s Style", strings.Title(strings.ReplaceAll(string(reportKind), "_", " ")), format)
}

const dailyBriefMarkdownTemplate = `# 🌤 Daily Snapshot — {{date}}

## At a glance
- Work: {{summary.workedEstimate}}
- Issues: {{summary.issueDoneCount}} done / {{summary.totalIssues}} total
- Habits: {{summary.habitsCompletedCount}} / {{summary.habitsDueCount}}
{{#if checkIn}}- Mood / energy: {{checkIn.mood}} / {{checkIn.energy}}{{/if}}

{{#if highlights}}
## ✅ Highlights
{{#each highlights}}
- {{this}}
{{/each}}
{{/if}}

{{#if risks}}
## ⚠️ Watchouts
{{#each risks}}
- {{this}}
{{/each}}
{{/if}}

## Work
{{#each repos}}
### {{name}}
{{#each streams}}
#### {{name}}
{{#each completedIssues}}
- ✅ #{{id}} {{title}} · {{workedEstimate}}
{{/each}}
{{#each activeIssues}}
- 🔄 #{{id}} {{title}} · {{workedEstimate}}
{{/each}}
{{#each attentionIssues}}
- ⚠️ #{{id}} {{title}} · {{workedEstimate}}
{{/each}}
{{/each}}
{{/each}}
`

const dailyVisualMarkdownTemplate = `# ✨ Daily Pulse — {{date}}

## 🧭 Summary Card
- Health: {{dayHealth}}
- Work / estimate: {{summary.workedEstimate}}
- Issues: {{summary.issueDoneCount}} ✅  {{summary.issueActiveCount}} 🔄  {{summary.issueBlockedCount}} ⛔
- Habits: {{summary.habitsCompletedCount}} / {{summary.habitsDueCount}}
{{#if metrics.burnout}}- Burnout: {{metrics.burnout.level}} ({{metrics.burnout.score}}){{/if}}
{{#if plan}}- Accountability: {{summary.accountabilityScore}} · failed {{summary.planFailedCount}} · delayed {{summary.delayedIssueCount}}{{/if}}

{{#if highlights}}
## 🌟 Wins
{{#each highlights}}
- {{this}}
{{/each}}
{{/if}}

{{#if risks}}
## 🚧 Friction
{{#each risks}}
- {{this}}
{{/each}}
{{/if}}

## 📌 Plan
{{#if plan}}
- Planned: {{plan.summary.plannedCount}}
- Completed: {{plan.summary.completedCount}}
- Failed: {{plan.summary.failedCount}}
{{/if}}

## 🗂 Work Board
{{#each repos}}
### {{name}}
{{#each streams}}
#### {{name}}
{{#if completedIssues}}
Completed
{{#each completedIssues}}
- ✅ #{{id}} {{title}} · {{workedEstimate}}
{{/each}}
{{/if}}
{{#if activeIssues}}
Active
{{#each activeIssues}}
- 🔄 #{{id}} {{title}} · {{workedEstimate}}
{{/each}}
{{/if}}
{{#if attentionIssues}}
Attention
{{#each attentionIssues}}
- ⚠️ #{{id}} {{title}} · {{workedEstimate}}
{{/each}}
{{/if}}
{{/each}}
{{/each}}
`

const dailyDeepMarkdownTemplate = `# 📝 Daily Review — {{date}}

Generated at {{generatedAt}}

## Overview
- Health: {{dayHealth}}
- Work / estimate: {{summary.workedEstimate}}
- Variance: {{summary.varianceTime}}
- Completed issues: {{summary.issueDoneCount}}
- Active issues: {{summary.issueActiveCount}}
- Habits: {{summary.habitsCompletedCount}} / {{summary.habitsDueCount}}
{{#if checkIn}}- Mood / energy: {{checkIn.mood}} / {{checkIn.energy}}{{/if}}
{{#if plan}}- Accountability: {{summary.accountabilityScore}} (failed {{summary.planFailedCount}}, delayed {{summary.delayedIssueCount}}){{/if}}

{{#if highlights}}
## Highlights
{{#each highlights}}
- {{this}}
{{/each}}
{{/if}}

{{#if risks}}
## Risks
{{#each risks}}
- {{this}}
{{/each}}
{{/if}}

{{#if plan}}
## Plan Accountability
- Planned: {{summary.plannedCount}}
- Completed: {{summary.planCompletedCount}}
- Failed: {{summary.planFailedCount}}
- Pending rollback: {{summary.planPendingRollbackCount}}
- Backlog pressure: {{summary.backlogPressure}}
- High-risk issues: {{summary.highRiskIssueCount}}
{{/if}}

{{#if checkIn}}
## Wellbeing
- Mood: {{checkIn.mood}}
- Energy: {{checkIn.energy}}
{{#if checkIn.sleepHours}}- Sleep: {{checkIn.sleepHours}}h{{/if}}
{{#if checkIn.screenTime}}- Screen time: {{checkIn.screenTime}}{{/if}}
{{#if metrics}}- Rest: {{metrics.restTime}}{{/if}}
{{#if checkIn.notes}}

{{checkIn.notes}}
{{/if}}
{{/if}}

## Issue Breakdown
{{#each repos}}
### {{name}}
{{#each streams}}
#### {{name}}
{{#each completedIssues}}
- ✅ #{{id}} {{title}}
  Status: {{status}}
  Worked / estimate: {{workedEstimate}}
{{/each}}
{{#each activeIssues}}
- 🔄 #{{id}} {{title}}
  Status: {{status}}
  Worked / estimate: {{workedEstimate}}
{{/each}}
{{#each attentionIssues}}
- ⚠️ #{{id}} {{title}}
  Status: {{status}}
  Worked / estimate: {{workedEstimate}}
{{/each}}
{{/each}}
{{/each}}
`

const dailyBriefPDFTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Daily Snapshot - {{date}}</title><link rel="stylesheet" href="report.css"></head>
<body class="report daily brief">
<header class="hero"><div class="eyebrow">🌤 Daily snapshot</div><h1>{{date}}</h1></header>
<section class="metric-grid compact">
<article class="metric-card"><div class="metric-label">Work</div><div class="metric-value">{{summary.workedEstimate}}</div></article>
<article class="metric-card"><div class="metric-label">Issues</div><div class="metric-value">{{summary.issueDoneCount}} / {{summary.totalIssues}}</div></article>
<article class="metric-card"><div class="metric-label">Habits</div><div class="metric-value">{{summary.habitsCompletedCount}} / {{summary.habitsDueCount}}</div></article>
{{#if plan}}<article class="metric-card"><div class="metric-label">Accountability</div><div class="metric-value">{{summary.accountabilityScore}}</div></article>{{/if}}
</section>
{{#if highlights}}<section class="section"><h2>Highlights</h2><ul class="bullet-list">{{#each highlights}}<li>{{this}}</li>{{/each}}</ul></section>{{/if}}
{{#if risks}}<section class="section"><h2>Watchouts</h2><ul class="bullet-list">{{#each risks}}<li>{{this}}</li>{{/each}}</ul></section>{{/if}}
</body></html>`

const dailyVisualPDFTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Daily Pulse - {{date}}</title><link rel="stylesheet" href="report.css"></head>
<body class="report daily visual">
<header class="hero accent"><div class="eyebrow">✨ Daily pulse</div><h1>{{date}}</h1><div class="subtle">Generated at {{generatedAt}}</div></header>
<section class="metric-grid">
<article class="metric-card emphasis"><div class="metric-label">Health</div><div class="metric-value">{{dayHealth}}</div></article>
<article class="metric-card emphasis"><div class="metric-label">Work / estimate</div><div class="metric-value">{{summary.workedEstimate}}</div></article>
<article class="metric-card"><div class="metric-label">Issues</div><div class="metric-value">{{summary.issueDoneCount}} ✅ / {{summary.issueActiveCount}} 🔄</div></article>
<article class="metric-card"><div class="metric-label">Habits</div><div class="metric-value">{{summary.habitsCompletedCount}} / {{summary.habitsDueCount}}</div></article>
{{#if plan}}<article class="metric-card"><div class="metric-label">Failed / delayed</div><div class="metric-value">{{summary.planFailedCount}} / {{summary.delayedIssueCount}}</div></article>{{/if}}
</section>
{{#if highlights}}<section class="section"><h2>🌟 Wins</h2><ul class="bullet-list">{{#each highlights}}<li>{{this}}</li>{{/each}}</ul></section>{{/if}}
{{#if risks}}<section class="section"><h2>🚧 Friction</h2><ul class="bullet-list">{{#each risks}}<li>{{this}}</li>{{/each}}</ul></section>{{/if}}
</body></html>`

const dailyDeepPDFTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Daily Review - {{date}}</title><link rel="stylesheet" href="report.css"></head>
<body class="report daily deep">
<header class="hero"><div class="eyebrow">📝 Daily review</div><h1>{{date}}</h1><div class="subtle">Generated at {{generatedAt}}</div></header>
<section class="metric-grid">
<article class="metric-card"><div class="metric-label">Health</div><div class="metric-value">{{dayHealth}}</div></article>
<article class="metric-card"><div class="metric-label">Variance</div><div class="metric-value">{{summary.varianceTime}}</div></article>
{{#if plan}}<article class="metric-card"><div class="metric-label">Accountability</div><div class="metric-value">{{summary.accountabilityScore}}</div></article>{{/if}}
</section>
{{#if highlights}}<section class="section"><h2>Highlights</h2><ul class="bullet-list">{{#each highlights}}<li>{{this}}</li>{{/each}}</ul></section>{{/if}}
{{#if risks}}<section class="section"><h2>Risks</h2><ul class="bullet-list">{{#each risks}}<li>{{this}}</li>{{/each}}</ul></section>{{/if}}
{{#if plan}}<section class="section"><h2>Plan Accountability</h2><div class="row"><span>Failed</span><span>{{summary.planFailedCount}}</span></div><div class="row"><span>Delayed issues</span><span>{{summary.delayedIssueCount}}</span></div><div class="row"><span>Backlog pressure</span><span>{{summary.backlogPressure}}</span></div></section>{{/if}}
<section class="section"><h2>Issues</h2>{{#each repos}}<article class="group"><h3>{{name}}</h3>{{#each streams}}<div class="subgroup"><h4>{{name}}</h4>{{#each completedIssues}}<div class="row row-good"><span>#{{id}} {{title}}</span><span>{{workedEstimate}}</span></div>{{/each}}{{#each activeIssues}}<div class="row"><span>#{{id}} {{title}}</span><span>{{workedEstimate}}</span></div>{{/each}}{{#each attentionIssues}}<div class="row row-warn"><span>#{{id}} {{title}}</span><span>{{workedEstimate}}</span></div>{{/each}}</div>{{/each}}</article>{{/each}}</section>
</body></html>`

const weeklyBriefMarkdownTemplate = `# 📅 Weekly Snapshot

{{startDate}} → {{endDate}}

## At a glance
- Focus days: {{summary.focusDays}}
- Check-ins: {{summary.checkInDays}}
- Worked: {{summary.workedTime}}
- Rest: {{summary.restTime}}
- Completed issues: {{summary.completedIssues}}

## Streaks
- Focus: {{streaks.currentFocusDays}} current / {{streaks.longestFocusDays}} best
- Check-ins: {{streaks.currentCheckInDays}} current / {{streaks.longestCheckInDays}} best

## Days
{{#each days}}
- {{date}} · {{workedTime}} · {{sessionCount}} sessions
{{/each}}
`

const weeklyVisualMarkdownTemplate = `# ✨ Weekly Pulse

## 🧭 Rollup
- Focus days: {{summary.focusDays}}
- Check-ins: {{summary.checkInDays}}
- Worked: {{summary.workedTime}}
- Rest: {{summary.restTime}}
- Completed issues: {{summary.completedIssues}}
- Abandoned issues: {{summary.abandonedIssues}}
{{#if summary.averageMood}}- Avg mood: {{summary.averageMood}}{{/if}}
{{#if summary.averageEnergy}}- Avg energy: {{summary.averageEnergy}}{{/if}}

## 🔥 Streaks
- Focus: {{streaks.currentFocusDays}} / {{streaks.longestFocusDays}}
- Check-ins: {{streaks.currentCheckInDays}} / {{streaks.longestCheckInDays}}

## 📌 Daily Cards
{{#each days}}
### {{date}}
- Worked: {{workedTime}}
- Sessions: {{sessionCount}}
- Issues: {{completedIssues}} ✅ / {{abandonedIssues}} 🚧 / {{totalIssues}} total
{{#if checkIn}}- Mood / energy: {{checkIn.mood}} / {{checkIn.energy}}{{/if}}
{{/each}}
`

const weeklyDeepMarkdownTemplate = `# 📝 Weekly Review

Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}

## Rollup
- Days covered: {{summary.days}}
- Focus days: {{summary.focusDays}}
- Check-in days: {{summary.checkInDays}}
- Worked: {{summary.workedTime}}
- Rest: {{summary.restTime}}
- Completed issues: {{summary.completedIssues}}
- Abandoned issues: {{summary.abandonedIssues}}
- Estimated time: {{summary.estimatedTime}}
{{#if summary.averageMood}}- Average mood: {{summary.averageMood}}{{/if}}
{{#if summary.averageEnergy}}- Average energy: {{summary.averageEnergy}}{{/if}}

## Streaks
- Focus: {{streaks.currentFocusDays}} current / {{streaks.longestFocusDays}} longest
- Check-ins: {{streaks.currentCheckInDays}} current / {{streaks.longestCheckInDays}} longest

## Daily Breakdown
{{#each days}}
### {{date}}
- Worked: {{workedTime}}
- Sessions: {{sessionCount}}
- Issues: {{totalIssues}} total / {{completedIssues}} completed / {{abandonedIssues}} abandoned
{{#if checkIn}}- Check-in: mood {{checkIn.mood}} / energy {{checkIn.energy}}{{/if}}
{{/each}}
`

const weeklyBriefPDFTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Weekly Snapshot</title><link rel="stylesheet" href="report.css"></head>
<body class="report weekly brief">
<header class="hero"><div class="eyebrow">📅 Weekly snapshot</div><h1>{{startDate}} → {{endDate}}</h1></header>
<section class="metric-grid compact"><article class="metric-card"><div class="metric-label">Focus days</div><div class="metric-value">{{summary.focusDays}}</div></article><article class="metric-card"><div class="metric-label">Worked</div><div class="metric-value">{{summary.workedTime}}</div></article></section>
<section class="section"><h2>Days</h2>{{#each days}}<div class="row"><span>{{date}}</span><span>{{workedTime}} · {{sessionCount}} sessions</span></div>{{/each}}</section>
</body></html>`

const weeklyVisualPDFTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Weekly Pulse</title><link rel="stylesheet" href="report.css"></head>
<body class="report weekly visual">
<header class="hero accent"><div class="eyebrow">✨ Weekly pulse</div><h1>{{startDate}} → {{endDate}}</h1></header>
<section class="metric-grid"><article class="metric-card emphasis"><div class="metric-label">Focus days</div><div class="metric-value">{{summary.focusDays}}</div></article><article class="metric-card emphasis"><div class="metric-label">Worked</div><div class="metric-value">{{summary.workedTime}}</div></article><article class="metric-card"><div class="metric-label">Check-ins</div><div class="metric-value">{{summary.checkInDays}}</div></article><article class="metric-card"><div class="metric-label">Rest</div><div class="metric-value">{{summary.restTime}}</div></article></section>
<section class="section"><h2>🔥 Streaks</h2><div class="row"><span>Focus</span><span>{{streaks.currentFocusDays}} / {{streaks.longestFocusDays}}</span></div><div class="row"><span>Check-ins</span><span>{{streaks.currentCheckInDays}} / {{streaks.longestCheckInDays}}</span></div></section>
</body></html>`

const weeklyDeepPDFTemplate = `<!doctype html>
<html><head><meta charset="utf-8"><title>Weekly Review</title><link rel="stylesheet" href="report.css"></head>
<body class="report weekly deep">
<header class="hero"><div class="eyebrow">📝 Weekly review</div><h1>{{startDate}} → {{endDate}}</h1><div class="subtle">Generated at {{generatedAt}}</div></header>
<section class="metric-grid"><article class="metric-card"><div class="metric-label">Worked</div><div class="metric-value">{{summary.workedTime}}</div></article><article class="metric-card"><div class="metric-label">Completed issues</div><div class="metric-value">{{summary.completedIssues}}</div></article></section>
<section class="section"><h2>Days</h2>{{#each days}}<article class="day-card"><div class="day-header"><h3>{{date}}</h3><div class="pill">{{workedTime}}</div></div><div class="day-grid"><div>Sessions: {{sessionCount}}</div><div>Issues: {{completedIssues}} completed / {{totalIssues}} total</div>{{#if checkIn}}<div>Check-in: mood {{checkIn.mood}} / energy {{checkIn.energy}}</div>{{/if}}</div></article>{{/each}}</section>
</body></html>`

const dailyBriefPDFStyles = `@page { margin: 18mm; } body { font-family: Inter, Arial, sans-serif; color: #142018; font-size: 11pt; } .hero { margin-bottom: 16px; } .eyebrow { color: #1d7a57; text-transform: uppercase; letter-spacing: .12em; font-size: 9pt; } h1 { margin: 4px 0 0; font-size: 24pt; } .metric-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin: 16px 0; } .metric-card { border: 1px solid #dbe9df; border-radius: 10px; padding: 10px 12px; background: #f8fbf8; } .metric-label { font-size: 8.5pt; color: #5c6b61; text-transform: uppercase; } .metric-value { font-size: 14pt; font-weight: 700; margin-top: 4px; } .section { margin-top: 16px; } h2 { font-size: 14pt; margin: 0 0 8px; } .bullet-list { padding-left: 18px; margin: 0; } .row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #edf3ee; }`
const dailyVisualPDFStyles = `@page { margin: 16mm; } body { font-family: "Avenir Next", Inter, sans-serif; color: #101418; font-size: 11pt; } .hero { margin-bottom: 16px; padding: 14px 16px; border-radius: 14px; background: linear-gradient(135deg, #effaf5, #f7fbff); } .eyebrow { color: #0f7a60; text-transform: uppercase; letter-spacing: .14em; font-size: 9pt; } h1 { margin: 4px 0 0; font-size: 24pt; } .subtle { color: #53616c; margin-top: 4px; } .metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 16px 0; } .metric-card { border-radius: 12px; padding: 12px 14px; background: #f6faf8; border: 1px solid #d9ebe1; } .metric-card.emphasis { background: #0f7a60; color: white; border-color: #0f7a60; } .metric-card.emphasis .metric-label { color: rgba(255,255,255,.78); } .metric-label { font-size: 8.5pt; text-transform: uppercase; letter-spacing: .08em; color: #5a6660; } .metric-value { font-size: 15pt; font-weight: 700; margin-top: 4px; } .section { margin-top: 18px; } h2 { font-size: 14pt; margin: 0 0 8px; } .bullet-list { padding-left: 18px; margin: 0; }`
const dailyDeepPDFStyles = `@page { margin: 18mm; } body { font-family: Georgia, "Times New Roman", serif; color: #1d1d1d; font-size: 11pt; line-height: 1.45; } .hero { margin-bottom: 18px; } .eyebrow { color: #5b6b63; text-transform: uppercase; letter-spacing: .12em; font-size: 9pt; } h1 { margin: 4px 0 0; font-size: 23pt; } .subtle { color: #69736f; margin-top: 4px; } .metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 16px 0; } .metric-card { border: 1px solid #e4e4e4; padding: 10px 12px; border-radius: 10px; } .metric-label { color: #626262; font-size: 8.5pt; text-transform: uppercase; } .metric-value { margin-top: 4px; font-size: 14pt; font-weight: 700; } .section { margin-top: 18px; } h2 { font-size: 14pt; border-bottom: 1px solid #ddd; padding-bottom: 4px; } h3 { margin: 12px 0 8px; } h4 { margin: 10px 0 6px; color: #3d4c45; } .bullet-list { padding-left: 18px; } .row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #efefef; } .row-good { color: #1b6d45; } .row-warn { color: #8a5b00; }`
const weeklyBriefPDFStyles = `@page { margin: 18mm; } body { font-family: Inter, Arial, sans-serif; color: #162018; font-size: 11pt; } .hero { margin-bottom: 16px; } .eyebrow { color: #2366c2; text-transform: uppercase; letter-spacing: .12em; font-size: 9pt; } h1 { margin: 4px 0 0; font-size: 22pt; } .metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 16px 0; } .metric-card { border: 1px solid #dde7f5; border-radius: 10px; padding: 10px 12px; background: #f8fbff; } .metric-label { font-size: 8.5pt; color: #5f6871; text-transform: uppercase; } .metric-value { font-size: 14pt; font-weight: 700; margin-top: 4px; } .section { margin-top: 16px; } .row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #edf1f7; }`
const weeklyVisualPDFStyles = `@page { margin: 16mm; } body { font-family: "Avenir Next", Inter, sans-serif; color: #122030; font-size: 11pt; } .hero { margin-bottom: 16px; padding: 14px 16px; border-radius: 14px; background: linear-gradient(135deg, #f3f8ff, #f9fbff); } .eyebrow { color: #275ec0; text-transform: uppercase; letter-spacing: .14em; font-size: 9pt; } h1 { margin: 4px 0 0; font-size: 23pt; } .metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 16px 0; } .metric-card { border-radius: 12px; padding: 12px 14px; background: #f7faff; border: 1px solid #dfe7f6; } .metric-card.emphasis { background: #275ec0; color: white; border-color: #275ec0; } .metric-card.emphasis .metric-label { color: rgba(255,255,255,.75); } .metric-label { font-size: 8.5pt; color: #5c6672; text-transform: uppercase; } .metric-value { font-size: 15pt; font-weight: 700; margin-top: 4px; } .section { margin-top: 18px; } .row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #edf1f7; }`
const weeklyDeepPDFStyles = `@page { margin: 18mm; } body { font-family: Georgia, "Times New Roman", serif; color: #202020; font-size: 11pt; line-height: 1.45; } .hero { margin-bottom: 18px; } .eyebrow { color: #4f5f80; text-transform: uppercase; letter-spacing: .12em; font-size: 9pt; } h1 { margin: 4px 0 0; font-size: 22pt; } .subtle { color: #6c7380; margin-top: 4px; } .metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 16px 0; } .metric-card { border: 1px solid #e4e8ef; padding: 10px 12px; border-radius: 10px; } .metric-label { color: #666; font-size: 8.5pt; text-transform: uppercase; } .metric-value { margin-top: 4px; font-size: 14pt; font-weight: 700; } .section { margin-top: 18px; } .day-card { border: 1px solid #eceff4; border-radius: 10px; padding: 10px 12px; margin-bottom: 10px; break-inside: avoid; } .day-header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 6px; } h3 { margin: 0; font-size: 12pt; } .pill { background: #eef3fb; color: #345ea8; border-radius: 999px; padding: 3px 8px; font-size: 9pt; font-weight: 700; } .day-grid { display: grid; gap: 4px; }`
