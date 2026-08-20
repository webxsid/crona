---
hosted: true
title: Exports and Reports
description: Generate summary dashboards, reports, CSV exports, and iCalendar files from your local workspace.
order: 5.7
---

When you want to keep a review, share a handoff, or move work into a calendar, export it from Crona. Files are rendered on your machine from your local data.

## Pick an output

| Format | Output Type | Primary Use Case |
| --- | --- | --- |
| **Markdown** | Text files | Note taking, personal archives, Git-controlled documentation. |
| **PDF** | Formatted documents | Printable dashboards, handoffs, and weekly reviews. |
| **CSV** | Tabular data | Raw time sheets, spreadsheet importing, custom data analysis. |
| **iCalendar (.ics)** | Calendar feeds | Syncing time blocks and issues to local calendar clients. |
| **Clipboard** | Markdown text | Quick paste into notes, chat, or issue trackers without writing a file. |

## Pick a report

Crona offers these report types:
- **Summary**: A glance-style single-day dashboard for context, daily signals, habits, and streaks.
- **Summary Range**: An inclusive date-range dashboard covering a week, month, rolling window, or custom range.
- **Daily**: Summary of completed issues, habit checks, check-in scores, and focus durations.
- **Weekly**: Rollup of daily metrics, habit compliance over time, and focus breakdowns.
- **Repository**: Status, notes, and session logs scoped to a top-level workspace.
- **Stream**: Metrics and issue summaries restricted to a specific repository subdivision.
- **Issue Rollup**: Aggregated focus durations, status histories, and commit notes for a single issue.

The TUI separates at-a-glance Summary exports from narrative reports. `crona export summary` produces a saved or clipboard-ready artifact; `crona summary` is the separate read-only command for inspecting a day or range directly in the terminal.

## PDF note

Markdown and CSV exports work out-of-the-box. PDF generation requires local renderer dependencies:
- **Summary, Daily, and Weekly PDFs**: Rendered via `weasyprint`.
- **Repository, Stream, and Issue Rollup PDFs**: Processed via `pandoc` combined with one of these PDF engines:
  - `tectonic`
  - `weasyprint`
  - `wkhtmltopdf`
  - `xelatex`
  - `pdflatex`

Tool discovery is handled dynamically by the daemon and can be verified in the TUI **Config** view.

## Make exports yours

Report assets are user-editable. In the **Config** view, you can access and open:
- **Handlebars (.hbs) Templates**: Used to structure Markdown and HTML outputs.
- **CSV Specs**: JSON files defining columns, date formatting, and database queries.
- **Export Paths**: Configurations to change the target directory for report and ICS output.

## A typical export

- Summary dashboards can target a selected day, calendar week, calendar month, last 7 days, last 30 days, or a custom inclusive range.
- Markdown outputs can be written to the reports directory or copied directly to the clipboard.
- Repo, stream, and issue-rollup exports prompt for their target, rather than requiring you to change the active context first.
- Generated report files are listed in the TUI Reports view and can emit completion alerts when local notifications are enabled.

Next: [Calendar and File Automation](calendar-and-file-automation.md) if another app should consume the files automatically.
