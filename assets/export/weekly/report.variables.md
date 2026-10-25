# Weekly Report Template Variables

Top-level:
- `{{startDate}}`
- `{{endDate}}`
- `{{generatedAt}}`

Summary:
- `{{summary.days}}`
- `{{summary.checkInDays}}`
- `{{summary.focusDays}}`
- `{{summary.workedTime}}`
- `{{summary.restTime}}`
- `{{summary.completedIssues}}`
- `{{summary.abandonedIssues}}`
- `{{summary.estimatedTime}}`
- `{{summary.averageMood}}`
- `{{summary.averageEnergy}}`

Streaks:
- `{{streaks.currentFocusDays}}`
- `{{streaks.longestFocusDays}}`
- `{{streaks.currentCheckInDays}}`
- `{{streaks.longestCheckInDays}}`

Days:
- `{{#each days}}`
  - `{{date}}`
  - `{{workedTime}}`
  - `{{sessionCount}}`
  - `{{totalIssues}}`
  - `{{completedIssues}}`
  - `{{abandonedIssues}}`
  - `{{checkIn.mood}}`
  - `{{checkIn.energy}}`

Time by tag (empty when no worked issue was tagged; a session counts toward every tag on its issue):
- `{{#each tagTime}} -> {{name}}, {{workedTime}}, {{workedSeconds}}, {{percent}}`
