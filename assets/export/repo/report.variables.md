# Repo Report Template Variables

Top-level:
- `{{startDate}}`
- `{{endDate}}`
- `{{generatedAt}}`
- `{{repo.name}}`
- `{{repo.description}}`

Summary:
- `{{summary.streamCount}}`
- `{{summary.issueCount}}`
- `{{summary.habitCount}}`
- `{{summary.sessionCount}}`

Streams:
- `{{#each streams}} -> {{name}}, {{description}}`

Issues:
- `{{#each issues}}`
  - `{{id}}`
  - `{{title}}`
  - `{{status}}`
  - `{{scope}}`
  - `{{estimateTime}}`
  - `{{description}}`
  - `{{notes}}`
  - `{{#each sessions}} -> {{summary}}, {{commit}}, {{context}}, {{work}}, {{notes}}`

Habits:
- `{{#each habits}} -> {{name}}, {{scheduleType}}`

Time by tag (empty when no worked issue was tagged; a session counts toward every tag on its issue):
- `{{#each tagTime}} -> {{name}}, {{workedTime}}, {{workedSeconds}}, {{percent}}`
