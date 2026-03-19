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
{{/each}}
{{/if}}
{{/each}}
{{/each}}
`

const fallbackDailyReportPDFTemplate = `# Daily Report - {{date}}
Generated at {{generatedAt}}
## Snapshot
- Day health: {{dayHealth}}
- Issues: {{summary.issueDoneCount}} done / {{summary.totalIssues}} total
- Habits: {{summary.habitsCompletedCount}} completed / {{summary.habitsDueCount}} due
- Worked / estimated: {{summary.workedEstimate}}
- Variance: {{summary.varianceTime}}
{{#if checkIn}}- Mood / energy: {{checkIn.mood}} / 5, {{checkIn.energy}} / 5{{/if}}
{{#if checkIn.sleepHours}}- Sleep: {{checkIn.sleepHours}}h{{/if}}
{{#if metrics.burnout}}- Burnout: {{metrics.burnout.level}} ({{metrics.burnout.score}}){{/if}}
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
## Habits
{{#each habitRepos}}
### {{name}}
{{#each streams}}
#### {{name}}
{{#each completedHabits}}
- {{name}} | {{status}}{{#if durationTime}} | {{durationTime}}{{/if}}
{{/each}}
{{#each pendingHabits}}
- {{name}} | {{status}}{{#if durationTime}} | {{durationTime}}{{/if}}
{{/each}}
{{/each}}
{{/each}}
## Issues
{{#each repos}}
### {{name}}
{{#each streams}}
#### {{name}}
{{#each completedIssues}}
- #{{id}} {{title}} | {{status}} | {{workedEstimate}}
{{/each}}
{{#each activeIssues}}
- #{{id}} {{title}} | {{status}} | {{workedEstimate}}
{{/each}}
{{#each attentionIssues}}
- #{{id}} {{title}} | {{status}} | {{workedEstimate}}
{{/each}}
{{/each}}
{{/each}}
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

const fallbackWeeklyReportPDFTemplate = `# Weekly Summary
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
## Days
{{#each days}}
- {{date}} | {{workedTime}} | {{sessionCount}} sessions
{{/each}}
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
