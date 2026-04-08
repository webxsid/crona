package export

const fallbackDailyReportTemplate = `# Daily Report - {{date}}

Generated at {{generatedAt}}

## Daily Snapshot

- Day health: {{dayHealth}}
- Issues: {{summary.totalIssues}} total, {{summary.issueDoneCount}} done, {{summary.issueActiveCount}} active, {{summary.issueBlockedCount}} blocked, {{summary.issueAbandonedCount}} abandoned
- Issue completion: {{summary.issueCompletion}}
- Worked / estimated: {{summary.workedEstimate}}
- Variance: {{summary.varianceTime}}
- Habits: {{summary.habitsCompletedCount}} completed / {{summary.habitsDueCount}} due
{{#if checkIn}}- Mood / energy: {{checkIn.mood}} / 5, {{checkIn.energy}} / 5{{/if}}
{{#if checkIn.sleepHours}}- Sleep: {{checkIn.sleepHours}}h{{/if}}
{{#if checkIn.screenTime}}- Screen time: {{checkIn.screenTime}}{{/if}}
{{#if metrics.burnout}}- Burnout: {{metrics.burnout.level}} ({{metrics.burnout.score}}){{/if}}

{{#if plan}}
## Plan Accountability

- Planned: {{summary.plannedCount}}
- Completed: {{summary.planCompletedCount}}
- Failed: {{summary.planFailedCount}}
- Abandoned: {{summary.planAbandonedCount}}
- Pending rollback: {{summary.planPendingRollbackCount}}
- Accountability score: {{summary.accountabilityScore}}
- Backlog pressure: {{summary.backlogPressure}}
- Delayed issues: {{summary.delayedIssueCount}}
- High-risk issues: {{summary.highRiskIssueCount}}
{{#if summary.avgDelayDays}}- Avg delay days: {{summary.avgDelayDays}}{{/if}}
{{#if summary.maxDelayDays}}- Max delay days: {{summary.maxDelayDays}}{{/if}}
{{/if}}

{{#if highlights}}
## Highlights

{{#each highlights}}
- {{this}}
{{/each}}
{{/if}}

{{#if risks}}
## Needs Attention

{{#each risks}}
- {{this}}
{{/each}}
{{/if}}

{{#if checkIn}}
## Wellbeing

- Mood: {{checkIn.mood}} / 5
- Energy: {{checkIn.energy}} / 5
{{#if checkIn.sleepHours}}- Sleep hours: {{checkIn.sleepHours}}{{/if}}
{{#if checkIn.sleepScore}}- Sleep score: {{checkIn.sleepScore}}{{/if}}
{{#if checkIn.screenTime}}- Screen time: {{checkIn.screenTime}}{{/if}}
{{#if metrics}}- Rest time: {{metrics.restTime}}{{/if}}
{{#if checkIn.notes}}

{{checkIn.notes}}
{{/if}}
{{/if}}

## Habits

{{#each habitRepos}}
### {{name}}

{{#each streams}}
#### {{name}}

{{#if completedHabits}}
##### Completed

{{#each completedHabits}}
###### {{name}}

- Status: {{status}}
{{#if durationTime}}- Time: {{durationTime}}{{/if}}
{{/each}}
{{/if}}

{{#if pendingHabits}}
##### Pending

{{#each pendingHabits}}
###### {{name}}

- Status: {{status}}
{{#if durationTime}}- Time: {{durationTime}}{{/if}}
{{/each}}
{{/if}}
{{/each}}
{{/each}}

## Issues

{{#each repos}}
### {{name}}

{{#each streams}}
#### {{name}}

{{#if completedIssues}}
##### Completed

{{#each completedIssues}}
###### #{{id}} {{title}}

- Status: {{status}}
- Worked / estimate: {{workedEstimate}}
{{/each}}
{{/if}}

{{#if activeIssues}}
##### Active

{{#each activeIssues}}
###### #{{id}} {{title}}

- Status: {{status}}
- Worked / estimate: {{workedEstimate}}
{{/each}}
{{/if}}

{{#if attentionIssues}}
##### Needs Attention

{{#each attentionIssues}}
###### #{{id}} {{title}}

- Status: {{status}}
- Worked / estimate: {{workedEstimate}}
{{#if planStatus}}- Plan status: {{planStatus}}{{/if}}
{{#if planFailureReason}}- Failure reason: {{planFailureReason}}{{/if}}
{{#if planCurrentDelayedDays}}- Delayed by: {{planCurrentDelayedDays}} day(s){{/if}}
{{/each}}
{{/if}}
{{/each}}
{{/each}}
`

const fallbackDailyReportPDFTemplate = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Daily Report - {{date}}</title>
  <link rel="stylesheet" href="report.css">
</head>
<body class="report report-daily theme-balanced">
  <header class="hero">
    <div class="eyebrow">Daily report</div>
    <h1>{{date}}</h1>
    <div class="subtle">Generated at {{generatedAt}}</div>
  </header>

  <section class="metric-grid">
    <article class="metric-card">
      <div class="metric-label">Day health</div>
      <div class="metric-value">{{dayHealth}}</div>
    </article>
    <article class="metric-card">
      <div class="metric-label">Issues</div>
      <div class="metric-value">{{summary.issueDoneCount}} / {{summary.totalIssues}}</div>
    </article>
    <article class="metric-card">
      <div class="metric-label">Habits</div>
      <div class="metric-value">{{summary.habitsCompletedCount}} / {{summary.habitsDueCount}}</div>
    </article>
    <article class="metric-card">
      <div class="metric-label">Worked / estimate</div>
      <div class="metric-value">{{summary.workedEstimate}}</div>
    </article>
    {{#if plan}}
    <article class="metric-card">
      <div class="metric-label">Accountability</div>
      <div class="metric-value">{{summary.accountabilityScore}}</div>
    </article>
    <article class="metric-card">
      <div class="metric-label">Failed / delayed</div>
      <div class="metric-value">{{summary.planFailedCount}} / {{summary.delayedIssueCount}}</div>
    </article>
    {{/if}}
  </section>

  {{#if highlights}}
  <section class="section">
    <h2>✨ Highlights</h2>
    <ul class="bullet-list">
      {{#each highlights}}<li>{{this}}</li>{{/each}}
    </ul>
  </section>
  {{/if}}

  {{#if risks}}
  <section class="section">
    <h2>🚧 Needs Attention</h2>
    <ul class="bullet-list">
      {{#each risks}}<li>{{this}}</li>{{/each}}
    </ul>
  </section>
  {{/if}}

  {{#if plan}}
  <section class="section">
    <h2>Plan Accountability</h2>
    <div class="row"><span>Planned</span><span>{{summary.plannedCount}}</span></div>
    <div class="row"><span>Completed</span><span>{{summary.planCompletedCount}}</span></div>
    <div class="row"><span>Failed</span><span>{{summary.planFailedCount}}</span></div>
    <div class="row"><span>Pending rollback</span><span>{{summary.planPendingRollbackCount}}</span></div>
    <div class="row"><span>Delayed issues</span><span>{{summary.delayedIssueCount}}</span></div>
    <div class="row"><span>High-risk issues</span><span>{{summary.highRiskIssueCount}}</span></div>
    <div class="row"><span>Backlog pressure</span><span>{{summary.backlogPressure}}</span></div>
  </section>
  {{/if}}

  <section class="section">
    <h2>Issues</h2>
    {{#each repos}}
    <article class="group">
      <h3>{{name}}</h3>
      {{#each streams}}
      <div class="subgroup">
        <h4>{{name}}</h4>
        {{#each completedIssues}}<div class="row row-good"><span>#{{id}} {{title}}</span><span>{{workedEstimate}}</span></div>{{/each}}
        {{#each activeIssues}}<div class="row"><span>#{{id}} {{title}}</span><span>{{workedEstimate}}</span></div>{{/each}}
        {{#each attentionIssues}}<div class="row row-warn"><span>#{{id}} {{title}}</span><span>{{workedEstimate}}</span></div>{{/each}}
      </div>
      {{/each}}
    </article>
    {{/each}}
  </section>
</body>
</html>
`

const fallbackDailyReportPDFStyles = `@page { margin: 18mm; }
body { font-family: Inter, "Helvetica Neue", Arial, sans-serif; color: #102017; font-size: 11pt; line-height: 1.45; }
.hero { margin-bottom: 18px; }
.eyebrow { text-transform: uppercase; letter-spacing: 0.14em; color: #1b8f6a; font-size: 9pt; margin-bottom: 6px; }
h1 { margin: 0; font-size: 24pt; }
.subtle { color: #5b6b63; margin-top: 4px; }
.metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 18px 0; }
.metric-card { border: 1px solid #cfe4d6; border-radius: 10px; padding: 10px 12px; background: #f8fcf9; }
.metric-label { color: #5b6b63; font-size: 9pt; text-transform: uppercase; letter-spacing: 0.08em; }
.metric-value { margin-top: 4px; font-size: 14pt; font-weight: 700; }
.section { margin-top: 18px; }
h2 { font-size: 14pt; border-bottom: 2px solid #dcebe1; padding-bottom: 4px; margin-bottom: 10px; }
h3 { margin: 14px 0 8px; font-size: 12pt; }
h4 { margin: 10px 0 6px; color: #2c5a48; font-size: 10.5pt; }
.bullet-list { margin: 0; padding-left: 18px; }
.group, .subgroup { break-inside: avoid; }
.row { display: flex; justify-content: space-between; gap: 12px; padding: 6px 0; border-bottom: 1px solid #eef5f0; }
.row-good { color: #1d6d47; }
.row-warn { color: #8c4f00; }
`

const fallbackDailyReportVariables = `# Daily Report Template Variables

The daily report template uses Handlebars-style variables.

The same variable set is available to both the markdown and PDF templates.

Supported blocks in v1:
- {{variable}}
- {{#if variable}}...{{/if}}
- {{#each items}}...{{/each}}

Top-level variables:
- {{date}}
- {{generatedAt}}
- {{summary.totalIssues}}
- {{summary.issueDoneCount}}
- {{summary.issueActiveCount}}
- {{summary.issueBlockedCount}}
- {{summary.issueAbandonedCount}}
- {{summary.issueCompletion}}
- {{summary.completedIssues}}
- {{summary.abandonedIssues}}
- {{summary.totalEstimatedMinutes}}
- {{summary.estimatedTime}}
- {{summary.workedSeconds}}
- {{summary.workedTime}}
- {{summary.workedEstimate}}
- {{summary.varianceTime}}
- {{summary.habitsDueCount}}
- {{summary.habitsCompletedCount}}
- {{summary.habitsPendingCount}}
- {{summary.habitCompletion}}
- {{summary.plannedCount}}
- {{summary.planCompletedCount}}
- {{summary.planFailedCount}}
- {{summary.planAbandonedCount}}
- {{summary.planPendingRollbackCount}}
- {{summary.accountabilityScore}}
- {{summary.backlogPressure}}
- {{summary.delayedIssueCount}}
- {{summary.highRiskIssueCount}}
- {{summary.avgDelayDays}}
- {{summary.maxDelayDays}}
- {{dayHealth}}
- {{#each highlights}} ... {{/each}}
- {{#each risks}} ... {{/each}}

Nested issue groups for the default template:
- {{#each repos}} ... {{/each}}
  - {{name}}
  - {{#each streams}} ... {{/each}}
    - {{name}}
    - {{#each completedIssues}} ... {{/each}}
    - {{#each activeIssues}} ... {{/each}}
    - {{#each attentionIssues}} ... {{/each}}
      - {{id}}
      - {{title}}
      - {{status}}
      - {{estimateMinutes}}
      - {{estimateTime}}
      - {{workedSeconds}}
      - {{workedTime}}
      - {{workedEstimate}}
      - {{planStatus}}
      - {{planFailureReason}}
      - {{planPendingFailureReason}}
      - {{planCurrentDelayedDays}}
      - {{planMaxDelayedDays}}
      - {{planFailScore}}

Nested habit groups for the default template:
- {{#each habitRepos}} ... {{/each}}
  - {{name}}
  - {{#each streams}} ... {{/each}}
    - {{name}}
    - {{#each completedHabits}} ... {{/each}}
    - {{#each pendingHabits}} ... {{/each}}
      - {{id}}
      - {{name}}
      - {{status}}
      - {{durationMinutes}}
      - {{durationTime}}
      - {{notes}}

Optional objects:
- {{checkIn.mood}}
- {{checkIn.energy}}
- {{checkIn.sleepHours}}
- {{checkIn.sleepScore}}
- {{checkIn.screenTimeMinutes}}
- {{checkIn.screenTime}}
- {{checkIn.notes}}

Collections:
- {{#each issues}} ... {{/each}}
  - {{id}}
  - {{title}}
  - {{repoName}}
  - {{streamName}}
  - {{status}}
  - {{estimateMinutes}}
  - {{estimateTime}}
  - {{workedSeconds}}
  - {{workedTime}}
  - {{workedEstimate}}
- {{#each sessions}} ... {{/each}}
  - {{id}}
  - {{issueId}}
  - {{issueTitle}}
  - {{repoName}}
  - {{streamName}}
  - {{startTime}}
  - {{endTime}}
  - {{durationSeconds}}
  - {{summary}}
- {{#each habits}} ... {{/each}}
  - {{id}}
  - {{name}}
  - {{repoName}}
  - {{streamName}}
  - {{status}}
  - {{durationMinutes}}
  - {{durationTime}}
  - {{notes}}

Metrics:
- {{metrics.sessionCount}}
- {{metrics.workedSeconds}}
- {{metrics.workedTime}}
- {{metrics.restSeconds}}
- {{metrics.restTime}}
- {{metrics.burnout.level}}
- {{metrics.burnout.score}}
`

const fallbackWeeklyReportTemplate = `# Weekly Summary

Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}

## Rollup

- Days: {{summary.days}}
- Check-ins: {{summary.checkInDays}}
- Focus days: {{summary.focusDays}}
- Worked: {{summary.workedTime}}
- Rest: {{summary.restTime}}
- Completed issues: {{summary.completedIssues}}
- Abandoned issues: {{summary.abandonedIssues}}
- Estimated time: {{summary.estimatedTime}}
{{#if summary.averageMood}}- Avg mood: {{summary.averageMood}}{{/if}}
{{#if summary.averageEnergy}}- Avg energy: {{summary.averageEnergy}}{{/if}}

## Streaks

- Focus: {{streaks.currentFocusDays}} current / {{streaks.longestFocusDays}} longest
- Check-ins: {{streaks.currentCheckInDays}} current / {{streaks.longestCheckInDays}} longest

## Days

{{#each days}}
### {{date}}

- Worked: {{workedTime}}
- Sessions: {{sessionCount}}
- Issues: {{totalIssues}} total / {{completedIssues}} completed / {{abandonedIssues}} abandoned
{{#if checkIn}}- Check-in: mood {{checkIn.mood}} / energy {{checkIn.energy}}{{/if}}
{{/each}}
`

const fallbackWeeklyReportPDFTemplate = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Weekly Summary - {{startDate}} to {{endDate}}</title>
  <link rel="stylesheet" href="report.css">
</head>
<body class="report report-weekly theme-balanced">
  <header class="hero">
    <div class="eyebrow">Weekly summary</div>
    <h1>{{startDate}} → {{endDate}}</h1>
    <div class="subtle">Generated at {{generatedAt}}</div>
  </header>

  <section class="metric-grid">
    <article class="metric-card"><div class="metric-label">Focus days</div><div class="metric-value">{{summary.focusDays}}</div></article>
    <article class="metric-card"><div class="metric-label">Check-ins</div><div class="metric-value">{{summary.checkInDays}}</div></article>
    <article class="metric-card"><div class="metric-label">Worked</div><div class="metric-value">{{summary.workedTime}}</div></article>
    <article class="metric-card"><div class="metric-label">Rest</div><div class="metric-value">{{summary.restTime}}</div></article>
  </section>

  <section class="section">
    <h2>🔥 Streaks</h2>
    <div class="row"><span>Focus</span><span>{{streaks.currentFocusDays}} current / {{streaks.longestFocusDays}} best</span></div>
    <div class="row"><span>Check-ins</span><span>{{streaks.currentCheckInDays}} current / {{streaks.longestCheckInDays}} best</span></div>
  </section>

  <section class="section">
    <h2>Days</h2>
    {{#each days}}
    <article class="day-card">
      <div class="day-header">
        <h3>{{date}}</h3>
        <div class="pill">{{workedTime}}</div>
      </div>
      <div class="day-grid">
        <div>Sessions: {{sessionCount}}</div>
        <div>Issues: {{completedIssues}} completed / {{totalIssues}} total</div>
        {{#if checkIn}}<div>Check-in: mood {{checkIn.mood}} / energy {{checkIn.energy}}</div>{{/if}}
      </div>
    </article>
    {{/each}}
  </section>
</body>
</html>
`

const fallbackWeeklyReportPDFStyles = `@page { margin: 18mm; }
body { font-family: Inter, "Helvetica Neue", Arial, sans-serif; color: #102017; font-size: 11pt; line-height: 1.45; }
.hero { margin-bottom: 18px; }
.eyebrow { text-transform: uppercase; letter-spacing: 0.14em; color: #2266c6; font-size: 9pt; margin-bottom: 6px; }
h1 { margin: 0; font-size: 22pt; }
.subtle { color: #5b6b63; margin-top: 4px; }
.metric-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; margin: 18px 0; }
.metric-card { border: 1px solid #d8e5f7; border-radius: 10px; padding: 10px 12px; background: #f8fbff; }
.metric-label { color: #5b6b63; font-size: 9pt; text-transform: uppercase; letter-spacing: 0.08em; }
.metric-value { margin-top: 4px; font-size: 14pt; font-weight: 700; }
.section { margin-top: 18px; }
h2 { font-size: 14pt; border-bottom: 2px solid #dfe8f5; padding-bottom: 4px; margin-bottom: 10px; }
.row { display: flex; justify-content: space-between; gap: 12px; padding: 6px 0; border-bottom: 1px solid #eef2f7; }
.day-card { border: 1px solid #e6edf8; border-radius: 10px; padding: 10px 12px; margin-bottom: 10px; break-inside: avoid; }
.day-header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 6px; }
h3 { margin: 0; font-size: 12pt; }
.pill { background: #ecf4ff; color: #1d4fa8; border-radius: 999px; padding: 3px 8px; font-size: 9pt; font-weight: 700; }
.day-grid { display: grid; gap: 4px; }
`

const fallbackWeeklyReportVariables = `# Weekly Report Template Variables

Top-level:
- {{startDate}}
- {{endDate}}
- {{generatedAt}}

Summary:
- {{summary.days}}
- {{summary.checkInDays}}
- {{summary.focusDays}}
- {{summary.workedTime}}
- {{summary.restTime}}
- {{summary.completedIssues}}
- {{summary.abandonedIssues}}
- {{summary.estimatedTime}}
- {{summary.averageMood}}
- {{summary.averageEnergy}}

Streaks:
- {{streaks.currentFocusDays}}
- {{streaks.longestFocusDays}}
- {{streaks.currentCheckInDays}}
- {{streaks.longestCheckInDays}}

Days:
- {{#each days}}
  - {{date}}
  - {{workedTime}}
  - {{sessionCount}}
  - {{totalIssues}}
  - {{completedIssues}}
  - {{abandonedIssues}}
  - {{checkIn.mood}}
  - {{checkIn.energy}}
`

const fallbackRepoReportTemplate = `# Repo Report - {{repo.name}}

Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}

{{#if repo.description}}
## Description

{{repo.description}}
{{/if}}

## Snapshot

- Streams: {{summary.streamCount}}
- Issues: {{summary.issueCount}}
- Habits: {{summary.habitCount}}
- Sessions: {{summary.sessionCount}}

## Streams

{{#each streams}}
- {{name}}{{#if description}}: {{description}}{{/if}}
{{/each}}

## Issues

{{#each issues}}
### #{{id}} {{title}}

- Status: {{status}}
- Scope: {{scope}}
- Estimate: {{estimateTime}}
{{#if description}}
Description
{{description}}
{{/if}}
{{#if notes}}
Issue Notes
{{notes}}
{{/if}}

Sessions
{{#each sessions}}
- {{summary}}
  Commit: {{commit}}
  Context: {{context}}
  Work: {{work}}
  Notes: {{notes}}
{{/each}}
{{/each}}

## Habits

{{#each habits}}
- {{name}} | {{scheduleType}}
{{/each}}
`

const fallbackRepoReportPDFTemplate = `# Repo Report - {{repo.name}}
Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}
## Snapshot
- Streams: {{summary.streamCount}}
- Issues: {{summary.issueCount}}
- Habits: {{summary.habitCount}}
- Sessions: {{summary.sessionCount}}
## Issues
{{#each issues}}
- #{{id}} {{title}} | {{status}} | {{estimateTime}}
{{/each}}
`

const fallbackRepoReportVariables = `# Repo Report Template Variables

Top-level:
- {{startDate}}
- {{endDate}}
- {{generatedAt}}
- {{repo.name}}
- {{repo.description}}

Summary:
- {{summary.streamCount}}
- {{summary.issueCount}}
- {{summary.habitCount}}
- {{summary.sessionCount}}

Streams:
- {{#each streams}}
  - {{name}}
  - {{description}}

Issues:
- {{#each issues}}
  - {{id}}
  - {{title}}
  - {{status}}
  - {{scope}}
  - {{estimateTime}}
  - {{description}}
  - {{notes}}
  - {{#each sessions}}
    - {{summary}}
    - {{commit}}
    - {{context}}
    - {{work}}
    - {{notes}}

Habits:
- {{#each habits}}
  - {{name}}
  - {{scheduleType}}
`

const fallbackStreamReportTemplate = `# Stream Report - {{stream.name}}

Repo: {{repo.name}}
Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}

{{#if stream.description}}
## Description

{{stream.description}}
{{/if}}

## Snapshot

- Issues: {{summary.issueCount}}
- Habits: {{summary.habitCount}}
- Sessions: {{summary.sessionCount}}

## Issues

{{#each issues}}
### #{{id}} {{title}}

- Status: {{status}}
- Estimate: {{estimateTime}}
{{#if description}}
Description
{{description}}
{{/if}}
{{#if notes}}
Issue Notes
{{notes}}
{{/if}}

Sessions
{{#each sessions}}
- {{summary}}
  Commit: {{commit}}
  Context: {{context}}
  Work: {{work}}
  Notes: {{notes}}
{{/each}}
{{/each}}

## Habits

{{#each habits}}
- {{name}} | {{scheduleType}}
{{/each}}
`

const fallbackStreamReportPDFTemplate = `# Stream Report - {{stream.name}}
Repo: {{repo.name}}
Range: {{startDate}} to {{endDate}}
## Snapshot
- Issues: {{summary.issueCount}}
- Habits: {{summary.habitCount}}
- Sessions: {{summary.sessionCount}}
## Issues
{{#each issues}}
- #{{id}} {{title}} | {{status}} | {{estimateTime}}
{{/each}}
`

const fallbackStreamReportVariables = `# Stream Report Template Variables

Top-level:
- {{startDate}}
- {{endDate}}
- {{generatedAt}}
- {{repo.name}}
- {{stream.name}}
- {{stream.description}}

Summary:
- {{summary.issueCount}}
- {{summary.habitCount}}
- {{summary.sessionCount}}

Issues:
- {{#each issues}}
  - {{id}}
  - {{title}}
  - {{status}}
  - {{estimateTime}}
  - {{description}}
  - {{notes}}
  - {{#each sessions}}
    - {{summary}}
    - {{commit}}
    - {{context}}
    - {{work}}
    - {{notes}}

Habits:
- {{#each habits}}
  - {{name}}
  - {{scheduleType}}
`

const fallbackIssueRollupReportTemplate = `# Session to Issue Rollup

Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}

## Rollup Table

{{#each issues}}
- #{{id}} {{title}} | {{status}} | {{sessionCount}} sessions | {{workedTime}} | {{estimateTime}} | {{scope}}
{{/each}}

## Detailed Issues

{{#each issues}}
### #{{id}} {{title}}

- Status: {{status}}
- Scope: {{scope}}
- Sessions: {{sessionCount}}
- Worked: {{workedTime}}
- Estimate: {{estimateTime}}
{{#if description}}
Description
{{description}}
{{/if}}
{{#if notes}}
Issue Notes
{{notes}}
{{/if}}

Sessions
{{#each sessions}}
- {{summary}}
  Commit: {{commit}}
  Context: {{context}}
  Work: {{work}}
  Notes: {{notes}}
{{/each}}
{{/each}}
`

const fallbackIssueRollupReportPDFTemplate = `# Session to Issue Rollup
Range: {{startDate}} to {{endDate}}
Generated at {{generatedAt}}
{{#each issues}}
- #{{id}} {{title}} | {{status}} | {{sessionCount}} sessions | {{workedTime}}
{{/each}}
`

const fallbackIssueRollupReportVariables = `# Issue Rollup Template Variables

Top-level:
- {{startDate}}
- {{endDate}}
- {{generatedAt}}

Issues:
- {{#each issues}}
  - {{id}}
  - {{title}}
  - {{status}}
  - {{scope}}
  - {{sessionCount}}
  - {{workedTime}}
  - {{estimateTime}}
  - {{description}}
  - {{notes}}
  - {{#each sessions}}
    - {{summary}}
    - {{commit}}
    - {{context}}
    - {{work}}
    - {{notes}}
`

const fallbackCSVExportSpec = `{
  "columns": [
    {"header": "Session ID", "field": "sessionId"},
    {"header": "Issue ID", "field": "issueId"},
    {"header": "Issue Title", "field": "issueTitle"},
    {"header": "Status", "field": "status"},
    {"header": "Repo", "field": "repo"},
    {"header": "Stream", "field": "stream"},
    {"header": "Start Time", "field": "startTime"},
    {"header": "End Time", "field": "endTime"},
    {"header": "Duration Seconds", "field": "durationSeconds"},
    {"header": "Commit", "field": "commit"},
    {"header": "Context", "field": "context"},
    {"header": "Work", "field": "work"},
    {"header": "Notes", "field": "notes"}
  ]
}
`

const fallbackCSVExportVariables = `# CSV Export Spec

The CSV export is configured by a JSON spec file with a single top-level key:
- columns

Each column entry supports:
- header: CSV header text
- field: field name from the session row dataset

Available row fields:
- sessionId
- issueId
- issueTitle
- status
- repo
- stream
- startTime
- endTime
- durationSeconds
- commit
- context
- work
- notes
`
